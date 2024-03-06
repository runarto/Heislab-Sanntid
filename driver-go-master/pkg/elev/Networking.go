package elev

import (
	"fmt"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func BroadcastElevatorStatus(e *utils.Elevator, statusTx chan utils.ElevatorStatus) {

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
				FromElevator: *e,
			}

			statusTx <- elevatorStatusMessage
		}
	}
}

func BroadcastAckMatrix(e *utils.Elevator, ackTx chan utils.AckMatrix) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.

	time.Sleep(5 * time.Second)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		if len(utils.Elevators) > 1 && e.IsMaster {

			ackStruct := utils.AckMatrix{
				Type:           "AckMatrix",
				OrderWatcher:   utils.OrderWatcher, // Use the correct field name as defined in your ElevatorStatus struct
				FromElevatorID: e.ID,
			}

			ackTx <- ackStruct // Broadcast the current status
		}
	}
}

func UpdateElevatorsOnNetwork(e *utils.Elevator) {

	// UpdateElevatorsOnNetwork updates the elevator information in the ActiveElevators array.
	// It takes a pointer to an Elevator struct as input and updates the elevator with the same ID in the array.
	// If the elevator does not exist in the array, it adds the elevator to the array.
	// Finally, it prints the local order array for the elevator.

	elevatorID := e.ID      // The ID of the elevator
	elevatorExists := false // Flag to check if the elevator exists in the ActiveElevators array
	e.IsActive = true

	for i, _ := range utils.Elevators {
		if utils.Elevators[i].ID == elevatorID {
			utils.Elevators[i] = *e // Update the elevator
			elevatorExists = true   // Set the elevatorExists flag to true
			break
		}
	}

	if !elevatorExists { // If the elevator does not exist in the ActiveElevators array
		utils.Elevators = append(utils.Elevators, *e) // Add the elevator to the ActiveElevators array
	}

	fmt.Println("Local order array for elevator", e.ID)
	orders.PrintLocalOrderSystem(e)

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
		FromElevatorID: orderComplete.FromElevatorID}

	GlobalUpdateCh <- orders

}

func HandlePeersUpdate(p peers.PeerUpdate, elevatorStatusTx chan utils.ElevatorStatus,
	orderArraysTx chan utils.MessageOrderArrays, newOrderTx chan utils.MessageNewOrder, e *utils.Elevator) {

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

	for _, peer := range p.Peers {
		found := false
		peerID, _ := strconv.Atoi(peer)
		for i, _ := range utils.Elevators {
			if utils.Elevators[i].ID == peerID {
				found = true
				utils.Elevators[i].IsActive = true
			}
		}

		if !found {

			elevatorStatusTx <- utils.ElevatorStatus{
				Type:         "ElevatorStatus",
				FromElevator: *e}

		}

	}

	for i, _ := range utils.Elevators {
		for _, peer := range p.Lost {
			peerID, _ := strconv.Atoi(peer)
			if utils.Elevators[i].ID == peerID && utils.Elevators[i].ID != e.ID {
				utils.Elevators[i].IsActive = false

			}
		}
	}

	if p.New != "" && e.IsMaster {
		peerID, _ := strconv.Atoi(p.New)

		if peerID != e.ID {
			fmt.Println("New peer: ", peerID)

			var LocalOrders [utils.NumButtons][utils.NumFloors]int

			for i, _ := range utils.Elevators {
				if utils.Elevators[i].ID == peerID {
					LocalOrders = utils.Elevators[i].LocalOrderArray
					break
				}
			}
			fmt.Println("Local order array found.")

			MessageOrderArrays := utils.MessageOrderArrays{
				Type:            "MessageOrderArrays",
				GlobalOrders:    utils.GlobalOrders,
				LocalOrderArray: LocalOrders,
				ToElevatorID:    peerID}

			orderArraysTx <- MessageOrderArrays

		}
	}

}

