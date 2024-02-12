package main

import (
	"Heislab-Sanntid/elevio"

	"github.com/runarto/Heislab-Sanntid/elevio"

	"math"
)

type Elevator struct {
	CurrentDirection elevio.MotorDirection // Elevator direction
	CurrentFloor     int                   // Last or current floor.
	doorOpen         bool                  // Door open/closed
	Obstruction      bool                  // Obstruction or not
	stopButton       bool                  // Stop button pressed or not
	ActiveOrders     []activeOrder         //List of structs
	NetworkAdress    string
	// Should probably contain it's network address in a string format.

}

type activeOrder struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}

func (e *Elevator) GoUp() {
	e.CurrentDirection = elevio.MD_Up
	elevio.SetMotorDirection(e.CurrentDirection)
}

func (e *Elevator) GoDown() {
	e.CurrentDirection = elevio.MD_Up
	elevio.SetMotorDirection(e.CurrentDirection)
}

func (e *Elevator) StopElevator() {
	e.CurrentDirection = elevio.MD_Stop
	elevio.SetMotorDirection(e.CurrentDirection)

}

func (e *Elevator) SetDoorState(state bool) {
	if state {
		e.doorOpen = true
	} else {
		e.doorOpen = false
	}
}

func (e *Elevator) NewActiveOrder(Order activeOrder) {
	e.ActiveOrders = append(e.ActiveOrders, Order)
}

func (e *Elevator) RemoveOrder(floor int, Button elevio.ButtonType) {
    for i, order := range e.ActiveOrders {
        if order.Floor == floor && order.Button == Button {
            // Remove the order from the slice
            e.ActiveOrders = append(e.ActiveOrders[:i], e.ActiveOrders[i+1:]...)
            break // Assuming only one order per floor/button combination, otherwise remove this
        }
    }
}

func (e *Elevator) SetButtonLamp(Button elevio.ButtonType, floor int, value bool) {
	elevio.SetButtonLamp(Button, floor, value)
}

func (e *Elevator) GetFloor() int {
	return elevio.GetFloor()
}

func (e *Elevator) SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func (e *Elevator) SetDoorOpenLamp(value bool) {
	elevio.SetDoorOpenLamp(value)
}

func (e *Elevator) SetMaster(value bool) {
	e.IsMaster = value
}

func (e *Elevator) ChooseBestOrder() activeOrder {

	//If no orders -> send current floor as order. Best to not call the function if there are no orders...
	if len(e.ActiveOrders) == 0 {
		var nullOrder activeOrder
		nullOrder.Floor = e.CurrentFloor
		nullOrder.Button = elevio.BT_Cab
		return nullOrder
	}

	bestOrder := e.ActiveOrders[0]

	for i := 0; i <= len(e.ActiveOrders); i++ {

		//Take orders on current floor first
		if e.CurrentFloor == e.ActiveOrders[i].Floor && e.CurrentDirection == elevio.MD_Stop {
			return e.ActiveOrders[i]
		}

		//Going upwards
		if e.CurrentDirection == elevio.MD_Up {

			//The best case order
			if e.CurrentFloor+1 == e.ActiveOrders[i].Floor && (e.ActiveOrders[i].Button != elevio.MD_Down || e.ActiveOrders[i].Floor == elevio._numFloors) {
				return e.ActiveOrders[i]
			}

			//Worst case - Neither order is above elevator search for closest order below current floor
			if e.CurrentFloor > e.ActiveOrders[i].Floor && e.CurrentFloor > bestOrder.Floor {
				if e.ActiveOrders[i].Floor > bestOrder.Floor {
					bestOrder = e.ActiveOrders[i]
					continue
				}
			}

			//Prioritize floors above current floor
			if e.CurrentFloor < e.ActiveOrders[i].Floor && e.CurrentFloor > bestOrder.Floor {
				bestOrder = e.ActiveOrders[i]
				continue
			}

			//Prioritize up and cab orders
			if e.ActiveOrders[i].Button != elevio.BT_HallDown && bestOrder.Button == elevio.BT_HallDown {
				bestOrder = e.ActiveOrders[i]
				continue
			}

			//Prioritize closest orders above elevator
			if e.ActiveOrders[i].Floor < bestOrder.Floor && e.ActiveOrders[i].Floor > e.CurrentFloor {
				bestOrder = e.ActiveOrders[i]
				continue
			}
		}

		//Going downwards
		if e.CurrentDirection == elevio.MD_Down {

			//The best case order
			if e.CurrentFloor-1 == e.ActiveOrders[i].Floor && (e.ActiveOrders[i].Button != elevio.MD_Up || e.ActiveOrders[i].Floor == 1) {
				return e.ActiveOrders[i]
			}

			//Worst case - Neither order is below elevator search for closest order above current floor
			if e.CurrentFloor < e.ActiveOrders[i].Floor && e.CurrentFloor < bestOrder.Floor {
				if e.ActiveOrders[i].Floor < bestOrder.Floor {
					bestOrder = e.ActiveOrders[i]
					continue
				}
			}

			//Prioritize floors below current floor
			if e.CurrentFloor > e.ActiveOrders[i].Floor && e.CurrentFloor < bestOrder.Floor {
				bestOrder = e.ActiveOrders[i]
				continue
			}

			//Prioritize down and cab orders
			if e.ActiveOrders[i].Button != elevio.BT_HallUp && bestOrder.Button == elevio.BT_HallUp {
				bestOrder = e.ActiveOrders[i]
				continue
			}

			//Prioritize closest orders below elevator
			if e.ActiveOrders[i].Floor > bestOrder.Floor && e.ActiveOrders[i].Floor < e.CurrentFloor {
				bestOrder = e.ActiveOrders[i]
				continue
			}
		}

		//Not moving
		if e.CurrentDirection == elevio.MD_Stop {

			//Prioritize cab orders
			if e.ActiveOrders[i].Button == elevio.BT_Cab && bestOrder.Button != elevio.BT_Cab {
				bestOrder = e.ActiveOrders[i]
				continue
			}

			//Prioritize closest cab orders
			if e.ActiveOrders[i].Button == elevio.BT_Cab && bestOrder.Button == elevio.BT_Cab {

				if (math.Abs(float64(e.CurrentFloor) - float64(e.ActiveOrders[i].Floor))) < (math.Abs(float64(e.CurrentFloor) - float64(e.ActiveOrders[i].Floor))) {
					bestOrder = e.ActiveOrders[i]
					continue
				}
			}

			//Not cab orders. Just pick closest?
			if e.ActiveOrders[i].Button != elevio.BT_Cab && bestOrder.Button != elevio.BT_Cab {

				if (math.Abs(float64(e.CurrentFloor) - float64(e.ActiveOrders[i].Floor))) < (math.Abs(float64(e.CurrentFloor) - float64(e.ActiveOrders[i].Floor))) {
					bestOrder = e.ActiveOrders[i]
					continue
				}
			}

			return e.ActiveOrders[0]
		}
	}

	return bestOrder
}

// TODO: Add function for removing order? What should the input be?
