// Package hw configines interactions with the lift hardware at the real time
// lab at The Department of Engineering Cybernetics at NTNU, Trondheim,
// Norway.
//
// This file is a golang port of elev.c from the hand out driver
// (https://github.com/TTK4145/Project)
package hw

import (
	"def"
	"library/colors"
	"errors"
	"fmt"
	"log"
)

var lampChannelMatrix = [def.N_FLOOR][def.N_ORDER_BUTTONS]int{
	{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
}
var buttonChannelMatrix = [def.N_FLOOR][def.N_ORDER_BUTTONS]int{
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
	{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
	{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

// Init initialises the lift hardware and moves the lift to a defined state.
// (Descending until it reaches a floor.)
func Init() error {
	// Init hardware
	if !ioInit() {
		return errors.New(colors.ColR + "Hardware driver: ioInit() failed!" + colors.ColR)
	}

	// Zero all floor button lamps
	for f := def.GROUND_FLOOR; f <= def.TOP_FLOOR; f++ {
		if f != def.GROUND_FLOOR {
			SetButtonLamp(f, def.BTN_DOWN, false)
		}
		if f != def.TOP_FLOOR {
			SetButtonLamp(f, def.BTN_UP, false)
		}
		SetButtonLamp(f, def.BTN_INTERNAL, false)
	}

	// Move to defined state
	SetMotorDir(def.DIR_DOWN)
	floor := GetFloor()
	for floor == -1 {
		floor = GetFloor()
	}
	SetMotorDir(def.DIR_STOP)
	SetFloorLamp(floor)

	SetStopLamp(false)
	SetDoorLamp(false)

	fmt.Println(colors.ColM + "Hardware initialised." + colors.ColN)
	return nil
}

func SetMotorDir(dirn int) {
	if dirn == 0 {
		ioWriteAnalog(MOTOR, 0)
	} else if dirn > 0 {
		ioClearBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	} else if dirn < 0 {
		ioSetBit(MOTORDIR)
		ioWriteAnalog(MOTOR, 2800)
	}
}

func SetDoorLamp(value bool) {
	if value {
		ioSetBit(LIGHT_DOOR_OPEN)
	} else {
		ioClearBit(LIGHT_DOOR_OPEN)
	}
}

func GetFloor() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return 0
	} else if ioReadBit(SENSOR_FLOOR2) {
		return 1
	} else if ioReadBit(SENSOR_FLOOR3) {
		return 2
	} else if ioReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func SetFloorLamp(floor int) {
	if floor < 0 || floor >= def.N_FLOOR {
		log.Printf("Error: Floor %d out of range!\n", floor)
		log.Println("No floor indicator will be set.")
		return
	}

	// Binary encoding. One light must always be on.
	if floor&0x02 > 0 {
		ioSetBit(LIGHT_FLOOR_IND1)
	} else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}

	if floor&0x01 > 0 {
		ioSetBit(LIGHT_FLOOR_IND2)
	} else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
}

func ReadButton(floor int, button int) bool {
	if floor < 0 || floor >= def.N_FLOOR {
		//log.Printf("Error: Floor %d out of range!\n", floor)
		return false
	}
	if button < 0 || button >= def.N_ORDER_BUTTONS {
		//log.Printf("Error: Button %d out of range!\n", button)
		return false
	}
	if button == def.BTN_UP && floor == def.N_FLOOR-1 {
		//log.Println("Button up from top floor does not exist!")
		return false
	}
	if button == def.BTN_DOWN && floor == 0 {
		//log.Println("Button down from ground floor does not exist!")
		return false
	}

	if ioReadBit(buttonChannelMatrix[floor][button]) {
		return true
	} else {
		return false
	}
}

func SetButtonLamp(floor int, button int, value bool) {
	if floor < 0 || floor >= def.N_FLOOR {
		log.Printf("Error: Floor %d out of range!\n", floor)
		return
	}
	if button == def.BTN_UP && floor == def.N_FLOOR-1 {
		log.Println("Button up from top floor does not exist!")
		return
	}
	if button == def.BTN_DOWN && floor == 0 {
		log.Println("Button down from ground floor does not exist!")
		return
	}
	if button != def.BTN_UP &&
		button != def.BTN_DOWN &&
		button != def.BTN_INTERNAL {
		log.Printf("Invalid button %d\n", button)
		return
	}

	if value {
		ioSetBit(lampChannelMatrix[floor][button])
	} else {
		ioClearBit(lampChannelMatrix[floor][button])
	}
}

func SetStopLamp(value bool) {
	if value {
		ioSetBit(LIGHT_STOP)
	} else {
		ioClearBit(LIGHT_STOP)
	}
}

func GetStopSignal() bool {
	return ioReadBit(STOP)
}
