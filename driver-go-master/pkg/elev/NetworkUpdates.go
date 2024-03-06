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

			HandlePeersUpdate(peerUpdate, channels.ElevatorStatusTx, channels.OrderArraysTx, channels.NewOrderTx, thisElevator)

		case order := <-channels.NewOrderRx:

			fmt.Println("---NEW ORDER RECEIVED---")

			HandleNewOrder(order.NewOrder, order.FromElevatorID, order.ToElevatorID, channels.OrderCompleteTx,
				channels.NewOrderTx, thisElevator, channels.ButtonCh, channels.GlobalUpdateCh)

		case orderComplete := <-channels.OrderCompleteRx:

			fmt.Println("---ORDER COMPLETE RECEIVED---")

			HandleOrderComplete(orderComplete, channels.GlobalUpdateCh, thisElevator)

		}
	}

}
