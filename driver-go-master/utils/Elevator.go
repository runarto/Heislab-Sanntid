package utils

import (
	"fmt"
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

var (
	Elevators      []Elevator
	ElevatorsMutex sync.Mutex
)

var (
	Master      bool
	MasterMutex sync.Mutex
)

type Order struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}

func GoUp(e Elevator) Elevator {

	e.CurrentDirection = Up
	elevio.SetMotorDirection(e.CurrentDirection)
	e = SetState(Moving, e)
	return e
}

func GoDown(e Elevator) Elevator {

	e.CurrentDirection = Down
	elevio.SetMotorDirection(e.CurrentDirection)
	e = SetState(Moving, e)
	return e
}

func StopElevator(e Elevator) Elevator {
	// e.CurrentDirection = elevio.MD_Stop
	e.CurrentDirection = Stopped
	elevio.SetMotorDirection(elevio.MD_Stop)
	e = SetState(Still, e)
	return e
}

func SetDoorState(state bool, e Elevator) Elevator {
	if state {
		elevio.SetDoorOpenLamp(state)
		SetState(DoorOpen, e)
	} else {
		elevio.SetDoorOpenLamp(state)
		SetState(Still, e)
	}
	return e
}

func SetState(state State, e Elevator) Elevator {
	fmt.Println("Setting state to: ", state)
	e.CurrentState = state
	return e
}

func Obstruction(state bool, e Elevator) Elevator {
	if state {
		elevio.SetStopLamp(state)
		SetDoorState(state, e)
	} else {
		elevio.SetStopLamp(state)
		SetDoorState(state, e)
	}
	return e
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

func StopBtnPressed(btn bool, e Elevator) Elevator {
	elevio.SetMotorDirection(elevio.MD_Stop)
	if btn {
		if elevio.GetFloor() != NotDefined {
			elevio.SetStopLamp(true)
			e = SetDoorState(Open, e)
		}
	} else {
		elevio.SetStopLamp(false)
	}

	return e
}

func PrintLocalOrderArray(e Elevator) {
	for i := 0; i < NumButtons; i++ {
		for j := 0; j < NumFloors; j++ {
			fmt.Print(e.LocalOrderArray[i][j], " ")
		}
		fmt.Println()
	}
}
