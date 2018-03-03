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
	var nextFloor int
	moveLimit := time.Second * 3
	infLimit := time.Second * 6666

	timer := time.NewTimer(infLimit)

	INFO("liftWatchdog/init" + "   |" + colors.ColG + " DONE" + colors.ColN)
	for {
		select {
		case status := <-ch.StatusUpdate:
			if status.LastFloor != nextFloor {
				//eventQueue <- liftCtrl.Event{EventType: liftCtrl.NON_FUNCTIONAL}
			}
			switch status.LastDir {
			case def.DIR_UP:
				if status.LastFloor < def.TOP_FLOOR {
					nextFloor = status.LastFloor + 1
					timer = time.NewTimer(moveLimit)
				} else {
					nextFloor = status.LastFloor
					timer = time.NewTimer(infLimit)
				}
			case def.DIR_DOWN:
				if status.LastFloor > def.GROUND_FLOOR {
					nextFloor = status.LastFloor - 1
					timer = time.NewTimer(moveLimit)
				} else {
					nextFloor = status.LastFloor
					timer = time.NewTimer(infLimit)
				}
			case def.DIR_STOP:
				nextFloor = status.LastFloor
				timer = time.NewTimer(infLimit)
			}
		case <-timer.C:
			liftCtrl.Send_HW_FAIL_event(ch.EventQueue)
			timer = time.NewTimer(infLimit)

		}
	}

}

func generateEvent(event liftCtrl.Event, eventQueue chan<- liftCtrl.Event, delay int) {
	time.Sleep(time.Second * time.Duration(delay))
	eventQueue <- event
}