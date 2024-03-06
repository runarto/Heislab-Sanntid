package orders

import (
	"fmt"
	"math"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func CalculateCost(e utils.Elevator, order utils.Order) int {

	cost := 0

	switch e.CurrentDirection {

	case utils.Up:

		if order.Floor > e.CurrentFloor {
			cost += order.Floor - e.CurrentFloor
		} else if order.Floor < e.CurrentFloor {
			cost += 2*utils.NumFloors - order.Floor - e.CurrentFloor
		} else {
			cost += 0
		}

	case utils.Down:
		if order.Floor < e.CurrentFloor {
			cost += e.CurrentFloor - order.Floor
		} else if order.Floor > e.CurrentFloor {
			cost += 2*utils.NumFloors - e.CurrentFloor - order.Floor
		} else {
			cost += 0
		}

	case utils.Stopped:
		cost += int(math.Abs(float64(order.Floor - e.CurrentFloor)))
	}

	for button := 0; button < utils.NumButtons; button++ {
		for floor := 0; floor < utils.NumFloors; floor++ {
			if e.LocalOrderArray[button][floor] == utils.True {
				cost += 2
			}
		}
	}

	return cost

}

// Function to find the best elevator for a given order
func ChooseElevator(order utils.Order) *utils.Elevator {
	//Initiate variables
	var BestElevator utils.Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	for i, _ := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			cost := CalculateCost(utils.Elevators[i], order)
			if cost <= lowestCost {
				lowestCost = cost
				BestElevator = utils.Elevators[i]
			}
		}
	}

	return &BestElevator
}

func CheckIfGlobalOrderIsActive(order utils.Order, ElevatorID int) bool {

	if order.Button == utils.Cab {
		if utils.GlobalOrders.CabOrderArray[ElevatorID][order.Floor] == utils.True {
			return true
		} else {
			return false
		}
	} else {
		if utils.GlobalOrders.HallOrderArray[order.Button][order.Floor] == utils.True {
			return true
		} else {
			return false
		}
	}

}

func UpdateGlobalOrderSystem(order utils.Order, ElevatorID int, value bool) {

	if value {
		if order.Button == utils.Cab {
			utils.GlobalOrders.CabOrderArray[ElevatorID][order.Floor] = utils.True
		} else {
			utils.GlobalOrders.HallOrderArray[order.Button][order.Floor] = utils.True
			fmt.Println("Turning lamp on")
			elevio.SetButtonLamp(order.Button, order.Floor, true)
		}
	} else {
		if order.Button == utils.Cab {
			utils.GlobalOrders.CabOrderArray[ElevatorID][order.Floor] = utils.False
		} else {
			utils.GlobalOrders.HallOrderArray[order.Button][order.Floor] = utils.False
			fmt.Println("Turning lamp off")
			elevio.SetButtonLamp(order.Button, order.Floor, false)
		}
	}

}


