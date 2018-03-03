package cabOrdersBackup


import (
	"def"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"library/orders"
)

func Dump(o orders.Orders) {
	file, _ := json.Marshal(o)
	err := ioutil.WriteFile("hw_backup", file, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func Get() (orders.Orders, error) {
	file, err := ioutil.ReadFile("hw_backup")
	if err != nil {
		return orders.New(), errors.New("read from file err")
	}
	
	backup := orders.New()
	if err := json.Unmarshal(file, &backup); err != nil {
		return orders.New(), errors.New("json.Unmarshal error")
	}

	cabOrders := orders.New()
	for floor := def.GROUND_FLOOR; floor < def.N_FLOOR; floor++ {
		if orders.ValidateFloorButtonCombination(def.BTN_INTERNAL, floor) {
			cabOrders.Update(backup.Get(def.BTN_INTERNAL,floor))
		}
	}
	return cabOrders, nil
}