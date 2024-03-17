package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func OrderHandler(e utils.Elevator, ButtonPressCh chan elevio.ButtonEvent, AllOrdersCh chan utils.GlobalOrderUpdate, OrderHandlerNetworkUpdateCh chan utils.Message, PeerUpdateCh chan peers.PeerUpdate,
	DoOrderCh chan utils.Order, LocalElevatorStateUpdateCh chan utils.Elevator, messageHandler chan utils.Message, IsOnlineCh chan bool, ActiveElevatorUpdate chan utils.Status,
	OfflineOrderCompleteCh chan utils.Order) {

	Online := false

	for {

		select {

		// Messages from network ----------------
		case update := <-OrderHandlerNetworkUpdateCh:
			fmt.Println("---ORDER HANDLER NETWORK UPDATE RECEIVED---")
			switch update.Type {
			case "MessageNewOrder":
				HandleNewOrder(update.Msg.(utils.MessageNewOrder), e, messageHandler, AllOrdersCh, DoOrderCh, Online, update.FromElevatorID, update.ToElevatorID)
			case "MessageOrderComplete":
				ProcessOrderComplete(update.Msg.(utils.MessageOrderComplete).Order, e, AllOrdersCh, update.FromElevatorID)
			}

		// Local events (and peer update)----------------

		case newOrder := <-ButtonPressCh:
			HandleButtonEvent(newOrder, e, messageHandler, AllOrdersCh, DoOrderCh, Online)
		case peerUpdate := <-PeerUpdateCh:
			fmt.Println("---PEER UPDATE RECEIVED---")
			HandlePeersUpdate(peerUpdate, IsOnlineCh, ActiveElevatorUpdate, &Online)

		// Offline updates ----------------
		case orderComplete := <-OfflineOrderCompleteCh:
			ProcessOrderComplete(orderComplete, e, AllOrdersCh, utils.ID)

		}
	}
}
