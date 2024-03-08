package orders

import (
	"math"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func CalculateCost(e utils.Elevator, order utils.Order) int {
	return int(math.Abs(float64(e.CurrentFloor - order.Floor)))
}

// Function to find the best elevator for a given order
func ChooseElevator(order utils.Order) utils.Elevator {
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

	return BestElevator
}
