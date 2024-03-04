package utils

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Elevator struct {
	CurrentState     State // Current state of the elevator (Moving, Still, Stop)
	CurrentDirection elevio.MotorDirection
	GeneralDirection int                        // Elevator direction
	CurrentFloor     int                        // Last or current floor. Starts at 0.
	DoorOpen         bool                       // Door open/closed
	Obstructed       bool                       // Obstruction or not
	StopButton       bool                       // Stop button pressed or not
	LocalOrderArray  [NumButtons][NumFloors]int // Array of active orders. First row is HallUp, second is HallDown, third is Cab
	IsMaster         bool                       // Is the elevator master or not
	ID               int                        // ID of the elevator
	IsActive         bool                       // Is the elevator active or not
}

var Elevators []Elevator

type Order struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}

func (e *Elevator) GoUp() {

	e.CurrentDirection = Up
	e.GeneralDirection = Up
	elevio.SetMotorDirection(e.CurrentDirection)
	e.SetState(Moving)
}

func (e *Elevator) GoDown() {

	e.CurrentDirection = Down
	e.GeneralDirection = Down
	elevio.SetMotorDirection(e.CurrentDirection)
	e.SetState(Moving)
}

func (e *Elevator) StopElevator() {
	// e.CurrentDirection = elevio.MD_Stop
	e.GeneralDirection = Stopped
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.SetState(Still)
}

func (e *Elevator) SetDoorState(state bool) {
	if state {
		e.DoorOpen = true
		elevio.SetDoorOpenLamp(state)
	} else {
		e.DoorOpen = false
		elevio.SetDoorOpenLamp(state)
	}
}

func (e *Elevator) SetState(state State) {
	e.CurrentState = state
}

func (e *Elevator) Obstruction(state bool) {
	if state {
		e.Obstructed = true
		elevio.SetStopLamp(state)
		e.SetDoorState(state)
	} else {
		e.Obstructed = false
		elevio.SetStopLamp(state)
		e.SetDoorState(state)
	}
}

func (e *Elevator) CheckIfMaster() bool {
	if e.IsMaster {
		return true
	} else {
		return false
	}
}

func (e *Elevator) SetLights() {
	for button := 0; button < NumButtons; button++ {
		for floor := 0; floor < NumFloors; floor++ {
			if e.LocalOrderArray[button][floor] == True {
				elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)
			}
		}
	}
}

func (e *Elevator) StopBtnPressed(btn bool) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	if btn {
		if elevio.GetFloor() != NotDefined {
			e.StopButton = true
			elevio.SetStopLamp(true)
			e.SetDoorState(Open)
		}
	} else {
		e.StopButton = false
		elevio.SetStopLamp(false)
	}
}
