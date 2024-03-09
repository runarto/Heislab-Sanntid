package net

import (
	"time"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func BroadcastElevatorStatus(e utils.Elevator, ch chan interface{}) {

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

			msg := utils.PackMessage("ElevatorStatus", e)
			ch <- msg

		}
	}
}

func BroadcastMasterOrderWatcher(e utils.Elevator, ch chan interface{}) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {

		if len(utils.Elevators) > 1 && utils.Master {

			msg := utils.PackMessage("MessageOrderWatcher", utils.MasterOrderWatcher.HallOrderArray,
				utils.MasterOrderWatcher.CabOrderArray, e.ID)

			ch <- msg
		}
	}
}
