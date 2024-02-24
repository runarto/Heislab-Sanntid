package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Elevator struct {
	CurrentState	 State                 // Current state of the elevator (Moving, Still, Stop)
	CurrentDirection elevio.MotorDirection 
	GeneralDirection int 				   // Elevator direction
	CurrentFloor     int                   // Last or current floor. Starts at 0. 
	doorOpen         bool                  // Door open/closed
	Obstruction      bool                  // Obstruction or not
	stopButton       bool                  // Stop button pressed or not
	LocalOrderArray [3][numFloors]int       // Array of active orders. First row is HallUp, second is HallDown, third is Cab
	isMaster 		bool				   // Is the elevator master or not
	ID				int					   // ID of the elevator		
	isActive 		bool				   // Is the elevator active or not                   	              
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
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.SetState(Still)
}

func (e *Elevator) SetDoorState(state bool) {
	if state {
		e.doorOpen = true
		elevio.SetDoorOpenLamp(state)
	} else {
		e.doorOpen = false
		elevio.SetDoorOpenLamp(state)
	}
}

func (e *Elevator) SetState(state State) {
	e.CurrentState = state
}

func (e *Elevator) isObstruction(state bool) {
	if state {
		e.Obstruction = true
	} else {
		e.Obstruction = false
	}
}

func (e *Elevator) StopButton(state bool) {
	if state {
		if elevio.GetFloor() != NotDefined {
			e.SetDoorState(Open)
		}
	}
}

func (e* Elevator) CheckIfMaster() bool {
	if e.isMaster {
		return true
	} else {
		return false
	}
}

