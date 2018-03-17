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

	
	lastOrdersSensedMatrix := createLastOrdersSensedMatrix()
	lastSensedFloor := -1

	INFO("hwPoll/init" + "               |" + colors.ColG + " DONE" + colors.ColN)

	// To here init

	// Main routine
	for {
		pollOrderPanel(ch.Order_to_SynchOrders, lastOrdersSensedMatrix)
		pollFloorSensor(ch.LiftCtrl_EventQueue, &lastSensedFloor)
	}
}

func pollOrderPanel(order_To_OrderDistr chan<- def.Order, lastOrdersSensedMatrix [][]bool) {
	for button := def.BTN_UP; button < def.N_ORDER_BUTTONS; button++ {
		for floor := def.GROUND_FLOOR; floor <= def.TOP_FLOOR; floor++ {
			status := hw.ReadButton(floor, button)
			if status && status != lastOrdersSensedMatrix[button][floor] {
				order_To_OrderDistr <- def.Order{Floor: floor, Button: button, Value: true,	Timestamp: time.Now()}
			}
			lastOrdersSensedMatrix[button][floor] = status
		}
	}
}

func pollFloorSensor(eventQueue chan<- liftCtrl.Event, lastFloorSensed *int) {
	floor := hw.GetFloor()
	// Add wrapper to if statement
	if floor != -1 && floor != *lastFloorSensed {
		liftCtrl.Send_NEW_FLOOR_event(eventQueue,floor)
	}
	*lastFloorSensed = floor
}

func createLastOrdersSensedMatrix() [][]bool {
	lastOrdersSensedMatrix := [][]bool{}
	up := make([]bool, 4)
	down := make([]bool, 4)
	internal := make([]bool, 4)
	lastOrdersSensedMatrix = append(lastOrdersSensedMatrix, up)
	lastOrdersSensedMatrix = append(lastOrdersSensedMatrix, down)
	lastOrdersSensedMatrix = append(lastOrdersSensedMatrix, internal)
	return lastOrdersSensedMatrix
}

