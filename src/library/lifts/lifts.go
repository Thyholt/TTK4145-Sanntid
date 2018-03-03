package lifts

import (
	"def"
	"fmt"
	"sort"
	_ "library/logger"
	. "library/assertionCheck"
	"time"
	"strconv"
	. "library/colors"
)

type Lifts struct {
	list map[int]*def.Lift
	onlineActivityLimit time.Duration
}


func New(onlineActivityLimit time.Duration) Lifts {
	var lifts = Lifts{list: make(map[int]*def.Lift), onlineActivityLimit: onlineActivityLimit}
	return lifts
}

func (lifts Lifts) Exist(ID int) bool {
	if _, ok := lifts.list[ID]; ok {
		return true
	}
	return false
}

func (lifts Lifts) Update(ID int, status def.Status) {
	if !lifts.Exist(ID) {
		lifts.list[ID] = &def.Lift{}
	}
	lifts.list[ID].ID = ID
	lifts.list[ID].Status = status
	lifts.list[ID].LastStatusUpdate = time.Now()
}

func (lifts Lifts) Delete(ID int){
	delete(lifts.list,ID)
}

func (lifts Lifts) Status(ID int) def.Status {
	ASSERTION_CHECK(!lifts.Exist(ID),"ID " + strconv.Itoa(ID) + " does not exist in lifts")
	return lifts.list[ID].Status
}

func (lifts Lifts) NetState(ID int) def.NetState {
	ASSERTION_CHECK(!lifts.Exist(ID),"ID " + strconv.Itoa(ID) + " does not exist in lifts")
	if time.Since(lifts.list[ID].LastStatusUpdate) < lifts.onlineActivityLimit * time.Millisecond {
		return def.ONLINE
	}
	return def.OFFLINE
}



func (lifts Lifts) IDs() []int {
	var IDs []int
	for id := range lifts.list {
		IDs = append(IDs, id)
	}
	return IDs
}

func (lifts Lifts) OnlineLiftIDs() []int {
	var onlineLiftsIDs []int
	for _, ID := range lifts.IDs() {
		if lifts.NetState(ID) {
			onlineLiftsIDs = append(onlineLiftsIDs, ID)
		}
	}
	return onlineLiftsIDs
}

func (lifts Lifts) Print() {
	fmt.Print("\n|-----------------------------" + ColB + "LIFTS" + ColN + "-------------------------------|\n")
	fmt.Print("|" + ColB + " ID	LAST_FLOOR LAST_DIR  OPERATIV ONLINE  DUR SINCE HEARTBEAT" + ColN + "|\n")

	var operative string
	var active string
	var dir string

	IDs := lifts.IDs()
	sort.Ints(IDs)

	for _,id := range IDs {
		if lifts.NetState(id) {
			active = ColG + "true" + ColN
		} else {
			active = ColR + "false" + ColN
		}
		if lifts.list[id].Status.Operative && lifts.NetState(id) == def.ONLINE {
			operative = ColG + "true" + ColN
		} else {
			operative = ColR + "false" + ColN
		}
		if lifts.list[id].Status.LastDir == def.DIR_UP && lifts.NetState(id) {
			dir = ColG + "↑" + ColN
		} else {
			dir = ColG + "↓" + ColN
		}

		fmt.Printf("| %-10s %-8d %-20s %-17s %-15s      "+ColY+" %s"+ColN+"\n", strconv.Itoa(id), lifts.list[id].Status.LastFloor, dir, operative, active, time.Since(lifts.list[id].LastStatusUpdate))
	}
}



