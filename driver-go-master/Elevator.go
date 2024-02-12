package main


import (
    "github.com/runarto/Heislab-Sanntid/elevio"
)


type Elevator struct {
    CurrentDirection elevio.MotorDirection // Elevator direction
    doorOpen bool // Door open/closed
	Obstruction bool // Obstruction or not
	stopButton bool // Stop button pressed or not
	ActiveOrders  []activeOrder //List of structs 
	NetworkAdress string
	IsMaster bool
	
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

func (e *Elevator) NewOrder(Order activeOrder) {
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