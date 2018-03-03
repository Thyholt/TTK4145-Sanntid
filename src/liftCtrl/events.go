package liftCtrl

import (
	"def"
)

type Event struct {
	eventType int
	button    int
	floor     int
	boolean   bool
}

func Send_EXE_ORDER_event(eventQueue chan<- Event, order def.Order) {
	eventQueue <- Event{eventType: evt_EXE_ORDER, floor: order.Floor, button: order.Button}

}

func Send_NEW_FLOOR_event(eventQueue chan<- Event, floor int) {
	eventQueue <- Event{
		eventType: evt_NEW_FLOOR,
		floor:     floor,
		button:    -1,
		boolean:   true}
}

func Send_HW_FAIL_event(eventQueue chan<- Event) {
	eventQueue <- Event{eventType: evt_HW_FAIL, boolean: true}
}

//events
const (
	evt_HW_FAIL = int(iota)
	evt_EXE_ORDER
	evt_NEW_FLOOR

	evt_DOOR_TIMER_OUT
	evt_RESTART_DONE
)
