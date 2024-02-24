package main

import (
	"math"
	"fmt"
	"time"
	"sync"
)


func AbsValue(x int, y int) int { 
	return int(math.Abs(float64(x - y)))
}


func (e *Elevator) HandleElevatorAtFloor(floor int) {

	if e.ElevatorAtFloor(floor) { // Check for active orders at floor
		e.StopElevator() // Stop the elevator
		e.SetDoorState(Open) // Open the door
		time.Sleep(1000 * time.Millisecond) // Wait for a second
		e.SetDoorState(Close) // Close the door
		fmt.Println("Order system: ")
		e.PrintLocalOrderSystem()
		amountOfOrders := e.CheckAmountOfActiveOrders() // Check the amount of active orders
		fmt.Println("Amount of active orders: ", amountOfOrders)
		if amountOfOrders > 0 {
			bestOrder = e.ChooseBestOrder() // Choose the best order
			fmt.Println("Best order: ", bestOrder)
			e.DoOrder(bestOrder)
			// DoOrder(order) // Move the elevator to the best order (pseudocode function to move the elevator to the best order
		} else {
			e.SetState(Still) // If no orders, set the state to still
		}
	}
}

func (e *Elevator) CheckIfOrderIsActive(order Order) bool {

	if e.LocalOrderArray[order.Button][order.Floor] == True {
		return true
	} else {
		return false
	}
}


func GlobalOrderSystemReceived(globalOrders GlobalOrderArray) {
	globalOrderArray = globalOrders
	// Update the local order system with the global order system

}



func DetermineMaster() {
    if len(Elevators) == 0 {
        return // No elevators available
    }

    // Start with the first elevator as the initial candidate for master
    masterCandidate := Elevators[0]

    // Iterate through the elevators to find the one with the lowest ElevatorID
    for _, elevator := range ActiveElevators[1:] {
        if elevator.ElevatorID < masterCandidate.ElevatorID {
            masterCandidate = elevator
        }
    }

    // Set the masterCandidate as the master and update the local state as needed
    // This is a simplified representation; actual implementation may require additional synchronization and communication
    masterCandidate.isMaster = true
	masterElevatorIP = masterCandidate.ElevatorIP

    // Broadcast or communicate the master election result as needed
    // This could involve sending a message to all elevators or updating a shared state
}




func UpdateActiveElevators(e Elevator) {
	elevatorID = e.ElevatorID // The ID of the elevator
	elevatorExists := false // Flag to check if the elevator exists in the ActiveElevators array

	for _, elevator := range Elevators {
		if elevator.ElevatorID == elevatorID {
			elevator = e // Update the elevator
			elevatorExists = true // Set the elevatorExists flag to true
			break
		}
	}

	if !elevatorExists { // If the elevator does not exist in the ActiveElevators array
		ActiveElevators = append(ActiveElevators, elevator) // Add the elevator to the ActiveElevators array
	}

}




