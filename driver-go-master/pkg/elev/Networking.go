package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func BroadcastElevatorStatus(e *utils.Elevator, statusTx chan utils.ElevatorStatus) {

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 {
			elevatorStatusMessage := utils.ElevatorStatus{
				Type:         "ElevatorStatus",
				FromElevator: *e, // Use the correct field name as defined in your ElevatorStatus struct
			}

			statusTx <- elevatorStatusMessage // Broadcast the current status
		}
	}
}

func BroadcastAckMatrix(e *utils.Elevator, ackTx chan utils.AckMatrix) {

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 && e.IsMaster {
			ackStruct := utils.AckMatrix{
				Type:         "AckMatrix",
				OrderWatcher: utils.OrderWatcher, // Use the correct field name as defined in your ElevatorStatus struct
			}

			ackTx <- ackStruct // Broadcast the current status
		}
	}
}

func UpdateElevatorsOnNetwork(e *utils.Elevator) {
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
