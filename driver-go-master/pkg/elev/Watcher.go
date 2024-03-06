package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func Bark(thisElevator *utils.Elevator, channels *utils.Channels) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		isMaster := thisElevator.IsMaster
		currentTime := time.Now()

		if isMaster {
			for button := 0; button < utils.NumButtons-1; button++ {
				for floor := 0; floor < utils.NumFloors; floor++ {
					timeSent := utils.MasterOrderWatcher.HallOrderArray[button][floor].Time

					if currentTime.Sub(timeSent) > utils.MasterTimeout &&
						!utils.MasterOrderWatcher.HallOrderArray[button][floor].Completed &&
						utils.MasterOrderWatcher.HallOrderArray[button][floor].Active {

						utils.MasterOrderWatcher.HallOrderArray[button][floor].Time = time.Now()

						// Resend the order to the network
						order := utils.Order{
							Floor:  floor,
							Button: elevio.ButtonType(button),
						}

						channels.MasterBarkCh <- order
					}
				}
			}
		} else {

			for button := 0; button < utils.NumButtons-1; button++ {
				for floor := 0; floor < utils.NumFloors; floor++ {
					timeSent := utils.SlaveOrderWatcher.HallOrderArray[button][floor].Time

					if currentTime.Sub(timeSent) > utils.SlaveTimeout &&
						!utils.SlaveOrderWatcher.HallOrderArray[button][floor].Confirmed {

						utils.SlaveOrderWatcher.HallOrderArray[button][floor].Time = time.Now()

						order := utils.Order{
							Floor:  floor,
							Button: elevio.ButtonType(button),
						}

						channels.SlaveBarkCh <- order
					}
				}
			}
		}
	}
}

func DetermineMaster(e *utils.Elevator) {

	// DetermineMaster determines the master elevator among the available elevators.
	// It sets the first active elevator as the initial master candidate and then compares it with other active elevators.
	// The elevator with the lowest ID becomes the master candidate.
	// Finally, it updates the local state and broadcasts the master election result if necessary.

	fmt.Println("Function: DetermineMaster")

	if len(utils.Elevators) == 0 {
		return // No elevators available
	}

	fmt.Println("Determining master")

	masterCandidate := utils.Elevators[0] // Set the first elevator as the master candidate

	for i, _ := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			masterCandidate = utils.Elevators[i] // Set the first active elevator as the master candidate
			break
		}
	}

	for i, _ := range utils.Elevators {
		fmt.Println("Elevator: ", utils.Elevators[i].ID, "isActive: ", utils.Elevators[i].IsActive)
		if utils.Elevators[i].IsActive {
			if utils.Elevators[i].ID < masterCandidate.ID { // If the elevator ID is less than the master candidate ID
				masterCandidate = utils.Elevators[i] // Set the elevator as the master candidate
			}
		}
	}

	for i, _ := range utils.Elevators {
		if masterCandidate.ID != utils.Elevators[i].ID {
			if utils.Elevators[i].IsMaster {
				utils.Elevators[i].IsMaster = false
			}
		}
	}

	// Set the masterCandidate as the master and update the local state as needed
	// This is a simplified representation; actual implementation may require additional synchronization and communication
	masterCandidate.IsMaster = true

	if masterCandidate.ID == e.ID {
		e.IsMaster = true
		fmt.Println("I am the master")
	} else {
		e.IsMaster = false
		fmt.Println("I am not the master")
		fmt.Println("The master is: ", masterCandidate.ID)

	}

	utils.MasterElevatorID = masterCandidate.ID // Set the master elevator ID

	// Broadcast or communicate the master election result as needed
	// This could involve sending a message to all elevators or updating a shared state
}

// To-Do function for handling motor stop?
// In the case of it happening, redistribute all active hall orders?

func Watchdog(channels *utils.Channels, thisElevator *utils.Elevator) {

	for {

		select {

		case order := <-channels.MasterBarkCh:

			fmt.Println("Master bark received, resending order", order)

			BestElevator := orders.ChooseElevator(order)

			if BestElevator.ID == thisElevator.ID {

				channels.ButtonCh <- elevio.ButtonEvent{
					Floor:  order.Floor,
					Button: order.Button,
				}

			} else {

				channels.NewOrderTx <- utils.MessageNewOrder{
					Type:           "MessageNewOrder",
					NewOrder:       order,
					ToElevatorID:   BestElevator.ID,
					FromElevatorID: thisElevator.ID}

			}

		case order := <-channels.SlaveBarkCh:

			fmt.Println("Slave bark received, resending order to master", order)

			channels.NewOrderTx <- utils.MessageNewOrder{
				Type:           "MessageNewOrder",
				NewOrder:       order,
				ToElevatorID:   utils.MasterElevatorID,
				FromElevatorID: thisElevator.ID}
		}
	}
}
