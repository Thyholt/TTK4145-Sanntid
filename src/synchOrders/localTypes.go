package synchOrders

import (
	"def"
	"library/orders"
	"time"
)

const LIFT_ACTIVITY_LIMIT = 500 // ms

//timer
const (
	timer_DURATION_BCAST_HEARTBEAT               = time.Millisecond * 10
	timer_DURATION_BCAST_ORDERS                  = time.Millisecond * 50
	timer_DURATION_PUSH_BESTFITORDER_TO_LIFTCTRL = time.Millisecond * 500
)

type heartbeat struct {
	ID        int
	LiftState def.Status
}

type netChannels struct {
	BcastOrders chan<- orders.Orders
	RecvOrders  <-chan orders.Orders

	BcastHeartbeat chan<- heartbeat
	RecvHeartbeat  <-chan heartbeat
}

type timerChannels struct {
	BcastOrders  <-chan bool
	BcastHeartbeat  <-chan bool
	PushBestFitOrderToLiftCtrl  <-chan bool
}

type intrnChannels struct {
	Net   	netChannels
	Timers 	timerChannels
}
