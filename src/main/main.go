package main

import (
	"def"
	"liftCtrl"
	"liftWatchdog"
	"synchOrders"
	"hwPoll"
	"library/network/localip"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)
// Illustration of dependencies 
// 			main 
//			/ \
// 		   /   \
func main() {
	//init flow between goroutines
	ch_eventQueue := make(chan liftCtrl.Event)
	ch_Order_hwPoll_to_synchOrders := make(chan def.Order)
	ch_CompleteOrder_liftCtrl_to_synchOrders := make(chan def.Order)
	ch_Status_liftCtrl_to_synchOrders := make(chan def.Status)
	ch_Status_liftCtrl_to_liftWatchdog := make(chan def.Status)

	chs_synchOrders := synchOrders.ExtrnChannels{
		EventQueue:           		ch_eventQueue,
		NewOrder_from_hwPoll:    		ch_Order_hwPoll_to_synchOrders,
		ComleteOrder_from_liftCtrl: ch_CompleteOrder_liftCtrl_to_synchOrders,
		Status_from_liftCtrl: 		ch_Status_liftCtrl_to_synchOrders}

	chs_liftCtrl := liftCtrl.Channels{
		EventQueue:           		  ch_eventQueue,
		CompleteOrder_to_SynchOrders: ch_CompleteOrder_liftCtrl_to_synchOrders,
		Status_to_SynchOrders: 		  ch_Status_liftCtrl_to_synchOrders,
		Status_to_LiftWatchdog:    	  ch_Status_liftCtrl_to_liftWatchdog}

	chs_hwPoll := hwPoll.Channels{
		LiftCtrl_EventQueue: ch_eventQueue,
		Order_to_SynchOrders: ch_Order_hwPoll_to_synchOrders}

	chs_liftWatchdog := liftWatchdog.Channels{
		EventQueue:           ch_eventQueue,
		StatusUpdate: ch_Status_liftCtrl_to_liftWatchdog}

	// Wrap to get ID function
	IP, err := localip.Get()
	if err != nil {  //temporary bug for different IPs
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		IP = "255.255.255." + strconv.Itoa(r1.Intn(100)+1)
	}

	ID_temp, _ := strconv.ParseInt(strings.Split(IP, ".")[3], 10, 0)
	ID := int(ID_temp)

	//run goroutines
	go hwPoll.Run(chs_hwPoll)
	go synchOrders.Run(ID, chs_synchOrders)
	go liftCtrl.Run(chs_liftCtrl)
	go liftWatchdog.Run(chs_liftWatchdog)

	//inf sleep
	fmt.Print("Main goes to sleep\n")
	time.Sleep(time.Second * 1000000)
}
