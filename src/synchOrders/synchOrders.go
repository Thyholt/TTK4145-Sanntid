package synchOrders

import (
	"def"
	"fmt"
	_ "library/assertionCheck"
	. "library/colors"
	. "library/logger"
	"library/hw"
	"library/lifts"
	"library/orders"
	"library/network/bcast"
	"library/cabOrdersBackup"
	"liftCtrl"
	"time"
)

type ExtrnChannels struct {
	EventQueue                 chan<- liftCtrl.Event
	NewOrder_from_hwPoll       <-chan def.Order
	ComleteOrder_from_liftCtrl <-chan def.Order
	Status_from_liftCtrl       <-chan def.Status
}

func Run(ID int, extrnChs ExtrnChannels) {
	
	// Initializiation order synchronization
	var intrnChs intrnChannels
	lifts := lifts.New(LIFT_ACTIVITY_LIMIT)
	localOrders := orders.New()

	initOrders(&localOrders)
	initNetwork(&intrnChs.Net)
	initTimers(&intrnChs.Timers)

	status := <-extrnChs.Status_from_liftCtrl
	lifts.Update(ID, status)

	INFO("synchOrders/init" + "     |" + ColG + " DONE" + ColN)

	for {
		select {
		//liftCtrl
		case status := <-extrnChs.Status_from_liftCtrl:
			lifts.Update(ID, status)
			break

		case completeOrder := <-extrnChs.ComleteOrder_from_liftCtrl:
			if completeOrder.Value {
				break
			}
			localOrders.Update(completeOrder)
			cabOrdersBackup.Dump(localOrders)

		//hwPoll
		case newOrder := <-extrnChs.NewOrder_from_hwPoll:
			localOrders.Update(newOrder)
			cabOrdersBackup.Dump(localOrders)
			break

		// network
		case recvOrders := <-intrnChs.Net.RecvOrders:
			localOrders.Merge(recvOrders)
			break

		case heartbeat := <-intrnChs.Net.RecvHeartbeat:
			lifts.Update(heartbeat.ID, heartbeat.LiftState)
			break

		//timers
		case <-intrnChs.Timers.BcastOrders:
			bcastHallOrders(intrnChs.Net.BcastOrders, localOrders)
			
		case <-intrnChs.Timers.BcastHeartbeat:
			bcastHeartbeat(intrnChs.Net.BcastHeartbeat, ID, lifts.Status(ID))
			break

		case <-intrnChs.Timers.PushBestFitOrderToLiftCtrl:
			lifts.Print()
			localOrders.Print()
			updateOrderbuttonLights(localOrders)
			removeOfflineLifts(lifts)
			liftCtrl.Send_EXE_ORDER_event(extrnChs.EventQueue, determNextOrderAmongOnlineLifts(ID, lifts, localOrders, def.ORDER_COMPLETION_LIMIT))
		}
	}
}

func bcastHeartbeat(ch chan<- heartbeat, ID int, status def.Status) {
	ch <- heartbeat{ID,status}
}

func bcastHallOrders(ch chan<- orders.Orders, localOrders orders.Orders) {
	hallOrders := orders.New()
	for floor := def.GROUND_FLOOR; floor < def.N_FLOOR; floor++ {
		for button := range []int{def.BTN_UP, def.BTN_DOWN,} {
			if orders.ValidateFloorButtonCombination(button, floor) {
				hallOrders.Update(localOrders.Get(button,floor))
			}
		}
	}
	ch <- hallOrders
}

func removeOfflineLifts(lifts lifts.Lifts) {
	for _, id := range lifts.IDs() {
		if lifts.NetState(id) == def.OFFLINE {
			lifts.Delete(id)
		}
	}
}

func updateOrderbuttonLights(o orders.Orders) {
	for floor := def.GROUND_FLOOR; floor < def.N_FLOOR; floor++ {
		for button := range []int{def.BTN_UP, def.BTN_DOWN, def.BTN_INTERNAL} {
			if orders.ValidateFloorButtonCombination(button, floor) {
				hw.SetButtonLamp(floor, button, o.Get(button, floor).Value)
				hw.SetButtonLamp(floor, def.BTN_INTERNAL, o.Get(def.BTN_INTERNAL, floor).Value)
			}
		}
	}
}

// orderCompletionLimit [ms]
func determNextOrderAmongOnlineLifts(liftID int, lifts lifts.Lifts, orders orders.Orders, orderCompletionLimit time.Duration) def.Order {
	if !lifts.Status(liftID).Operative {
		return def.Order{Value: false}
	}

	// see if any order has exceeded the completionLimit, and if so choose the closest one
	temp_order, temp_dist := determClosestOrderAndDist(orders, lifts.Status(liftID),orderCompletionLimit)
	if temp_order.Value {
		WARNING("orderCompletionLimit passed")
		return temp_order
	}

	// see if closest order is an cab order
	temp_order, temp_dist = determClosestOrderAndDist(orders, lifts.Status(liftID), def.NULL * time.Second)
	if temp_order.Value && temp_order.Button == def.BTN_INTERNAL {
		return temp_order
	}

	// see if self lift is closest to an hall order
	temp_dist = def.INF
	order, dist := determClosestOrderAndDist(orders, lifts.Status(liftID),def.NULL * time.Second)
	if !order.Value {
		return def.Order{Value: false}
	}

	for _ , id := range lifts.IDs() {
		if id == liftID || (lifts.NetState(id) == def.OFFLINE || lifts.Status(id).Operative == false) {
			continue
		}

		temp_order, temp_dist = determClosestOrderAndDist(orders, lifts.Status(id),def.NULL * time.Second)
		if compareOrderSimilarity(order,temp_order) {
			if temp_dist < dist || (temp_dist == dist && id < liftID) {
				return def.Order{Value: false}
			}
		}
	}

	return order
}

