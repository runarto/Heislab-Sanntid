package elev

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/pkg/utils"
	"github.com/runarto/Heislab-Sanntid/Network/peers"
)

func NetworkUpdate(channels *utils.Channels, thisElevator *utils.Elevator, peerUpdateCh chan peers.PeerUpdate) {

	for {
		select {

		case peerUpdate := <- peerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, thisElevator, channels)

		case order := <-channels.NewOrderRx:

			if order.FromElevatorID != thisElevator.ID {

				fmt.Println("---NEW ORDER RECEIVED---")
				HandleNewOrder(order.NewOrder, order.FromElevatorID, order.ToElevatorID, thisElevator, channels)
			}

		case orderComplete := <-channels.OrderCompleteRx:

			if orderComplete.FromElevatorID != thisElevator.ID {

				fmt.Println("---ORDER COMPLETE RECEIVED---")
				HandleOrderComplete(orderComplete, channels.GlobalUpdateCh, thisElevator)

			}

		}
	}

}
