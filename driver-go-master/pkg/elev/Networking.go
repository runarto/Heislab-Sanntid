package elev

import (
	"fmt"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
	"github.com/runarto/Heislab-Sanntid/elevio"
)

func BroadcastElevatorStatus(thisElevator *utils.Elevator, channels *utils.Channels) {

	// BroadcastElevatorStatus broadcasts the elevator status periodically to other elevators.
	// It takes an elevator pointer and a channel for transmitting the elevator status.
	// The function sleeps for 5 seconds before starting the periodic broadcasting.
	// It uses a ticker to send the elevator status every 5 seconds, but only if there are more than one elevator in the system.
	// The elevator status message includes the type "ElevatorStatus" and the elevator information.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 {

			elevatorStatusMessage := utils.ElevatorStatus{
				Type:         "ElevatorStatus",
				FromElevator: *thisElevator,
			}

			channels.ElevatorStatusTx <- elevatorStatusMessage
		}
	}
}

func BroadcastMasterOrderWatcher(thisElevator *utils.Elevator, OrderWatcherCh chan utils.MessageOrderWatcher) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {

		if len(utils.Elevators) > 1 && thisElevator.IsMaster {

			MasterOrderWatcher := utils.MessageOrderWatcher{
				Type:           "AckMatrix",
				OrderWatcher:   utils.MasterOrderWatcher, // Use the correct field name as defined in your ElevatorStatus struct
				FromElevatorID: thisElevator.ID,
			}

			OrderWatcherCh <- MasterOrderWatcher // Broadcast the current status
		}
	}
}

func UpdateElevatorsOnNetwork(ElevatorID int, active bool) {

	if active {

		for i, _ := range utils.Elevators {
			if utils.Elevators[i].ID == ElevatorID {
				utils.Elevators[i].IsActive = true
				return
			}
		}

	} else {

		for i, _ := range utils.Elevators {
			if utils.Elevators[i].ID == ElevatorID {
				utils.Elevators[i].IsActive = false
				return
			}
		}

	}

	fmt.Println("Elevator not found, wait for status update")

}

func HandleOrderComplete(orderComplete utils.MessageOrderComplete, GlobalUpdateCh chan utils.GlobalOrderUpdate, thisElevator *utils.Elevator) {

	// HandleOrderComplete handles an order completion message.
	// It takes an order completion message as input and updates the local and global order systems.

	if orderComplete.FromElevatorID == thisElevator.ID {

		return

	}

	fmt.Println("Function: HandleOrderComplete")

	ordersDone := orderComplete.Orders

	orders := utils.GlobalOrderUpdate{
		Orders:         ordersDone,
		FromElevatorID: orderComplete.FromElevatorID,
		New: false,
		Completed: true,
		}

	GlobalUpdateCh <- orders

}

func HandlePeersUpdate(p peers.PeerUpdate, thisElevator *utils.Elevator, channels *utils.Channels) {

	// HandlePeersUpdate handles the update of peers in the system.
	// It receives a PeerUpdate struct containing information about the updated peers.
	// It also receives channels for transmitting elevator status, order arrays, and new orders.
	// The function updates the IsActive status of existing elevators based on the updated peers.
	// If a new peer is detected, it sends the elevator status to the elevatorStatusTx channel.
	// If a lost peer is detected, it sets the IsActive status to false and redistributes hall orders.
	// If a new peer is detected and the current elevator is the master, it sends the order arrays to the new peer.
	// Finally, it calls the DetermineMaster function to determine the master elevator.

	fmt.Println("Function: HandlePeersUpdate")

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)

	var NewPeersMessage utils.NewPeersMessage

	MasterID, _ := strconv.Atoi(p.Peers[0])

	if MasterID == thisElevator.ID {
		thisElevator.IsMaster = true
	} else {
		thisElevator.IsMaster = false
	
	}

	utils.MasterElevatorID = MasterID



	for i, _ := range p.Peers {

		peerID, _ := strconv.Atoi(p.Peers[i])

		found := false

		for j, _ := range utils.Elevators {

			if utils.Elevators[j].ID == peerID {

				found = true
				continue

			}

		}

		if !found {

			go func() {

			channels.ElevatorStatusTx <- utils.ElevatorStatus{
				Type:         "ElevatorStatus",
				FromElevator: *thisElevator}
			
			}()

		}
	}

	if p.New != "" {

		newElevatorID, _ := strconv.Atoi(p.New)
		NewPeersMessage.NewPeers = append(NewPeersMessage.NewPeers, newElevatorID)
	}

	if p.Lost != nil {

		for i, _ := range p.Lost {

			lostElevatorID, _ := strconv.Atoi(p.Lost[i])
			NewPeersMessage.LostPeers = append(NewPeersMessage.LostPeers, lostElevatorID)

		}
	}

	go func() {

		if len(NewPeersMessage.NewPeers) > 0 || len(NewPeersMessage.LostPeers) > 0 {

			channels.PeersOnlineCh <- NewPeersMessage

		}

	}()

}

