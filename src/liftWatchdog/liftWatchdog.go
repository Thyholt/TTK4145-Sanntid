package liftWatchdog

import (
	"def"
	"time"
	"liftCtrl"
	"library/colors"
	. "library/logger"
)

type Channels struct {
	StatusUpdate <-chan def.Status
	EventQueue   chan<- liftCtrl.Event
}

func Run(ch Channels) {
	moveLimit := time.Second * 3
	infLimit := time.Second * 6666

	timer := time.NewTimer(infLimit)

	INFO("liftWatchdog/init" + "   |" + colors.ColG + " DONE" + colors.ColN)
	for {
		select {
		case status := <-ch.StatusUpdate:
			if status.LastDir == def.DIR_UP || status.LastDir == def.DIR_DOWN{
				if status.LastFloor < def.TOP_FLOOR || status.LastFloor > def.GROUND_FLOOR{
					timer = time.NewTimer(moveLimit)
				} else {
					timer = time.NewTimer(infLimit)
				}
			} else if status.LastDir == def.DIR_STOP {
				timer = time.NewTimer(infLimit)
			}
		case <-timer.C:
			WARNING("Send_LIFT_OBSTRUCTION_event")
			liftCtrl.Send_LIFT_OBSTRUCTION_event(ch.EventQueue)
			timer = time.NewTimer(infLimit)

		}
	}

}

func generateEvent(event liftCtrl.Event, eventQueue chan<- liftCtrl.Event, delay int) {
	time.Sleep(time.Second * time.Duration(delay))
	eventQueue <- event
}