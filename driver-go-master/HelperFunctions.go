package main

import (
	"math"
	"fmt"
	"time"
)


func AbsValue(x int, y int) int { 
	return int(math.Abs(float64(x - y)))
}


func (e *Elevator) HandleElevatorAtFloor(floor int, OrderCompleteTx chan MessageOrderComplete) {

	if e.HandleOrdersAtFloor(floor, OrderCompleteTx) { // If true, orders have been handled at the floor

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

			e.DoOrder(bestOrder, OrderCompleteTx) // Set elevator in direction of best order
			
		} else {

			e.SetState(Still) // If no orders, set the state to still
			e.GeneralDirection = Stopped
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


// func GlobalOrderSystemReceived(globalOrders GlobalOrderArray) {
// 	globalOrderArray = globalOrders
// 	// Update the local order system with the global order system

// }



func (e *Elevator) DetermineMaster() {

    if len(Elevators) == 0 {
        return // No elevators available
    }

	fmt.Println("Determining master")

	masterCandidate := Elevators[0] // Set the first elevator as the master candidate

	for _, elevator := range Elevators {
		if elevator.isActive {
			masterCandidate = elevator // Set the first active elevator as the master candidate
			break
		}
	}

	for _, elevator := range Elevators {
		fmt.Println("Elevator: ", elevator.ID, "isActive: ", elevator.isActive)
		if elevator.isActive {
			if elevator.ID < masterCandidate.ID { // If the elevator ID is less than the master candidate ID
				masterCandidate = elevator // Set the elevator as the master candidate
			}
		}
	}

   

    // Set the masterCandidate as the master and update the local state as needed
    // This is a simplified representation; actual implementation may require additional synchronization and communication
    masterCandidate.isMaster = true
	fmt.Println("The master now is elevator: ", masterCandidate.ID)
	if masterCandidate.ID == e.ID {
		e.isMaster = true
		fmt.Println("I am the master")
	}

    // Broadcast or communicate the master election result as needed
    // This could involve sending a message to all elevators or updating a shared state
}




func UpdateElevatorsOnNetwork(e Elevator) {
	elevatorID := e.ID // The ID of the elevator
	elevatorExists := false // Flag to check if the elevator exists in the ActiveElevators array

	for _, elevator := range Elevators {
		if elevator.ID == elevatorID {
			elevator = e // Update the elevator
			elevatorExists = true // Set the elevatorExists flag to true
			break
		}
	}

	if !elevatorExists { // If the elevator does not exist in the ActiveElevators array
		Elevators = append(Elevators, e) // Add the elevator to the ActiveElevators array
	}

}




