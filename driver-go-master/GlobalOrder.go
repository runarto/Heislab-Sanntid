package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
	"fmt"
)


func FindBestElevator(button ButtonEvent, elevators Elevators) {

	orderDir := button.Button // Up or Down (0/1)
	orderFloor := button.Floor // The floor the order is from
	lowestCost := 1000 // Arbitrarily high number
	bestElevators := []Elevator

	for i := 0; i < len(elevators); i++ {
		costOfElevator := CalculateCost(elevators[i], button)
		if costOfElevator < lowestCost {
			lowestCost = costOfElevator
			bestElevators = append(bestElevators, elevators[i])
		}
	}

	if len(bestElevators) == 1 {
		chosenElevator := bestElevators[0]
		if chosenElevator.NetworkAdress == localAddress { // Need to check this somehow
			// This is the local elevator; handle the order locally
			HandleOrderLocally(button) // Pseudocode function to handle the order
		} else {
			// Send the order to the best elevator's network address
			SendOrder(chosenElevator.NetworkAdress, button)
		}
	} else if len(bestElevators) > 1 {
		// If there are multiple best elevators, choose the one with the fewest active orders
		chosenElevator := bestElevators[0] // Start with the first one as the best candidate
		for _, elevator := range bestElevators[1:] { // Start checking from the second item
			if len(elevator.ActiveOrders) < len(chosenElevator.ActiveOrders) {
				chosenElevator = elevator // Found a better candidate
			}
		}
	
		// Now that you have chosen the best elevator, check if it's local or remote
		if chosenElevator.NetworkAdress == localAddress {
			HandleOrderLocally(button) // Handle the order locally if it's the local elevator
		} else {
			SendOrder(chosenElevator.NetworkAdress, button) // Otherwise, send the order to the chosen elevator's network address
		}
	} else {
		// No best elevators found, handle this scenario appropriately
	}

}


// Idea: calculate the cost of each elevator and choose the one with the lowest cost

func CalculateCost(elevator Elevator, order ButtonEvent) int {

	// Goal: minimize the cost of the elevator, and travel time for any order. 
	
    // Determine if the elevator direction matches the order direction
    directionMatch := (elevator.CurrentDirection == Up && order.Floor > elevator.CurrentFloor) ||
                      (elevator.CurrentDirection == Down && order.Floor < elevator.CurrentFloor)

    if directionMatch {
        cost += 0 // Ideal as it's already going in the right direction
    } else {
        cost += 2 // Higher cost if it's going in the wrong direction or idle
    }

    // Consider the number of active orders as part of the cost
	for i := 0; i < len(elevator.ActiveOrders); i++ {
		cost += abs(order.Floor - elevator.ActiveOrders[i].Floor) // Add the distance to each active order
	}

	return cost 
}
