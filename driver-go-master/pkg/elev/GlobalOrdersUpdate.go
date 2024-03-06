package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func GlobalUpdates(channels *utils.Channels, thisElevator *utils.Elevator) {

	for {
		select {

		case elevatorStatus := <-channels.ElevatorStatusRx:

			fmt.Println("---ELEVATOR STATUS RECEIVED---")

			HandleNewElevatorStatus(elevatorStatus, &elevatorStatus.FromElevator, channels.GlobalUpdateCh)

		case GlobalUpdate := <-channels.GlobalUpdateCh:

			fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")

			HandleNewGlobalOrderUpdate(GlobalUpdate, thisElevator)

		case peerUpdate := <-channels.PeersOnlineCh:

			HandleNewPeerUpdate(peerUpdate)

		case WatcherUpdate := <-channels.OrderWatcher:

			if WatcherUpdate.New {

				for i, _ := range WatcherUpdate.Orders {

					if WatcherUpdate.Orders[i].Button != utils.Cab {

						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Active = true
						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Completed = false
						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Time = time.Now()

					}

				}
			}

			if WatcherUpdate.Complete {

				for i, _ := range WatcherUpdate.Orders {

					if WatcherUpdate.Orders[i].Button != utils.Cab {

						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Active = false
						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Completed = true
						utils.MasterOrderWatcher.HallOrderArray[WatcherUpdate.Orders[i].Button][WatcherUpdate.Orders[i].Floor].Time = time.Now()

					}
				}
			}

		case MasterOrderWatcher := <-channels.MasterOrderWatcherRx:

			utils.MasterOrderWatcher = MasterOrderWatcher.OrderWatcher

		}
	}
}

func HandleNewGlobalOrderUpdate(GlobalUpdate utils.GlobalOrderUpdate, thisElevator *utils.Elevator) {

	for i, _ := range GlobalUpdate.Orders {

		if GlobalUpdate.FromElevatorID == thisElevator.ID {

			if GlobalUpdate.IsComplete {

				orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], thisElevator.ID, false)

			} else {

				if !orders.CheckIfGlobalOrderIsActive(GlobalUpdate.Orders[i], thisElevator.ID) {

					orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], thisElevator.ID, true)
				}

			}

		} else if GlobalUpdate.FromElevatorID != thisElevator.ID {

			if GlobalUpdate.IsComplete {

				orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID, false)

			} else {

				if !orders.CheckIfGlobalOrderIsActive(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID) {

					orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID, true)
				}

			}

		}
	}
}

func HandleNewPeerUpdate(peerUpdate utils.NewPeersMessage) {

	if len(peerUpdate.NewPeers) > 0 {

		for i, _ := range peerUpdate.NewPeers {

			fmt.Println("New peer detected: ", peerUpdate.NewPeers[i])
			UpdateElevatorsOnNetwork(peerUpdate.NewPeers[i], true)

		}

	}

	if len(peerUpdate.LostPeers) > 0 {

		for i, _ := range peerUpdate.LostPeers {

			fmt.Println("Peer lost: ", peerUpdate.LostPeers[i])
			UpdateElevatorsOnNetwork(peerUpdate.LostPeers[i], false)

		}

	}

}

func HandleNewElevatorStatus(elevatorStatus utils.ElevatorStatus, thisElevator *utils.Elevator, GlobalOrderArrayUpdateCh chan utils.GlobalOrderUpdate) {

	if elevatorStatus.FromElevator.ID == thisElevator.ID {
		return
	}

	UpdateElevatorStatus(&elevatorStatus.FromElevator)

	var ActiveOrders []utils.Order

	for button := 0; button < utils.NumButtons; button++ {
		for floor := 0; floor < utils.NumFloors; floor++ {
			if elevatorStatus.FromElevator.LocalOrderArray[button][floor] == utils.True {

				order := utils.Order{
					Floor:  floor,
					Button: elevio.ButtonType(button)}

				ActiveOrders = append(ActiveOrders, order)

			}
		}
	}

	update := utils.GlobalOrderUpdate{
		Orders:         ActiveOrders,
		FromElevatorID: elevatorStatus.FromElevator.ID,
		IsComplete:     false,
		IsNew:          true}

	GlobalOrderArrayUpdateCh <- update

	DetermineMaster(thisElevator)

}

func UpdateElevatorStatus(fromElevator *utils.Elevator) {

	found := false

	for i, _ := range utils.Elevators {

		if utils.Elevators[i].ID == fromElevator.ID {

			utils.Elevators[i] = *fromElevator
			return
		}
	}

	if !found {

		utils.Elevators = append(utils.Elevators, *fromElevator)

	}

	UpdateElevatorsOnNetwork(fromElevator.ID, true)
}
