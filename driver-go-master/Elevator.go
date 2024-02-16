package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Elevator struct {
	CurrentState	 State                 // Current state of the elevator (Moving, Still, Stop)
	CurrentDirection elevio.MotorDirection // Elevator direction
	CurrentFloor     int                   // Last or current floor. Starts at 0. 
	doorOpen         bool                  // Door open/closed
	Obstruction      bool                  // Obstruction or not
	stopButton       bool                  // Stop button pressed or not
	LocalOrderArray [3][numFloors]int       // Array of active orders. First row is HallUp, second is HallDown, third is Cab
	NetworkAdress    string				   // IP address
	IsMaster         bool                  // Master or not
	ElevatorID 	 	 int				   // ID of the elevator (0, 1, 2, ...) 
		                                   // Perhaps be used for determining new master?
}

type Order struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}

func (e *Elevator) GoUp() {
	e.CurrentDirection = elevio.MD_Up
	elevio.SetMotorDirection(e.CurrentDirection)
	SetState(Moving)
}

func (e *Elevator) GoDown() {
	e.CurrentDirection = elevio.MD_Up
	elevio.SetMotorDirection(e.CurrentDirection)
	SetState(Moving)
}

func (e *Elevator) StopElevator() {
	// e.CurrentDirection = elevio.MD_Stop
	elevio.SetMotorDirection(e.CurrentDirection)
	SetState(Still)
}

func (e *Elevator) SetDoorState(state bool) {
	if state {
		e.doorOpen = true
		elevio.SetDoorOpenLamp(On)
	} else {
		e.doorOpen = false
		elevio.SetDoorOpenLamp(Off)
	}
}

func (e *Elevator) SetState(state State) {
	e.CurrentState = state
}

