package orders

import (
	"fmt"
	"math"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func CalculateCost(e utils.Elevator, order utils.Order) int {

	cost := 0
	cost += int(math.Abs(float64(e.CurrentFloor - order.Floor)))
	orderDirection := order.Button

	for b := 0; b < utils.NumButtons; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if e.LocalOrderArray[b][f] {
				if b == utils.Cab {
					cost += 0
				} else if b == int(orderDirection) {
					cost += 0
				} else {
					cost += 2
				}

				cost += int(math.Abs(float64(f - order.Floor)))
			}
		}
	}

	return cost
}

// Function to find the best elevator for a given order
func ChooseElevator(order utils.Order) utils.Elevator {
	//Initiate variables
	var BestElevator utils.Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	for i := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			cost := CalculateCost(utils.Elevators[i], order)
			fmt.Println("Cost for elevator ", utils.Elevators[i].ID, " is: ", cost)
			if cost <= lowestCost {
				lowestCost = cost
				BestElevator = utils.Elevators[i]
			}
		}
	}

	return BestElevator
}
