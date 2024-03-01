package main

import (
	"time"
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Ack struct {
	active bool
	completed bool
	time time.Time
}

type GlobalOrderStruct struct {
	HallOrderArray [2][numFloors]Ack              // Represents the hall orders
	CabOrderArray  [numOfElevators][numFloors]Ack // Represents the cab orders
}

var ackStruct = GlobalOrderStruct{
	HallOrderArray: [2][numFloors]Ack{},
	CabOrderArray:  [numOfElevators][numFloors]Ack{},
}



func OrderCompleted(order Order, e *Elevator) {
	button := order.Button
	floor := order.Floor

	if button == Cab {
		ackStruct.CabOrderArray[e.ID][floor].completed = true
		ackStruct.CabOrderArray[e.ID][floor].time = time.Now()
		ackStruct.CabOrderArray[e.ID][floor].active = false
	} else {
		ackStruct.HallOrderArray[button][floor].completed = true
		ackStruct.HallOrderArray[button][floor].time = time.Now()
		ackStruct.HallOrderArray[button][floor].active = false
	}
}


func OrderActive(order Order, e *Elevator) {
	button := order.Button
	floor := order.Floor

	if button == Cab {
		ackStruct.CabOrderArray[e.ID][floor].active = true
		ackStruct.CabOrderArray[e.ID][floor].completed = false
		ackStruct.CabOrderArray[e.ID][floor].time = time.Now()

	} else {
		ackStruct.HallOrderArray[button][floor].active = true
		ackStruct.HallOrderArray[button][floor].completed = false
		ackStruct.HallOrderArray[button][floor].time = time.Now()
	}
}


func CheckIfOrderIsComplete(e *Elevator, newOrderTx chan MessageNewOrder) {
	currentTime := time.Now()
	var orders []Order

	HallOrderArray := ackStruct.HallOrderArray

	for button := 0; button < 2; button++ {
		for floor := 0; floor < numFloors; floor++ {
			if HallOrderArray[button][floor].active == true && HallOrderArray[button][floor].completed == false {
				if currentTime.Sub(HallOrderArray[button][floor].time) > 10*time.Second {
					orders = append(orders, Order{floor, elevio.ButtonType(button)})
					HallOrderArray[button][floor].active = false
				}
			}
		}
	}


	for i, _ := range orders {

		bestElevator := chooseElevator(Elevators, orders[i])

		newOrder := MessageNewOrder{
			Type:         "MessageNewOrder",
			NewOrder:     orders[i],
			E:            *e,
			ToElevatorID: bestElevator.ID,
		}

		if bestElevator.ID == e.ID {
			e.UpdateOrderSystem(orders[i])
		} else {
			newOrderTx <- newOrder
		}

		OrderActive(orders[i], e)

	}


}

