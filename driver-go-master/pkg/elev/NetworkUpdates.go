package elev

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func NetworkUpdate(channels *utils.Channels, thisElevator *utils.Elevator) {

	for {
		select {

		case peerUpdate := <-channels.PeerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, thisElevator, channels)

		case order := <-channels.NewOrderRx:

			fmt.Println("hello")

			if order.FromElevatorID != thisElevator.ID {

				fmt.Println("---NEW ORDER RECEIVED---")
				HandleNewOrder(order.NewOrder, order.FromElevatorID, order.ToElevatorID, thisElevator, channels)
			}

		case orderComplete := <-channels.OrderCompleteRx:

			if orderComplete.FromElevatorID != thisElevator.ID {

				fmt.Println("---ORDER COMPLETE RECEIVED---")
				HandleOrderComplete(orderComplete, channels, thisElevator)

			}

		}
	}

}
