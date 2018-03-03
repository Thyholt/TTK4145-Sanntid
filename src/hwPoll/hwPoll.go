package hwPoll

import (
	"def"
	"time"
	"liftCtrl"
	"library/hw"
	"library/colors"
	. "library/logger"
)

type Channels struct{
	LiftCtrl_EventQueue chan<- liftCtrl.Event
	Order_to_SynchOrders chan<- def.Order
}

func Run(ch Channels) {
	lastOrdersSensed := [][]bool{}
	up := make([]bool, 4)
	down := make([]bool, 4)
	internal := make([]bool, 4)
	lastOrdersSensed = append(lastOrdersSensed, up)
	lastOrdersSensed = append(lastOrdersSensed, down)
	lastOrdersSensed = append(lastOrdersSensed, internal)

	lastSensedFloor := -1

	INFO("hwPoll/init" + "               |" + colors.ColG + " DONE" + colors.ColN)

	for {
		pollOrderPanel(ch.Order_to_SynchOrders, lastOrdersSensed)
		pollFloorSensor(ch.LiftCtrl_EventQueue, &lastSensedFloor)
	}
}

func pollOrderPanel(order_To_OrderDistr chan<- def.Order, lastOrdersSensed [][]bool) {
	for button := def.BTN_UP; button < def.N_ORDER_BUTTONS; button++ {
		for floor := def.GROUND_FLOOR; floor <= def.TOP_FLOOR; floor++ {
			status := hw.ReadButton(floor, button)
			if status && status != lastOrdersSensed[button][floor] {
				lastOrdersSensed[button][floor] = true
				order_To_OrderDistr <- def.Order{Floor:     floor,
												Button:    	button,
												Value: 		true,
												Timestamp: 	time.Now().Unix()}
			}
			lastOrdersSensed[button][floor] = status
		}
	}
}

func pollFloorSensor(eventQueue chan<- liftCtrl.Event, lastFloorSensed *int) {
	//fmt.Println("pollFloorSensor\n")
	floor := hw.GetFloor()
	if floor != -1 && floor != *lastFloorSensed {
		liftCtrl.Send_NEW_FLOOR_event(eventQueue,floor)
	}
	*lastFloorSensed = floor
}