// info:   minElapsedTime [ms]
func orderExceedesMinElapsedTime(order def.Order, minElapsedTime time.Duration) bool {
	return time.Since(order.Timestamp) >= minElapsedTime * time.Millisecond
}

// neglects timestamp
func compareOrderSimilarity(order1, order2 def.Order) bool {
	return order1.Floor == order2.Floor && order1.Button == order2.Button && order1.Value == order2.Value 
}

// info:   minElapsedTime [ms]
func determClosestOrderAndDist(orders orders.Orders, liftStatus def.Status, minElapsedTime time.Duration) (def.Order,int) {
	var order def.Order
	var distance int
	if liftStatus.LastFloor != def.GROUND_FLOOR && orders.Get(def.BTN_DOWN, liftStatus.LastFloor).Value {
		return orders.Get(def.BTN_DOWN, liftStatus.LastFloor), def.NULL
	} else if liftStatus.LastFloor != def.TOP_FLOOR && orders.Get(def.BTN_UP, liftStatus.LastFloor).Value {
		return orders.Get(def.BTN_UP, liftStatus.LastFloor), def.NULL
	}

	if liftStatus.LastDir == def.DIR_UP {
		if order, distance = searchUp(orders, liftStatus.LastFloor, def.TOP_FLOOR, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		} else if order, distance = searchDown(orders, def.TOP_FLOOR, def.GROUND_FLOOR, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		} else if order, distance = searchUp(orders, def.GROUND_FLOOR, liftStatus.LastFloor, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		}
	} else if liftStatus.LastDir == def.DIR_DOWN {
		if order, distance = searchDown(orders, liftStatus.LastFloor, def.GROUND_FLOOR, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		} else if order, distance = searchUp(orders, def.GROUND_FLOOR, def.TOP_FLOOR, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		} else if order, distance = searchDown(orders, def.TOP_FLOOR, liftStatus.LastFloor, distance); order.Value {
			if orderExceedesMinElapsedTime(order,minElapsedTime) {return order, distance}
		}
	}
	return def.Order{Value: false}, def.INF
}

// info:   looks for order in which order.Value is true
// return: order, distance
func searchDown(orders orders.Orders, top int, buttom int, dist int) (def.Order, int) {
	for floor := top; floor > buttom; floor-- {
		if order := orders.Get(def.BTN_DOWN, floor); order.Value {
			return order, dist
		} else if order := orders.Get(def.BTN_INTERNAL, floor); order.Value {
			return order, dist
		}
		dist += 1
	}
	return def.Order{Value: false}, dist
}

// info:   looks for order in which order.Value is true
// return: order, distance
func searchUp(orders orders.Orders, buttom int, top int, dist int) (def.Order, int) {
	for floor := buttom; floor < top; floor++ {
		if order := orders.Get(def.BTN_UP, floor); order.Value {
			return order, dist
		} else if order := orders.Get(def.BTN_INTERNAL, floor); order.Value {
			return order, dist
		}
		dist += 1
	}
	return def.Order{Value: false}, dist
}

func initOrders(localOrders *orders.Orders) {
	if cabOrders, err := cabOrdersBackup.Get(); err == nil {
		localOrders.Merge(cabOrders)
	} else {
		fmt.Println("cabOrdersBackup failed:", err)
	}
}

func initTimers(timerChs *timerChannels) {
	timer_bcastOrders_ch := make(chan bool)
	timer_BcastHeartbeat_ch := make(chan bool)
	timer_PushBestFitOrderToLiftCtrl_ch := make(chan bool)

	timerChs.BcastOrders = timer_bcastOrders_ch
	timerChs.BcastHeartbeat = timer_BcastHeartbeat_ch
	timerChs.PushBestFitOrderToLiftCtrl = timer_PushBestFitOrderToLiftCtrl_ch

	go func() {
		ticker_bcast_heartbeat := time.NewTicker(timer_DURATION_BCAST_HEARTBEAT)
		ticker_bcast_orders := time.NewTicker(timer_DURATION_BCAST_ORDERS)
		ticker_push_bestfit := time.NewTicker(timer_DURATION_PUSH_BESTFITORDER_TO_LIFTCTRL)
		for {
			select {
			case <- ticker_bcast_heartbeat.C:
				timer_BcastHeartbeat_ch <- true
			case <- ticker_bcast_orders.C:
				timer_bcastOrders_ch <- true
			case <- ticker_push_bestfit.C:
				timer_PushBestFitOrderToLiftCtrl_ch <- true
			}
		}
	}()
}

func initNetwork(netChs *netChannels) {
	bcastOrders_ch := make(chan orders.Orders)
	recvOrders_ch := make(chan orders.Orders)

	bcastHeartbeat_ch := make(chan heartbeat)
	recvHeartbeat_ch := make(chan heartbeat)

	go bcast.Transmitter(def.PORT, bcastOrders_ch, bcastHeartbeat_ch)
	go bcast.Receiver(def.PORT, recvOrders_ch, recvHeartbeat_ch)

	netChs.BcastOrders = bcastOrders_ch
	netChs.RecvOrders = recvOrders_ch

	netChs.BcastHeartbeat = bcastHeartbeat_ch
	netChs.RecvHeartbeat = recvHeartbeat_ch

}
