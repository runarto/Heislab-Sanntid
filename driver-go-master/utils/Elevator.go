package utils

import (
	"fmt"
	"math"
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
		e = SetDoorState(state, e)
	} else {
		elevio.SetStopLamp(state)
		e = SetDoorState(state, e)
	}
	e = SetState(DoorOpen, e)
	return e
}

func CalculateCost(e Elevator, order Order) int {
	cost := 0
	cost += int(math.Abs(float64(e.CurrentFloor - order.Floor)))

	var orderDirection int
	if order.Button == HallUp {
		orderDirection = Up
	} else if order.Button == HallDown {
		orderDirection = Down
	}

	for f := 0; f < NumFloors; f++ {
		if e.LocalOrderArray[HallUp][f] || e.LocalOrderArray[HallDown][f] {
			cost += int(math.Abs(float64(f - order.Floor)))
			if (e.LocalOrderArray[HallUp][f] && orderDirection == Down) ||
				(e.LocalOrderArray[HallDown][f] && orderDirection == Up) {
				cost += 2
			}
		}
	}

	switch e.CurrentState {
	case Moving:
		if int(e.CurrentDirection) == orderDirection {
			cost -= 1
		} else {
			cost += 2
		}
	case Still:
		cost -= 2
	case DoorOpen:
		cost += 1
	}

	return cost
}

// Function to find the best elevator for a given order
func ChooseElevator(order Order) Elevator {
	//Initiate variables
	var BestElevator Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	ElevatorsMutex.Lock()
	defer ElevatorsMutex.Unlock()
	for i := range Elevators {
		if Elevators[i].IsActive {
			cost := CalculateCost(Elevators[i], order)
			fmt.Println("Cost for elevator ", Elevators[i].ID, " is: ", cost)
			if cost <= lowestCost {
				lowestCost = cost
				BestElevator = Elevators[i]
			}
		}
	}

	return BestElevator
}
