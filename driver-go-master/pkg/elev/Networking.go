package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func BroadcastElevatorStatus(e *utils.Elevator, statusTx chan utils.ElevatorStatus) {

	// BroadcastElevatorStatus broadcasts the elevator status periodically to other elevators.
	// It takes an elevator pointer and a channel for transmitting the elevator status.
	// The function sleeps for 5 seconds before starting the periodic broadcasting.
	// It uses a ticker to send the elevator status every 5 seconds, but only if there are more than one elevator in the system.
	// The elevator status message includes the type "ElevatorStatus" and the elevator information.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 {
			elevatorStatusMessage := utils.ElevatorStatus{
				Type:         "ElevatorStatus",
				FromElevator: *e,
			}

			statusTx <- elevatorStatusMessage
		}
	}
}

func BroadcastAckMatrix(e *utils.Elevator, ackTx chan utils.AckMatrix) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 && e.IsMaster {

			ackStruct := utils.AckMatrix{
				Type:           "AckMatrix",
				OrderWatcher:   utils.OrderWatcher, // Use the correct field name as defined in your ElevatorStatus struct
				FromElevatorID: e.ID,
			}

			ackTx <- ackStruct // Broadcast the current status
		}
	}
}

func UpdateElevatorsOnNetwork(e *utils.Elevator) {

	// UpdateElevatorsOnNetwork updates the elevator information in the ActiveElevators array.
	// It takes a pointer to an Elevator struct as input and updates the elevator with the same ID in the array.
	// If the elevator does not exist in the array, it adds the elevator to the array.
	// Finally, it prints the local order array for the elevator.

	elevatorID := e.ID      // The ID of the elevator
	elevatorExists := false // Flag to check if the elevator exists in the ActiveElevators array
	e.IsActive = true

	for i, _ := range utils.Elevators {
		if utils.Elevators[i].ID == elevatorID {
			utils.Elevators[i] = *e // Update the elevator
			elevatorExists = true   // Set the elevatorExists flag to true
			break
		}
	}

	if !elevatorExists { // If the elevator does not exist in the ActiveElevators array
		utils.Elevators = append(utils.Elevators, *e) // Add the elevator to the ActiveElevators array
	}

	fmt.Println("Local order array for elevator", e.ID)
	orders.PrintLocalOrderSystem(e)

}
