package liftCtrl

import (
	"def"
	"fmt"
	"time"
	"library/hw"
	. "library/colors"
	. "library/logger"
)

//----------------------------
//types

type Channels struct {
	EventQueue                   chan Event
	CompleteOrder_to_SynchOrders chan<- def.Order
	Status_to_SynchOrders        chan<- def.Status
	Status_to_LiftWatchdog       chan<- def.Status
}

type stateFunc func(Event, Channels,*def.Status,*def.Order) (next stateFunc)

//----------------------------
//fsm

func Run(ch Channels) {
	stateFunc := stateIDLE

	var liftStatus def.Status
	var currentOrder def.Order

	hw.Init()

	if floor := hw.GetFloor(); floor > -1 {
		liftStatus.LastFloor = floor
	}
	liftStatus.LastDir = def.DIR_DOWN
	liftStatus.Operative = true
	pushStatusToChannels(ch,liftStatus,def.DIR_STOP)

	INFO("liftCtrl/init" + "           |" + ColG + " DONE" + ColN)

	for {
		stateFunc = stateFunc(<-ch.EventQueue, ch, &liftStatus, &currentOrder)
	}
}

func stateIDLE(event Event, ch Channels, liftStatus *def.Status, currentOrder *def.Order) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		if event.boolean == false {
			break
		}

		*currentOrder = def.Order{Floor: event.floor, Button: event.button, Value: event.boolean}
		nextDir := determDir(*liftStatus,*currentOrder)

		if nextDir == def.DIR_UP || nextDir == def.DIR_DOWN {
			hw.SetMotorDir(nextDir)
			liftStatus.LastDir = nextDir
			pushStatusToChannels(ch,*liftStatus,nextDir)
			return stateMOVE
		} else if nextDir == def.DIR_STOP {
			go completeOrder(ch,liftStatus,currentOrder)
			return stateFLOOR
		}
	default:
		WARNING("Unexpected event")
		fmt.Println(event)
	}
	return stateIDLE
}

func stateMOVE(event Event, ch Channels, liftStatus *def.Status, currentOrder *def.Order) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		newOrder := def.Order{Floor: event.floor, Button: event.button, Value: event.boolean}
		if compareSimilarityOfOrders(newOrder, *currentOrder) {
			return stateMOVE
		}

		nextDir := determDir(*liftStatus,newOrder)
		if nextDir == def.DIR_UP || nextDir == def.DIR_DOWN {
			*currentOrder = newOrder
			hw.SetMotorDir(nextDir)
			liftStatus.LastDir = nextDir
			pushStatusToChannels(ch,*liftStatus, liftStatus.LastDir)
			return stateMOVE

		} else if nextDir == def.DIR_STOP && liftStatus.LastFloor == hw.GetFloor(){
			*currentOrder = newOrder
			go completeOrder(ch,liftStatus,currentOrder)
			return stateFLOOR
		}

	case evt_NEW_FLOOR:
		hw.SetFloorLamp(event.floor)
		liftStatus.LastFloor = event.floor
		liftStatus.Operative = true
		pushStatusToChannels(ch,*liftStatus, liftStatus.LastDir)

		if nextDir := determDir(*liftStatus,*currentOrder); nextDir == def.DIR_STOP {
			go completeOrder(ch,liftStatus,currentOrder)
			return stateFLOOR
		}
	case evt_LIFT_OBSTRUCTION:
			liftStatus.Operative = false
			closestFloor := determClosestFloor(*liftStatus)
			*currentOrder = def.Order{Button: def.BTN_INTERNAL, Floor: closestFloor, Value: true} 
	default:
		WARNING("Unexpected event")
		fmt.Println(event)
	}
	return stateMOVE
}

func stateFLOOR(event Event, ch Channels, liftStatus *def.Status, currentOrder *def.Order) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		*currentOrder = def.Order{Floor: event.floor, Button: event.button, Value: event.boolean}
		break

	case evt_DOOR_TIMER_OUT:
		return stateIDLE

	default:
		WARNING("Unexpected event")
		fmt.Println(event)
	}
	return stateFLOOR
}


// //Utilities
func determClosestFloor(liftStatus def.Status) int {
	if liftStatus.LastFloor > def.GROUND_FLOOR && liftStatus.LastFloor < def.TOP_FLOOR {
		return liftStatus.LastFloor + liftStatus.LastDir 
	}
	return liftStatus.LastFloor
}

/* return true if equal floor, button, and value are equal, else returns false */
func compareSimilarityOfOrders(order1, order2 def.Order) bool {
	if order1.Floor == order2.Floor && order1.Button == order2.Button && order1.Value == order2.Value {
		return true
	} 
	return false
}

func pushStatusToChannels(ch Channels, liftStatus def.Status, currentDir int) {
	ch.Status_to_LiftWatchdog <- def.Status{LastFloor: liftStatus.LastFloor, LastDir: currentDir}
	ch.Status_to_SynchOrders <- liftStatus
}

func determDir(status def.Status, order def.Order) int {
	if order.Floor > status.LastFloor {
		return def.DIR_UP

	} else if order.Floor < status.LastFloor {
		return def.DIR_DOWN
	}
	return def.DIR_STOP
}

func completeOrder(ch Channels, liftStatus *def.Status, currentOrder *def.Order) {
	hw.SetDoorLamp(true)
	hw.SetMotorDir(def.DIR_STOP)
	pushStatusToChannels(ch,*liftStatus,def.DIR_STOP)

	currentOrder.Value = false
	currentOrder.Timestamp = time.Now().Unix()

	if currentOrder.Button == def.BTN_UP || currentOrder.Button == def.BTN_DOWN {
		hw.SetButtonLamp(currentOrder.Floor, currentOrder.Button, false)
		ch.CompleteOrder_to_SynchOrders <- *currentOrder
	}
	ch.CompleteOrder_to_SynchOrders <- def.Order{Floor: currentOrder.Floor, Button: def.BTN_INTERNAL, Value: false, Timestamp: time.Now().Unix()}
	hw.SetButtonLamp(currentOrder.Floor, def.BTN_INTERNAL, false)

	timer := time.NewTimer(time.Second * 3)
	<-timer.C

	hw.SetDoorLamp(false)
	ch.EventQueue <- Event{eventType: evt_DOOR_TIMER_OUT}
}