func HandleNewOrder(newOrder utils.Order, fromElevatorID int, toElevatorID int, thisElevator *utils.Elevator, channels *utils.Channels) {

	// HandleNewOrder handles a new order received by an elevator.
	// It updates the global order system, updates the acknowledgment structure,
	// and delegates the order to the appropriate elevator if necessary.
	// Parameters:
	//   - newOrder: The new order to be handled.
	//   - fromElevator: The elevator from which the order is received.
	//   - toElevatorID: The ID of the elevator to which the order is delegated.
	//   - orderCompleteTx: The channel to send the order completion message.
	//   - newOrderTx: The channel to send the new order message.
	//   - e: The current elevator.

	fmt.Println("Function: HandleNewOrder")

	if toElevatorID == utils.NotDefined && fromElevatorID != thisElevator.ID {

		fmt.Println("Update global order system. Order update for all.")
		// Update global order system

		orders := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		channels.GlobalUpdateCh <- orders

	}

	if thisElevator.IsMaster && toElevatorID == thisElevator.ID {

		// Update global order system locally
		// Find the best elevator for the order
		// Send the order to the best elevator ( if hall order )

		if newOrder.Button == utils.Cab {

			orders.UpdateLocalOrderSystem(newOrder, thisElevator)

			orders := utils.GlobalOrderUpdate{
				Orders:         []utils.Order{newOrder},
				FromElevatorID: fromElevatorID,
				IsComplete:     false,
				IsNew:          true}

			channels.GlobalUpdateCh <- orders

		} else {

			fmt.Println("I am master. I got a new order to delegate")

			ack := utils.OrderConfirmed{
				Type:           "OrderConfirmed",
				Confirmed:      true,
				FromElevatorID: thisElevator.ID}

			channels.AckTx <- ack

			bestElevator := orders.ChooseElevator(newOrder)
			fmt.Println("The best elevator for this order is", bestElevator.ID)

			if bestElevator.ID == thisElevator.ID {

				orders.UpdateLocalOrderSystem(newOrder, thisElevator)

				orderUpdate := utils.GlobalOrderUpdate{
					Orders:         []utils.Order{newOrder},
					FromElevatorID: fromElevatorID,
					IsComplete:     false,
					IsNew:          true}

				channels.GlobalUpdateCh <- orderUpdate

				amountOfOrders := orders.CheckAmountOfActiveOrders(thisElevator)

				if amountOfOrders > 0 {

					if thisElevator.CurrentState == utils.Still {
			
			
						utils.BestOrder = orders.ChooseBestOrder(thisElevator) // Choose the best order
			
			
						if utils.BestOrder.Floor == thisElevator.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
			
							HandleElevatorAtFloor(utils.BestOrder.Floor, channels, thisElevator) // Handle the elevator at the floor
			
						} else { 
			
							fmt.Println("The best order is ", utils.BestOrder)
							DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order
			
						}
			
					} else if thisElevator.CurrentState == utils.Moving {
			
						utils.BestOrder = orders.ChooseBestOrder(thisElevator) // Choose the best order
						fmt.Println("The best order is ", utils.BestOrder)
						DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order	
					}
			
				} else {
			
					thisElevator.StopElevator()
			
				}

			} else {

				newOrder := utils.MessageNewOrder{
					Type:           "MessageNewOrder",
					NewOrder:       newOrder,
					ToElevatorID:   bestElevator.ID, // Use the correct field name as defined in your ElevatorStatus struct
					FromElevatorID: thisElevator.ID}

				channels.NewOrderTx <- newOrder
			}
		}

	} else if !thisElevator.IsMaster && toElevatorID == thisElevator.ID {

		fmt.Println("New order received: ", newOrder, "from master elevator.")

		orders.UpdateLocalOrderSystem(newOrder, thisElevator)

		orderUpdate := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		channels.GlobalUpdateCh <- orderUpdate

		amountOfOrders := orders.CheckAmountOfActiveOrders(thisElevator)

		if amountOfOrders > 0 {

			if thisElevator.CurrentState == utils.Still {
	
	
				utils.BestOrder = orders.ChooseBestOrder(thisElevator) // Choose the best order
	
	
				if utils.BestOrder.Floor == thisElevator.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
	
					HandleElevatorAtFloor(utils.BestOrder.Floor, channels, thisElevator) // Handle the elevator at the floor
	
				} else { 
	
					fmt.Println("The best order is ", utils.BestOrder)
					DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order
	
				}
	
			} else if thisElevator.CurrentState == utils.Moving {
	
				utils.BestOrder = orders.ChooseBestOrder(thisElevator) // Choose the best order
				fmt.Println("The best order is ", utils.BestOrder)
				DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order	
			}
	
		} else {
	
			thisElevator.StopElevator()
	
		}

	} else if !thisElevator.IsMaster && toElevatorID != thisElevator.ID && fromElevatorID != thisElevator.ID {

		orders := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		channels.GlobalUpdateCh <- orders

	}
}

func WaitForAck(ackCh chan utils.OrderConfirmed, timeout time.Duration,
	newOrder utils.Order, thisElevator *utils.Elevator) {

	fmt.Println("Function: WaitForAck")

	for {
		select {

		case ack := <-ackCh:

			if ack.Confirmed && ack.FromElevatorID == utils.MasterElevatorID {

				if newOrder.Button == utils.Cab {

					utils.SlaveOrderWatcher.CabOrderArray[thisElevator.ID][newOrder.Floor].Confirmed = true

				} else {

					utils.SlaveOrderWatcher.HallOrderArray[newOrder.Button][newOrder.Floor].Confirmed = true
				}

				fmt.Println("Order confirmed by master.")

				return
			}

		case <-time.After(timeout):

			fmt.Println("Timeout")

		}

	}
}