func HandleNewOrder(newOrder utils.Order, fromElevatorID int, toElevatorID int,
	orderCompleteTx chan utils.MessageOrderComplete, newOrderTx chan utils.MessageNewOrder,
	e *utils.Elevator, button chan elevio.ButtonEvent, GlobalOrderUpdateCh chan utils.GlobalOrderUpdate) {

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

	if toElevatorID == utils.NotDefined && fromElevatorID != e.ID {

		fmt.Println("Update global order system. Order update for all.")
		// Update global order system

		orders := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		GlobalOrderUpdateCh <- orders

	}

	if e.IsMaster && toElevatorID == e.ID {

		fmt.Println("I am master. I got a new order to delegate")

		// Update global order system locally
		// Find the best elevator for the order
		// Send the order to the best elevator ( if hall order )

		if newOrder.Button == utils.Cab {

			order := elevio.ButtonEvent{
				Floor:  newOrder.Floor,
				Button: elevio.ButtonType(newOrder.Button)}

			orders := utils.GlobalOrderUpdate{
				Orders:         []utils.Order{newOrder},
				FromElevatorID: fromElevatorID,
				IsComplete:     false,
				IsNew:          true}

			GlobalOrderUpdateCh <- orders

			button <- order // Handle the order locally

		} else {

			bestElevator := orders.ChooseElevator(newOrder)
			fmt.Println("The best elevator for this order is", bestElevator.ID)

			if bestElevator.ID == e.ID {

				order := elevio.ButtonEvent{
					Floor:  newOrder.Floor,
					Button: elevio.ButtonType(newOrder.Button)}

				orders := utils.GlobalOrderUpdate{
					Orders:         []utils.Order{newOrder},
					FromElevatorID: fromElevatorID,
					IsComplete:     false,
					IsNew:          true}

				GlobalOrderUpdateCh <- orders

				button <- order // Handle the order locally

			} else {

				newOrder := utils.MessageNewOrder{
					Type:           "MessageNewOrder",
					NewOrder:       newOrder,
					ToElevatorID:   bestElevator.ID, // Use the correct field name as defined in your ElevatorStatus struct
					FromElevatorID: e.ID}

				newOrderTx <- newOrder
			}
		}

	} else if !e.IsMaster && toElevatorID == e.ID {

		fmt.Println("New order received: ", newOrder, "from master elevator.")

		order := elevio.ButtonEvent{
			Floor:  newOrder.Floor,
			Button: elevio.ButtonType(newOrder.Button)}

		button <- order // Handle the order locally

		orders := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		GlobalOrderUpdateCh <- orders

	} else if !e.IsMaster && toElevatorID != e.ID && fromElevatorID != e.ID {

		order := utils.Order{
			Floor:  newOrder.Floor,
			Button: elevio.ButtonType(newOrder.Button)}

		orders := utils.GlobalOrderUpdate{
			Orders:         []utils.Order{order},
			FromElevatorID: fromElevatorID,
			IsComplete:     false,
			IsNew:          true}

		GlobalOrderUpdateCh <- orders

	}
}

func HandleNewElevatorStatus(elevatorStatus utils.ElevatorStatus, e *utils.Elevator, GlobalOrderArrayUpdateCh chan utils.GlobalOrderUpdate) {

	if elevatorStatus.FromElevator.ID == e.ID {
		return
	}

	UpdateElevatorsOnNetwork(&elevatorStatus.FromElevator)

	var ActiveOrders []utils.Order

	for button := 0; button < utils.NumButtons; button++ {
		for floor := 0; floor < utils.NumFloors; floor++ {
			if elevatorStatus.FromElevator.LocalOrderArray[button][floor] == utils.True {

				order := utils.Order{
					Floor:  floor,
					Button: elevio.ButtonType(button)}

				ActiveOrders = append(ActiveOrders, order)

			}
		}
	}

	update := utils.GlobalOrderUpdate{
		Orders:       ActiveOrders,
		FromElevator: elevatorStatus.FromElevator}

	GlobalOrderArrayUpdateCh <- update

}
