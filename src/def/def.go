package def

import "time"

const INF = 99999999
const NULL = 0
const NONE = -99

//==================== Network config ========================|
const PORT = 16666

//==================== Operation limits ======================|

const ORDER_COMPLETION_LIMIT = 20000 // ms
const DOOR_OPEN_DURATION = 2500		 // ms


//button
const N_ORDER_BUTTONS = 3

const (
	BTN_UP = int(iota)
	BTN_DOWN
	BTN_INTERNAL
)

//direction
const (
	DIR_DOWN = int(iota - 1)
	DIR_STOP
	DIR_UP
)

//floor
const (
	N_FLOOR      = 4 //MODIFY
	GROUND_FLOOR = 0 //MODIFY
	TOP_FLOOR    = N_FLOOR - 1
)

type NetState bool
const (
	OFFLINE = NetState(false)
	ONLINE = NetState(true)
)

//===================== Standard types ======================|

type Order struct {
	Floor  		int
	Button 		int
	Value 		bool
	Timestamp 	time.Time //unix
}

type Status struct {
	LastFloor int
	LastDir   int
	Operative bool
}

type Lift struct {
	ID               int
	LastStatusUpdate time.Time
	Status           Status
}
