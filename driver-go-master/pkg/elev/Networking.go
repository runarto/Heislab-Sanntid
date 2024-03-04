package elev

import (
	"time"

	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func BroadcastElevatorStatus(elevator utils.Elevator, statusTx chan utils.ElevatorStatus) {

	if len(utils.Elevators) > 1 {
		elevatorStatusMessage := utils.ElevatorStatus{
			Type:      "ElevatorStatus",
			E:         elevator, // Use the correct field name as defined in your ElevatorStatus struct
			AckStruct: utils.OrderWatcher,
		}

		// Initial broadcast
		statusTx <- elevatorStatusMessage

		// Optional: Periodic updates
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for range ticker.C {
			statusTx <- elevatorStatusMessage // Broadcast the current status
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

}
