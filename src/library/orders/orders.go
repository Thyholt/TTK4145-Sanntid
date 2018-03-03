package orders

import (
	"def"
	"fmt"
	. "library/assertionCheck"
	. "library/colors"
	_ "library/logger"
	"strconv"
)

// index: button, floor
type Orders [][]def.Order

func New() Orders {
	var orders = Orders{}
	up := make([]def.Order, def.N_FLOOR)
	down := make([]def.Order, def.N_FLOOR)
	internal := make([]def.Order, def.N_FLOOR)
	orders = append(orders, up)
	orders = append(orders, down)
	orders = append(orders, internal)
	return orders
}

// Saves the newest order
func (orders Orders) Update(order def.Order) {
	ASSERTION_ERROR(!ValidateFloorButtonCombination(order.Button, order.Floor), "Invalid function arguments. Button: "+strconv.Itoa(order.Button)+" Floor: "+strconv.Itoa(order.Floor))
	if order.Timestamp > orders[order.Button][order.Floor].Timestamp {
		orders[order.Button][order.Floor].Value = order.Value
		orders[order.Button][order.Floor].Timestamp = order.Timestamp
	}
}

func (orders Orders) Get(button, floor int) def.Order {
	ASSERTION_ERROR(!ValidateFloorButtonCombination(button, floor), "Invalid function arguments. Button: "+strconv.Itoa(button)+" Floor: "+strconv.Itoa(floor))
	orders[button][floor].Button = button
	orders[button][floor].Floor = floor
	return orders[button][floor]
}

func (orders Orders) Merge(ordersToMerge Orders) {
	for floor := def.GROUND_FLOOR; floor < def.N_FLOOR; floor++ {
		for button := range []int{def.BTN_UP, def.BTN_DOWN, def.BTN_INTERNAL} {
			if ValidateFloorButtonCombination(button, floor) {
				if ordersToMerge[button][floor].Timestamp > orders[button][floor].Timestamp {
					orders[button][floor].Value = ordersToMerge[button][floor].Value
					orders[button][floor].Timestamp = ordersToMerge[button][floor].Timestamp
				}
			}
		}
	}
}

func ValidateFloorButtonCombination(button, floor int) bool {
	if button == def.BTN_UP && floor < def.TOP_FLOOR && floor >= def.GROUND_FLOOR {
		return true
	} else if button == def.BTN_DOWN && floor <= def.TOP_FLOOR && floor > def.GROUND_FLOOR {
		return true
	} else if button == def.BTN_INTERNAL && floor <= def.TOP_FLOOR && floor >= def.GROUND_FLOOR {
		return true
	}
	return false
}

func (orders Orders) Print() {
	fmt.Printf("|-----------------------------------------------------------------|\n")
	fmt.Printf("" + ColB + "         GROUND_FLOOR      FIRST        SECOND       TOP_FLOOR" + ColN + "\n")

	fmt.Printf("| UP:	    ")
	for i := 0; i < 3; i++ {
		if orders.Get(def.BTN_UP, i).Value {
			fmt.Printf("(%s)        ", ColG+"true" + ColN)
		} else {
			fmt.Printf("(%s)       ", ColR+"false" + ColN)
		}
	}
	fmt.Print("        ")

	fmt.Printf("\n| DOWN:                   ")
	for i := 1; i < 4; i++ {
		if orders.Get(def.BTN_DOWN, i).Value {
			fmt.Printf("(%s)        ", ColG+"true" + ColN)
		} else { 
			fmt.Printf("(%s)       ", ColR+"false" + ColN)
		}
	}

	fmt.Printf("\n| CAB:      ")
	for i := 0; i < 4; i++ {
		if orders.Get(def.BTN_INTERNAL, i).Value {
			fmt.Printf("(%s)        ", ColG+"true"+ColN)
		} else {
			fmt.Printf("(%s)       ", ColR+"false"+ColN)
		}
	}
	fmt.Printf("\n|-----------------------------------------------------------------|\n")
}
