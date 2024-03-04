package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

// Idea: calculate the cost of each elevator and choose the one with the lowest cost

func CalculateCost(e utils.Elevator, order utils.Order) int {

	cost := CheckAmountOfActiveOrders(&e)
	return cost

}

// Function to find the best elevator for a given order
func ChooseElevator(order utils.Order) *utils.Elevator {
	//Initiate variables
	var bestElevator utils.Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	for i, _ := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			cost := CalculateCost(utils.Elevators[i], order)
			if cost <= lowestCost {
				lowestCost = cost
				bestElevator = utils.Elevators[i]
			}
		}
	}

	return &bestElevator
}

func CheckIfGlobalOrderIsActive(order utils.Order, e *utils.Elevator) bool {

	HallOrderArray := utils.GlobalOrders.HallOrderArray
	CabOrderArray := utils.GlobalOrders.CabOrderArray

	if order.Button == utils.Cab {
		if CabOrderArray[e.ID][order.Floor] == utils.True {
			return true
		} else {
			return false
		}
	} else {
		if HallOrderArray[order.Button][order.Floor] == utils.True {
			return true
		} else {
			return false
		}
	}

}

func UpdateGlobalOrderSystem(order utils.Order, e *utils.Elevator, value bool) {
	if value {
		if order.Button == utils.Cab {
			utils.GlobalOrders.CabOrderArray[e.ID][order.Floor] = utils.True
		} else {
			utils.GlobalOrders.HallOrderArray[order.Button][order.Floor] = utils.True
			fmt.Println("Turning lamp on")
			elevio.SetButtonLamp(order.Button, order.Floor, true)
		}
	} else {
		if order.Button == utils.Cab {
			utils.GlobalOrders.CabOrderArray[e.ID][order.Floor] = utils.False
		} else {
			utils.GlobalOrders.HallOrderArray[order.Button][order.Floor] = utils.False
			fmt.Println("Turning lamp off")
			elevio.SetButtonLamp(order.Button, order.Floor, false)
		}
	}
}

func RedistributeHallOrders(offlineElevator *utils.Elevator, newOrderTx chan utils.MessageNewOrder, e *utils.Elevator) { // Should this perhaps be a pointer

	for button := 0; button < utils.NumButtons-1; button++ { // Don't change Cab-orders
		for floor := 0; floor < utils.NumFloors; floor++ {
			if offlineElevator.LocalOrderArray[button][floor] == utils.True {

				Order := utils.Order{
					Floor:  floor,
					Button: elevio.ButtonType(button)}

				UpdateGlobalOrderSystem(Order, offlineElevator, false)
				offlineElevator.LocalOrderArray[button][floor] = utils.False

				bestElevator := ChooseElevator(Order)

				UpdateGlobalOrderSystem(Order, bestElevator, true)

				if bestElevator.ID == e.ID {

					e.LocalOrderArray[button][floor] = utils.True
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)

				} else {

					newOrder := utils.MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     Order,
						E:            *bestElevator,
						ToElevatorID: bestElevator.ID}

					newOrderTx <- newOrder
				}

			}
		}
	}
}
