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
			cabOrdersBackup.Dump(localOrders)
			localOrders.Update(completeOrder)

		//hwPoll
		case newOrder := <-extrnChs.NewOrder_from_hwPoll:
			cabOrdersBackup.Dump(localOrders)
			localOrders.Update(newOrder)
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
			bcastHeartbeat(intrnChs.Net.BcastHeartbeat,ID,lifts.Status(ID))
			break

		case <-intrnChs.Timers.PushBestFitOrderToLiftCtrl:
			lifts.Print()
			localOrders.Print()
			updateOrderbuttonLights(localOrders)
			removeOfflineLifts(lifts)
			if bestFitOrder := determNextOrderAmongOnlineLifts(ID, lifts, localOrders); bestFitOrder.Value {
				liftCtrl.Send_EXE_ORDER_event(extrnChs.EventQueue, bestFitOrder)
			}
			break

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

func determNextOrderAmongOnlineLifts(liftID int, lifts lifts.Lifts, orders orders.Orders) def.Order {
	// see if closest order is an cab order
	temp_order, temp_dist := determClosestOrderAndDist(orders, lifts.Status(liftID))
	if temp_order.Value && temp_order.Button == def.BTN_INTERNAL && lifts.Status(liftID).Operative == true {
		return temp_order
	}

	// see if self is closest to an hall order
	temp_dist = def.INF

	order, dist := determClosestOrderAndDist(orders, lifts.Status(liftID))
	for _ , id := range lifts.IDs() {
		if lifts.NetState(id) == def.OFFLINE || lifts.Status(id).Operative == false {
			break 
		}
		if temp_order, temp_dist = determClosestOrderAndDist(orders, lifts.Status(id)); temp_order.Value && temp_dist < dist {
			return def.Order{Value: false}
		}
	}
	if lifts.Status(liftID).Operative {
		return order
	}
	return def.Order{Value: false}
}


// Determs next order and distance
// If successful, next order has value = true, else value = false
func determClosestOrderAndDist(orders orders.Orders, liftStatus def.Status) (def.Order,int) {
	var floor, button, distance int
	var done bool
	if liftStatus.LastFloor != def.GROUND_FLOOR && orders.Get(def.BTN_DOWN, liftStatus.LastFloor).Value {
		return orders.Get(def.BTN_DOWN, liftStatus.LastFloor), def.NULL
	} else if liftStatus.LastFloor != def.TOP_FLOOR && orders.Get(def.BTN_UP, liftStatus.LastFloor).Value {
		return orders.Get(def.BTN_UP, liftStatus.LastFloor), def.NULL
	}

	if liftStatus.LastDir == def.DIR_UP {
		if floor, button, distance, done = searchUp(orders, liftStatus.LastFloor, def.TOP_FLOOR, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		} else if floor, button, distance, done = searchDown(orders, def.TOP_FLOOR, def.GROUND_FLOOR, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		} else if floor, button, distance, done = searchUp(orders, def.GROUND_FLOOR, liftStatus.LastFloor, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		}
	} else if liftStatus.LastDir == def.DIR_DOWN {
		if floor, button, distance, done = searchDown(orders, liftStatus.LastFloor, def.GROUND_FLOOR, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		} else if floor, button, distance, done = searchUp(orders, def.GROUND_FLOOR, def.TOP_FLOOR, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		} else if floor, button, distance, done = searchDown(orders, def.TOP_FLOOR, liftStatus.LastFloor, distance); done {
			return def.Order{Button: button, Floor: floor, Value: true}, distance
		}
	}
	return def.Order{Value: false}, def.INF
}

//floor, button, distance, and done
func searchDown(orders orders.Orders, top int, buttom int, dist int) (int, int, int, bool) {
	for floor := top; floor > buttom; floor-- {
		if done := orders.Get(def.BTN_DOWN, floor).Value; done {
			return floor, def.BTN_DOWN, dist, true
		} else if done := orders.Get(def.BTN_INTERNAL, floor).Value; done {
			return floor, def.BTN_INTERNAL, dist, true
		}
		dist += 1
	}
	return def.NONE, def.NONE, dist, false
}

//floor, button, distance, and done
func searchUp(orders orders.Orders, buttom int, top int, dist int) (int, int, int, bool) {
	for floor := buttom; floor < top; floor++ {
		if done := orders.Get(def.BTN_UP, floor).Value; done {
			return floor, def.BTN_UP, dist, true
		} else if done := orders.Get(def.BTN_INTERNAL, floor).Value; done {
			return floor, def.BTN_INTERNAL, dist, true
		}
		dist += 1
	}
	return def.NONE, def.NONE, dist, false
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
