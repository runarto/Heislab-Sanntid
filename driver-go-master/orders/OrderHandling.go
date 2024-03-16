package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

// Only local stuff

func OrderHandler(e utils.Elevator, ButtonCh chan elevio.ButtonEvent, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	NewOrderRx <-chan utils.MessageNewOrder, OrderComplete <-chan utils.MessageOrderComplete, PeerUpdateCh chan peers.PeerUpdate,
	DoOrderCh chan utils.Order, LocalStateUpdateCh chan utils.Elevator, MasterUpdateCh chan int, ch chan interface{}, IsOnlineCh chan bool, ActiveElevatorUpdate chan utils.Status,
	WatcherUpdate chan utils.OrderWatcher, LocalOrdersUpdate chan [utils.NumButtons][utils.NumFloors]bool, continueChannel chan bool) {

	LocalOrders := [utils.NumButtons][utils.NumFloors]bool{}
	Online := false

	for {

		select {

		case newOrder := <-ButtonCh:
			HandleButtonEvent(newOrder, e, ch, GlobalUpdateCh, LocalOrders, DoOrderCh, WatcherUpdate, IsOnlineCh, Online)
		case newOrder := <-NewOrderRx:
			HandleNewOrder(newOrder, LocalOrders, e, ch, GlobalUpdateCh, DoOrderCh, WatcherUpdate, IsOnlineCh, Online)
		case orderComplete := <-OrderComplete:
			fmt.Println("---ORDER COMPLETE RECEIVED---")
			ProcessOrderComplete(orderComplete, e, GlobalUpdateCh)
		case peerUpdate := <-PeerUpdateCh:
			fmt.Println("---PEER UPDATE RECEIVED---")
			HandlePeersUpdate(peerUpdate, IsOnlineCh, MasterUpdateCh, ActiveElevatorUpdate, &Online)
		case val := <-MasterUpdateCh:
			fmt.Println("---MASTER UPDATE RECEIVED---")
			HandleMasterUpdate(val, e, ch, continueChannel)

		}
	}
}
