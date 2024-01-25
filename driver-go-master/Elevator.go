package main


import (
    "Heislab-Sanntid/elevio"
)


type Elevator struct {
    CurrentDirection elevio.MotorDirection // Elevator direction
    doorOpen bool // Door open/closed
	Obstruction bool // Obstruction or not
	stopButton bool // Stop button pressed or not
	ActiveOrders  []activeOrder //List of structs 
	NetworkAdress string
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
	if state{
		e.doorOpen = true
	} else {
		e.doorOpen = false
	}
}

func (e *Elevator) NewActiveOrder(Order activeOrder) {
	e.ActiveOrders = append(e.ActiveOrders, Order)
}

// TODO: Add function for removing order? What should the input be?
