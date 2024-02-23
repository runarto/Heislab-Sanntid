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




func UpdateActiveElevators(elevator Elevator) {
	elevatorID = elevator.ElevatorID
	elevatorExists := false

	for _, elevator := range Elevators {
		if elevator.ElevatorID == elevatorID {
			elevator.isActive = true
			elevatorExists = true
			break
		}
	}

	if !elevatorExists {
		ActiveElevators = append(ActiveElevators, elevator)
	}

}



func (e *Elevator) MessageType (messageType int, messageBytes []byte, conn *net.UDPAddr) {

	switch messageType {
		case 0x01:
			var msg MessageGlobalOrderArray
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				log.Fatal(err)
			}
			
			globalOrderArray = msg.globalOrders

		case 0x02:
			var msg MessageNewOrder
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				log.Fatal(err)
			}

			fromElevator := msg.e // The elevator that sent the order
			newOrder := msg.newOrder // The new order
			toElevatorID := msg.toElevatorID // The elevator to send the order to

			UpdateActiveElevators(fromElevator) // Update the active elevators array, if needed
			
			// Logic for handling new order
			if e.isMaster { // If the elevator is the master
				if newOrder.Button == Cab { // If the order is a cab order
					// Update order arrays
				} else {
					// Calculate the best elevator for the order
					// Broadcast order to the best elevator
				} else {
					if toElevatorID == e.ElevatorID { // If the order is for the elevator
						// Update order arrays	

					}
				}
			}
			//

			

		case 0x03:
			var msg MessageOrderComplete
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				log.Fatal(err)
			}

			fromElevator := msg.e // The elevator that completed the order
			completedOrder := msg.order // The completed order
			fromElevatorID := msg.fromElevatorID // The elevator that completed the order

			UpdateActiveElevators(fromElevator) // Update the active elevators array, if needed
			// Logic for handling completed order


		case 0x04:
			var msg MessageElevator
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				log.Fatal(err)
			}

				// Check if the elevator is already in the ActiveElevators array
			newElevator := msg.e


			UpdateActiveElevators(newElevator)

			fmt.Println("Determining master")
			DetermineMaster() // Re-evaluate the master elevator
			

		case 0x05:
			var msg MessageAlive
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				log.Fatal(err)
			}

			message := msg.s
			fromElevator := msg.e

			if message == "Ping" {
				UpdateActiveElevators(fromElevator) // Update the active elevators array, if needed
				msg := MessageAlive{"Pong", e} // Create a pong message
				serializedMsg, err := msg.Serialize()
				if err != nil {
					log.Printf("Error serializing message: %v", err)
				}

				_, err = conn.Write(serializedMsg)
				if err != nil {
					return err
				}
			} else if message == "Pong" {
				fromElevator.resetCounter() // Reset the counter for the elevator
			}

			




		// Start counters for the elevators? 
		// If the counter reaches 20, the elevator is considered dead
		// If the elevator is considered dead, remove it from the ActiveElevators array
		// DetermineMaster() // Re-evaluate the master elevator
		
		
		}
}

func (e *Elevator) resetCounter() {
	for _, elevator := range ActiveElevators {
		if e.ID == elevator.ID {
			e.counter = 0
			// e.startCounter()
		}
	}
}