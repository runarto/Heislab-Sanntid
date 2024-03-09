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
			fmt.Println("Button pressed")

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

			fmt.Println("New order received")

			order := newOrder.NewOrder

			if utils.Master && newOrder.FromElevatorID != e.ID {

				ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh)

			} else if newOrder.ToElevatorID == e.ID {

				DoOrderCh <- order

			}

		case orderComplete := <-OrderCompleteRx:

			if e.ID != orderComplete.FromElevatorID {

				fmt.Println("Order complete received")

				ProcessOrderComplete(orderComplete, e, GlobalUpdateCh)

			}

		case peerUpdate := <-PeerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, IsOnlineCh, MasterUpdateCh, ActiveElevatorUpdate)

		case E := <-ElevatorStatusRx:

			fmt.Println("Elevator status received")

			if E.FromElevator.ID != e.ID {

				ProcessElevatorStatus(E.FromElevator)

			}

		case val := <-MasterUpdateCh:

			fmt.Println("Master update received")
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
