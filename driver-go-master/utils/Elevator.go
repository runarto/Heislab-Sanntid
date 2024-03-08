package utils

import (
	"sync"

	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Elevator struct {
	CurrentState     State // Current state of the elevator (Moving, Still, Stop)
	CurrentDirection elevio.MotorDirection
	CurrentFloor     int                         // Last or current floor. Starts at 0.
	LocalOrderArray  [NumButtons][NumFloors]bool // Array of active orders. First row is HallUp, second is HallDown, third is Cab
	ID               int                         // ID of the elevator
	IsActive         bool                        // Is the elevator active or not
}

var Elevators []Elevator

var (
	Master      bool
	MasterMutex sync.Mutex
)

type Order struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}

func (e *Elevator) GoUp() {

	e.CurrentDirection = Up
	elevio.SetMotorDirection(e.CurrentDirection)
	e.SetState(Moving)
}

func (e *Elevator) GoDown() {

	e.CurrentDirection = Down
	elevio.SetMotorDirection(e.CurrentDirection)
	e.SetState(Moving)
}

func (e *Elevator) StopElevator() {
	// e.CurrentDirection = elevio.MD_Stop
	e.CurrentDirection = Stopped
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.SetState(Still)
}

func (e *Elevator) SetDoorState(state bool) {
	if state {
		elevio.SetDoorOpenLamp(state)
		e.SetState(DoorOpen)
	} else {
		elevio.SetDoorOpenLamp(state)
		e.SetState(Still)
	}
}

func (e *Elevator) SetState(state State) {
	e.CurrentState = state
}

func (e *Elevator) Obstruction(state bool) {
	if state {
		elevio.SetStopLamp(state)
		e.SetDoorState(state)
	} else {
		elevio.SetStopLamp(state)
		e.SetDoorState(state)
	}
}

func CheckIfMaster() bool {
	if Master {
		return true
	} else {
		return false
	}
}

func (e *Elevator) SetLights() {
	for b := 0; b < NumButtons; b++ {
		for f := 0; f < NumFloors; f++ {
			if e.LocalOrderArray[b][f] {
				elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			}
		}
	}
}

func (e *Elevator) StopBtnPressed(btn bool) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	if btn {
		if elevio.GetFloor() != NotDefined {
			elevio.SetStopLamp(true)
			e.SetDoorState(Open)
		}
	} else {
		elevio.SetStopLamp(false)
	}
}
