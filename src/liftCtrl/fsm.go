package liftCtrl

import (
	"def"
	"fmt"
	. "library/colors"
	"library/hw"
	. "library/logger"
	"time"
)

//----------------------------
//types

type Channels struct {
	EventQueue                   chan Event
	CompleteOrder_to_SynchOrders chan<- def.Order
	Status_to_SynchOrders        chan<- def.Status
	Status_to_LiftWatchdog       chan<- def.Status
}

type liftCtrlState struct {
	Status       def.Status
	CurrentOrder def.Order
}

type stateFunc func(Event, Channels, *liftCtrlState) (next stateFunc)

//----------------------------
//fsm

func Run(ch Channels) {
	stateFunc := stateIDLE
	var liftState liftCtrlState

	hw.Init()

	if floor := hw.GetFloor(); floor > -1 {
		liftState.Status.LastFloor = floor
	}
	liftState.Status.LastDir = def.DIR_UP
	liftState.Status.Operative = true

	liftState.pushStatusToChannels(ch)

	INFO("liftCtrl/init" + "           |" + ColG + " DONE" + ColN)

	for {
		stateFunc = stateFunc(<-ch.EventQueue, ch, &liftState)
	}
}

func stateIDLE(event Event, ch Channels, liftState *liftCtrlState) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		fmt.Println(event)
	/*

		if event.floor == def.NONE {
			break
		}
		liftState.CurrentOrder = def.Order{Floor: event.floor, Button: event.button}
		nextDir := determDir(*liftState)

		if nextDir == def.DIR_UP || nextDir == def.DIR_DOWN {
			hw.SetMotorDir(nextDir)
			liftState.Status.LastDir = nextDir
			(*liftState).pushStatusToChannels(ch)
			return stateMOVE
		} else if nextDir == def.DIR_STOP {
			go completeOrder(liftState, ch)
			return stateFLOOR
		}*/
/*
	default:
		if nextState := generalEventHandler(event, ch, liftState); nextState != nil {
			return nextState
		}
	}*/
	}
	return stateIDLE
}

func stateMOVE(event Event, ch Channels, liftState *liftCtrlState) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		break
	case evt_NEW_FLOOR:
		hw.SetFloorLamp(event.floor)
		liftState.Status.LastFloor = event.floor
		(*liftState).pushStatusToChannels(ch)

		if nextDir := determDir(*liftState); nextDir == def.DIR_STOP {
			go completeOrder(liftState, ch)
			return stateFLOOR
		}

	default:
		if nextState := generalEventHandler(event, ch, liftState); nextState != nil {
			return nextState
		}
	}
	return stateMOVE
}

func stateFLOOR(event Event, ch Channels, liftState *liftCtrlState) stateFunc {
	switch event.eventType {
	case evt_EXE_ORDER:
		break

	case evt_DOOR_TIMER_OUT:
		return stateIDLE

	default:
		if nextState := generalEventHandler(event, ch, liftState); nextState != nil {
			return nextState
		}
	}
	return stateFLOOR
}

func generalEventHandler(event Event, ch Channels, liftState *liftCtrlState) stateFunc {
	switch event.eventType {
	/*
		case evt_HW_FAIL:
			go hwFailureProcedure(ch, liftState)
			return stateFLOOR
	*/
	default:
		WARNING("Unexpected event")
		fmt.Println(event)
	}
	return nil
}

// //Utilities
func (liftState liftCtrlState) pushStatusToChannels(ch Channels) {
	ch.Status_to_LiftWatchdog <- liftState.Status
	ch.Status_to_SynchOrders <- liftState.Status
}

func determDir(liftState liftCtrlState) int {
	/*
	if liftState.CurrentOrder == def.DummyOrder {
		return liftState.Status.LastDir
	}*/

	if liftState.CurrentOrder.Floor > liftState.Status.LastFloor {
		return def.DIR_UP

	} else if liftState.CurrentOrder.Floor < liftState.Status.LastFloor {
		return def.DIR_DOWN
	}
	return def.DIR_STOP
}

func completeOrder(liftState *liftCtrlState, ch Channels) {
	hw.SetDoorLamp(true)
	hw.SetMotorDir(def.DIR_STOP)
	if liftState.CurrentOrder.Button == def.BTN_UP || liftState.CurrentOrder.Button == def.BTN_DOWN {
		hw.SetButtonLamp(liftState.CurrentOrder.Floor, liftState.CurrentOrder.Button, false)
		ch.CompleteOrder_to_SynchOrders <- liftState.CurrentOrder
	}
	hw.SetButtonLamp(liftState.CurrentOrder.Floor, def.BTN_INTERNAL, false)
	ch.CompleteOrder_to_SynchOrders <- def.Order{Button: def.BTN_INTERNAL, Floor: liftState.CurrentOrder.Floor}

	timer := time.NewTimer(time.Second * 3)
	<-timer.C

	hw.SetDoorLamp(false)
	ch.EventQueue <- Event{eventType: evt_DOOR_TIMER_OUT}
}
