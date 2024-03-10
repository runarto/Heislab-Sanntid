package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

// Only local stuff

func OrderHandler(e utils.Elevator, ButtonCh chan elevio.ButtonEvent, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	NewOrderRx <-chan utils.MessageNewOrder, OrderCompleteRx <-chan utils.MessageOrderComplete, PeerUpdateCh chan peers.PeerUpdate,
	DoOrderCh chan utils.Order, LocalStateUpdateCh chan utils.Elevator, ElevatorStatusRx <-chan utils.MessageElevatorStatus,
	MasterUpdateCh chan int, ch chan interface{}, IsOnlineCh chan bool, ActiveElevatorUpdate chan utils.Status) {

	for {

		select {

		case newOrder := <-ButtonCh:
			fmt.Println("---LOCAL BUTTON PRESSED---")

			order := utils.Order{
				Floor:  newOrder.Floor,
				Button: newOrder.Button,
			}

			if order.Button == utils.Cab {

				fmt.Println("Sending cab order to FSM...")
				DoOrderCh <- order // Send to FSM

				GlobalUpdateCh <- utils.GlobalOrderUpdate{
					Order:          order,
					FromElevatorID: e.ID,
					IsComplete:     false,
					IsNew:          true}

			} else {

				ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh)

			}

		case newOrder := <-NewOrderRx:

			order := newOrder.NewOrder

			if utils.Master && newOrder.FromElevatorID != e.ID {

				fmt.Println("---NEW ORDER RECEIVED---")

				ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh)

			} else if newOrder.ToElevatorID == e.ID {

				fmt.Println("---NEW ORDER RECEIVED---")

				DoOrderCh <- order

			}

		case orderComplete := <-OrderCompleteRx:

			fmt.Println("---ORDER COMPLETE RECEIVED---")

			ProcessOrderComplete(orderComplete, e, GlobalUpdateCh)

		case peerUpdate := <-PeerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, IsOnlineCh, MasterUpdateCh, ActiveElevatorUpdate)

		case E := <-ElevatorStatusRx:

			if E.FromElevator.ID != e.ID {

				fmt.Println("---ELEVATOR STATUS RECEIVED---")

				ProcessElevatorStatus(E.FromElevator)

			}

		case val := <-MasterUpdateCh:

			fmt.Println("---MASTER UPDATE RECEIVED---")
			fmt.Println("Master ID: ", val)

			utils.MasterMutex.Lock()
			utils.MasterIDmutex.Lock()
			if val == e.ID {
				utils.MasterID = val
				utils.Master = true
			} else {
				utils.MasterID = val
				utils.Master = false
			}
			utils.MasterIDmutex.Unlock()
			utils.MasterMutex.Unlock()

		}
	}

}
