package orders

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/utils"
)

// Only local stuff

func OrderHandler(c *utils.Channels, e utils.Elevator) {

	for {

		select {

		case newOrder := <-c.ButtonCh:
			fmt.Println("Button pressed")

			order := utils.Order{
				Floor:  newOrder.Floor,
				Button: newOrder.Button,
			}

			if order.Button == utils.Cab {

				c.DoOrderCh <- order // Send to FSM

				if !utils.Master {
					SendOrder(order, e, c, utils.MasterID) // Send to network
				} else {

					c.GlobalUpdateCh <- utils.GlobalOrderUpdate{
						Order:          order,
						FromElevatorID: e.ID,
						IsComplete:     false,
						IsNew:          true}

				}

			} else {

				ProcessNewOrder(order, e, c)

			}

		case newOrder := <-c.NewOrderRx:

			fmt.Println("New order received")

			order := newOrder.NewOrder

			if utils.Master && newOrder.FromElevatorID != e.ID {

				ProcessNewOrder(order, e, c)

			} else if newOrder.ToElevatorID == e.ID {

				c.DoOrderCh <- order

			}

		case orderComplete := <-c.OrderCompleteRx:

			fmt.Println("Order complete received")

			if e.ID != orderComplete.FromElevatorID {

				ProcessOrderComplete(orderComplete, e, c)

			}

		case peerUpdate := <-c.PeerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			HandlePeersUpdate(peerUpdate, c)

			c.LocalStateUpdateCh <- e // Update the local elevator instance

		case E := <-c.ElevatorStatusRx:

			fmt.Println("Elevator status received")

			if E.FromElevator.ID != e.ID {

				ProcessElevatorStatus(E.FromElevator, c)

			}

		case val := <-c.MasterUpdateCh:

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
