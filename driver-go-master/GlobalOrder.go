package main

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/elevio"
)

// Idea: calculate the cost of each elevator and choose the one with the lowest cost

func CalculateCost(elevator Elevator, order Order) int {

	cost := elevator.CheckAmountOfActiveOrders()
	return cost

}

// Function to find the best elevator for a given order
func chooseElevator(elevators []Elevator, order Order) Elevator {
	//Initiate variables
	var bestElevator Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	for i, _ := range elevators {
		if Elevators[i].isActive {
			cost := CalculateCost(Elevators[i], order)
			if cost <= lowestCost {
				lowestCost = cost
				bestElevator = Elevators[i]
			}
		}
	}

	return bestElevator
}

func CheckIfGlobalOrderIsActive(order Order, Elevator Elevator) bool {

	HallOrderArray := globalOrderArray.HallOrderArray
	CabOrderArray := globalOrderArray.CabOrderArray

	if order.Button == Cab {
		if CabOrderArray[Elevator.ID][order.Floor] == True {
			return true
		} else {
			return false
		}
	} else {
		if HallOrderArray[order.Button][order.Floor] == True {
			return true
		} else {
			return false
		}
	}

}

func UpdateGlobalOrderSystem(order Order, Elevator Elevator, value bool) {

	HallOrderArray := globalOrderArray.HallOrderArray
	CabOrderArray := globalOrderArray.CabOrderArray

	if value {
		if order.Button == Cab {
			CabOrderArray[Elevator.ID][order.Floor] = True
		} else {
			HallOrderArray[order.Button][order.Floor] = True
			fmt.Println("Turning lamp on")
			elevio.SetButtonLamp(order.Button, order.Floor, true)
		}
	} else {
		if order.Button == Cab {
			CabOrderArray[Elevator.ID][order.Floor] = False
		} else {
			HallOrderArray[order.Button][order.Floor] = False
			fmt.Println("Turning lamp off")
			elevio.SetButtonLamp(order.Button, order.Floor, false)
		}
	}

}

func (e *Elevator) RedistributeHallOrders(offlineElevator Elevator, newOrderTx chan MessageNewOrder) { // Should this perhaps be a pointer

	for button := 0; button < numButtons-1; button++ { // Don't change Cab-orders
		for floor := 0; floor < numFloors; floor++ {
			if offlineElevator.LocalOrderArray[button][floor] == True {

				Order := Order{floor, elevio.ButtonType(button)}

				UpdateGlobalOrderSystem(Order, offlineElevator, false)
				offlineElevator.LocalOrderArray[button][floor] = False

				bestElevator := chooseElevator(Elevators, Order)

				UpdateGlobalOrderSystem(Order, bestElevator, true)

				if bestElevator.ID == e.ID {

					e.LocalOrderArray[button][floor] = True
					elevio.SetButtonLamp(elevio.ButtonType(button), floor, true)

				} else {

					newOrder := MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     Order,
						E:            bestElevator,
						ToElevatorID: bestElevator.ID}

					newOrderTx <- newOrder
				}

			}
		}
	}
}
