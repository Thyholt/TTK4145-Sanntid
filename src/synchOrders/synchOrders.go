package synchOrders

import (
	"def"
	//"fmt"
	_ "library/assertionCheck"
	. "library/colors"
	. "library/logger"
	"library/hw"
	"library/lifts"
	"library/network/bcast"
	"library/orders"
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
	orders := orders.New()
	orders.Update(def.Order{def.BTN_UP, def.GROUND_FLOOR+2, true, time.Now().Unix()})

	initNetwork(&intrnChs.Net)
	initTimers(&intrnChs.Timers)

	INFO("synchOrders/init" + "     |" + ColG + " DONE" + ColN)

	for {
		select {
		//liftCtrl
		case status := <-extrnChs.Status_from_liftCtrl:
			removeOfflineLifts(lifts)
			lifts.Update(ID, status)
			break

		case completeOrder := <-extrnChs.ComleteOrder_from_liftCtrl:
			orders.Update(completeOrder)

		//hwPoll
		case newOrder := <-extrnChs.NewOrder_from_hwPoll:
			orders.Update(newOrder)
			break

		// network
		case recvOrders := <-intrnChs.Net.RecvOrders:
			orders.Merge(recvOrders)
			break
		case heartbeat := <-intrnChs.Net.RecvHeartbeat:
			lifts.Update(heartbeat.ID, heartbeat.LiftState)
			break

		//timers
		case <-intrnChs.Timers.BcastOrders:
			intrnChs.Net.BcastOrders <- orders

		case <-intrnChs.Timers.BcastHeartbeat:
			intrnChs.Net.BcastHeartbeat <- heartbeat{ID,lifts.Status(ID)}
			break

		case <-intrnChs.Timers.PushBestFitOrderToLiftCtrl:
			lifts.Print()
			orders.Print()

			if bestFitOrder := determNextOrder(ID, lifts, orders); bestFitOrder.Value {
				liftCtrl.Send_EXE_ORDER_event(extrnChs.EventQueue, bestFitOrder)
			}
			break

		}
	}
}


func removeOfflineLifts(lifts lifts.Lifts) {
	for _, id := range lifts.IDs() {
		if lifts.NetState(id) == def.OFFLINE {
			lifts.Delete(id)
		}
	}
}

func updateOrderbuttonLights(orders orders.Orders) {
	for floor := def.GROUND_FLOOR; floor < def.N_FLOOR; floor++ {
		for button := range []int{def.BTN_UP, def.BTN_DOWN, def.BTN_INTERNAL} {
			if validateFloorButtonCombination(button, floor) {
				hw.SetButtonLamp(floor, button, orders.Get(button, floor).Value)
				hw.SetButtonLamp(floor, def.BTN_INTERNAL, orders.Get(def.BTN_INTERNAL, floor).Value)
			}
		}
	}
}

func validateFloorButtonCombination(button, floor int) bool {
	if button != def.BTN_UP && button != def.BTN_INTERNAL && button != def.BTN_DOWN {
		return false
	}
	if floor < def.GROUND_FLOOR || floor > def.TOP_FLOOR {
		return false
	}
	if (button == def.BTN_DOWN && floor == def.GROUND_FLOOR) || (button == def.BTN_UP && floor == def.TOP_FLOOR) {
		return false
	}
	return true
}


func determNextOrder(liftID int, lifts lifts.Lifts, orders orders.Orders) def.Order {
	nextOrder, distance := determClosestOrderAndDist(orders, lifts.Status(liftID))
	for _ , id := range lifts.IDs() {
		if id == liftID || lifts.NetState(id) == def.OFFLINE || lifts.Status(id).Operative == false {
			break 
		}
		if _, testDistance := determClosestOrderAndDist(orders, lifts.Status(id)); testDistance < distance {
			return def.Order{Value: false}
		}
		INFO("HERE?")
	}
	return nextOrder
}


// Determs next order and distance
// If successful, next order has value = true, else value = false
func determClosestOrderAndDist(orders orders.Orders, liftStatus def.Status) (def.Order,int) {
	var floor, button, distance int
	var done bool
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