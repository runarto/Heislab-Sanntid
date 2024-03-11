package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

// Only local stuff

func OrderHandler(e utils.Elevator, ButtonCh chan elevio.ButtonEvent, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	NewOrder <-chan utils.MessageNewOrder, OrderComplete <-chan utils.MessageOrderComplete, PeerUpdateCh chan peers.PeerUpdate,
	DoOrderCh chan utils.Order, LocalStateUpdateCh chan utils.Elevator, MasterUpdateCh chan int, ch chan interface{}, IsOnlineCh chan bool, ActiveElevatorUpdate chan utils.Status,
	WatcherUpdate chan utils.OrderWatcher) {

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
					Order:         order,
					ForElevatorID: e.ID,
					IsComplete:    false,
					IsNew:         true}

				msg := utils.PackMessage("MessageNewOrder", order, utils.NotDefined, e.ID)
				ch <- msg

			} else {

				ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh, WatcherUpdate, IsOnlineCh)

			}

		case newOrder := <-NewOrder:

			order := newOrder.NewOrder

			if utils.Master && newOrder.ToElevatorID == utils.MasterID {

				fmt.Println("---NEW ORDER TO DELEGATE---")

				ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh, WatcherUpdate, IsOnlineCh)

			} else if !utils.Master && newOrder.ToElevatorID == e.ID {

				fmt.Println("---NEW ORDER RECEIVED---")

				DoOrderCh <- order

			} else if newOrder.ToElevatorID != e.ID {

				go func() {
					GlobalUpdateCh <- utils.GlobalOrderUpdate{
						Order:         order,
						ForElevatorID: newOrder.FromElevatorID,
						IsComplete:    false,
						IsNew:         true}
				}()
			}

		case orderComplete := <-OrderComplete:

			fmt.Println("---ORDER COMPLETE RECEIVED---")

			go ProcessOrderComplete(orderComplete, e, GlobalUpdateCh)

		case peerUpdate := <-PeerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, IsOnlineCh, MasterUpdateCh, ActiveElevatorUpdate)

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
