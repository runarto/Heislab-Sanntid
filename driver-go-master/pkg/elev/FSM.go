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

func NullButtons() {

	// NullButtons turns off all elevator buttons and the stop lamp.

	elevio.SetStopLamp(false)
	for f := 0; f < utils.NumFloors; f++ {
		for b := 0; b < utils.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func InitializeElevator(e *utils.Elevator) {

	NullButtons()

	fmt.Println("Function: InitializeElevator")

	floor := elevio.GetFloor()
	direction := utils.Up // 1 for up, -1 for down
	maxTime := 2000       // maximum time to move in one direction

	// Start moving up
	e.GoUp()

	// Start a timer
	startTime := time.Now()

	for floor == utils.NotDefined {
		floor = elevio.GetFloor()

		// If we've been moving in one direction for more than maxTime milliseconds
		// without finding a floor, switch direction
		if time.Since(startTime).Milliseconds() > int64(maxTime) {
			if direction == 1 {
				e.GoDown()
				direction = -1
			} else {
				e.GoUp()
				direction = 1
			}

			// Reset the timer
			startTime = time.Now()
		}
	}

	// Stop the elevator when a floor is found
	e.StopElevator()
}
func FloorLights(floor int, e *utils.Elevator) {

	// FloorLights sets the floor indicator light and updates the current floor of the elevator.
	// It takes the floor number and a pointer to the elevator as input.
	// The floor number should be between 0 and NumFloors-1.

	if floor >= 0 && floor <= utils.NumFloors-1 {
		elevio.SetFloorIndicator(floor)
		e.CurrentFloor = floor
	}
}

func HandleOrdersAtFloor(floor int, OrderCompleteTx chan utils.MessageOrderComplete,
	e *utils.Elevator) bool {

	// HandleOrdersAtFloor handles the orders at a specific floor.
	// It takes the current floor, a channel for order completion messages,
	// and a pointer to the elevator as input.
	// It checks for active orders at the floor and takes appropriate actions based on the elevator's current direction.
	// It updates the local and global order systems, as well as the acknowledgement structure.
	// Finally, it sends an order completion message through the provided channel.
	// Returns true if there are active orders at the floor, false otherwise.

	fmt.Println("Function: HandleOrdersAtFloor")
	// Update the current floor
	var ordersDone []utils.Order // Number of orders done

	for button := 0; button < utils.NumButtons; button++ {
		if e.LocalOrderArray[button][floor] == utils.True { // If there is an active order at the floor

			if floor == utils.BestOrder.Floor {

				if e.CurrentDirection == utils.Up && button == utils.HallUp || e.CurrentDirection == utils.Stopped && button == utils.HallUp {

					fmt.Println("HandleOrdersAtFloor: HallUp order at floor: ", floor)

					Order := utils.Order{
						Floor:  floor,
						Button: utils.HallUp}

					ordersDone = append(ordersDone, Order)
					// HallUp order, and the elevator is going up (take order)
					continue
				}

				if (e.CurrentDirection == utils.Up && button == utils.HallDown) && (e.LocalOrderArray[utils.HallUp][floor] == utils.False) {

					check := orders.CheckHallOrdersAbove(floor, e)

					if check.Button == elevio.ButtonType(button) && check.Floor == floor { // There are no orders above the current floor

						fmt.Println("HandleOrdersAtFloor: HallDown order at floor", floor, ", with no orders up.")

						Order := utils.Order{
							Floor:  floor,
							Button: utils.HallDown}

						ordersDone = append(ordersDone, Order) // Update the local order array
						// HallDown order, and the elevator is going up (take order)
						continue
					}
				}

				if e.CurrentDirection == utils.Down && button == utils.HallDown || e.CurrentDirection == utils.Stopped && button == utils.HallDown {

					fmt.Println("HandleOrdersAtFloor: HallDown order at floor: ", floor)

					Order := utils.Order{
						Floor:  floor,
						Button: utils.HallDown}
					ordersDone = append(ordersDone, Order) // Update the local order array
					// HallDown order, and the elevator is going down (take order)
					continue
				}

				if (e.CurrentDirection == utils.Down && button == utils.HallUp) && (e.LocalOrderArray[utils.HallDown][floor] == utils.False) {

					check := orders.CheckHallOrdersBelow(floor, e)

					if check.Button == elevio.ButtonType(button) && check.Floor == floor { // There are no orders below the current floor

						fmt.Println("HandleOrdersAtFloor: HallUp order at floor", floor, ", with no orders down.")

						Order := utils.Order{
							Floor:  floor,
							Button: utils.HallUp}

						ordersDone = append(ordersDone, Order) // Update the local order array
						// HallUp order, and the elevator is going down (take order)
						continue
					}
				}

			}

			if button == utils.Cab {
				fmt.Println("Cab order at floor: ", floor)
				Order := utils.Order{
					Floor:  floor,
					Button: utils.Cab}

				ordersDone = append(ordersDone, Order) // Update the local order array
				// Cab order (take order)
				continue
			}

		}
	}

	if len(ordersDone) > 0 {
		for i := 0; i < len(ordersDone); i++ {

			fmt.Println("Order done: ", ordersDone[i])
			orders.UpdateLocalOrderSystem(ordersDone[i], e)         // Update the local order array
			orders.UpdateGlobalOrderSystem(ordersDone[i], e, false) // Update the global order system
			OrderCompleted(ordersDone[i], e)                        // Update the ackStruct

		}

		OrderCompleteTx <- utils.MessageOrderComplete{Type: "OrderComplete",
			Orders:         ordersDone,
			FromElevator:   *e,
			FromElevatorID: e.ID}

		fmt.Println("Function HandleOrdersAtFloor: true")

		return true // There are active orders at the floor

	} else {

		fmt.Println("Function HandleOrdersAtFloor: false")
		return false // There are no active orders at the floor
	}

}

func HandleElevatorAtFloor(floor int, OrderCompleteTx chan utils.MessageOrderComplete, e *utils.Elevator) {

	// HandleElevatorAtFloor handles the elevator's behavior when it reaches a specific floor.
	// It handles the orders at the floor, stops the elevator, opens the door, waits for a second,
	// closes the door, prints the order system, checks the amount of active orders, chooses the best order,
	// and sets the elevator in the direction of the best order if there are active orders.
	// If there are no orders, it stops the elevator and sets the state to still.

	fmt.Println("Function: HandleElevatorAtFloor")

	if HandleOrdersAtFloor(floor, OrderCompleteTx, e) && elevio.GetFloor() != utils.NotDefined { // If true, orders have been handled at the floor

		e.StopElevator()                    // Stop the elevator
		e.SetDoorState(utils.Open)          // utils.Open the door
		time.Sleep(1500 * time.Millisecond) // Wait for a second
		e.SetDoorState(utils.Close)         // utils.Close the door

		fmt.Println("Order system: ")
		orders.PrintLocalOrderSystem(e)

		amountOfOrders := orders.CheckAmountOfActiveOrders(e) // Check the amount of active orders

		fmt.Println("Amount of active orders: ", amountOfOrders)

		if amountOfOrders > 0 {

			utils.BestOrder = orders.ChooseBestOrder(e) // Choose the best order

			fmt.Println("Best order: ", utils.BestOrder)

			DoOrder(utils.BestOrder, OrderCompleteTx, e) // Set elevator in direction of best order

		} else {

			fmt.Println("No orders, stopped elevator.")
			e.SetState(utils.Still) // If no orders, set the state to still
			e.GeneralDirection = utils.Stopped
		}
	}
}

func HandleButtonEvent(newOrderTx chan utils.MessageNewOrder, orderCompleteTx chan utils.MessageOrderComplete, newOrder utils.Order, e *utils.Elevator) {

	// HandleButtonEvent handles a button event by processing the new order and updating the global order system.
	// It takes in the following parameters:
	// - newOrderTx: a channel for sending a new order message
	// - orderCompleteTx: a channel for sending an order complete message
	// - newOrder: the new order to be processed
	// - e: a pointer to the elevator object

	fmt.Println("Function: HandleButtonEvent")

	if !orders.CheckIfGlobalOrderIsActive(newOrder, e) { // Check if the order is already active

		orders.UpdateGlobalOrderSystem(newOrder, e, true) // Update the global order system

		button := newOrder.Button
		//floor := newOrder.Floor

		if button == utils.Cab {

			fmt.Println("Cab order")

			orders.UpdateGlobalOrderSystem(newOrder, e, true)
			OrderActive(newOrder, e, time.Now())

			newOrderTx <- utils.MessageNewOrder{
				Type:         "MessageNewOrder",
				NewOrder:     newOrder,
				FromElevator: *e,
				ToElevatorID: utils.NotDefined}

			if orders.CheckIfOrderIsActive(newOrder, e) { // Check if the order is active
				if utils.BestOrder.Floor == e.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
					HandleElevatorAtFloor(utils.BestOrder.Floor, orderCompleteTx, e) // Handle the elevator at the floor
				} else {
					fmt.Println("Best order is", utils.BestOrder)
					DoOrder(utils.BestOrder, orderCompleteTx, e) // Move the elevator to the best order
				}

			} else {

				fmt.Println("Going into ProcessElevatorOrders")
				ProcessElevatorOrders(newOrder, orderCompleteTx, e)
			}
		} else {

			fmt.Println("Hall order")

			if e.IsMaster {

				fmt.Println("This is the master")
				// Handle order locally (remember lights)
				bestElevator := orders.ChooseElevator(newOrder)

				if bestElevator.ID == e.ID {

					fmt.Println("Handling locally")
					orders.UpdateGlobalOrderSystem(newOrder, e, true)
					OrderActive(newOrder, e, time.Now())
					ProcessElevatorOrders(newOrder, orderCompleteTx, e)

					newOrder := utils.MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     newOrder,
						FromElevator: *e, // Use the correct field name as defined in your ElevatorStatus struct
						ToElevatorID: utils.NotDefined}

					fmt.Println("Sending order")
					newOrderTx <- newOrder

				} else {

					fmt.Println("Sending order")

					

					orders.UpdateGlobalOrderSystem(newOrder, bestElevator, true)
					OrderActive(newOrder, bestElevator, time.Now())


					newOrder := utils.MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     newOrder,
						FromElevator: *e, // Use the correct field name as defined in your ElevatorStatus struct
						ToElevatorID: bestElevator.ID}

					newOrderTx <- newOrder

					if orders.CheckAmountOfActiveOrders(e) > 0 {
						DoOrder(utils.BestOrder, orderCompleteTx, e)
					}

				}

			} else {

				// Send order to master

				newOrderTx <- utils.MessageNewOrder{
					Type:         "MessageNewOrder",
					NewOrder:     newOrder,
					FromElevator: *e,
					ToElevatorID: utils.MasterElevatorID}

				if orders.CheckAmountOfActiveOrders(e) > 0 {

					DoOrder(utils.BestOrder, orderCompleteTx, e)

				} else {

					e.StopElevator()

				}

			}
		}
	}
}

func ProcessElevatorOrders(newOrder utils.Order, orderCompleteTx chan utils.MessageOrderComplete, e *utils.Elevator) {

	// ProcessElevatorOrders processes the elevator orders by updating the local order system,
	// checking the amount of active orders, choosing the best order, and handling the elevator
	// at the floor or moving it to the best order.
	//
	// Parameters:
	// - newOrder: The new order to be processed.
	// - orderCompleteTx: The channel for sending order complete messages.
	// - e: The elevator object.
	//
	// Returns: None.

	fmt.Println("Function: ProcessElevatorOrders")

	orders.UpdateLocalOrderSystem(newOrder, e)

	orders.PrintLocalOrderSystem(e)

	amountOfOrders := orders.CheckAmountOfActiveOrders(e)

	fmt.Println("Amount of active orders: ", amountOfOrders)

	if amountOfOrders > 0 {

		utils.BestOrder = orders.ChooseBestOrder(e) // Choose the best order
		fmt.Println("Best order: ", utils.BestOrder)

		if utils.BestOrder.Floor == e.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
			HandleElevatorAtFloor(utils.BestOrder.Floor, orderCompleteTx, e) // Handle the elevator at the floor
		} else {
			DoOrder(utils.BestOrder, orderCompleteTx, e) // Move the elevator to the best order
		}
	} else {
		e.StopElevator()
	}
}

func HandleNewOrder(newOrder utils.Order, fromElevator *utils.Elevator, toElevatorID int,
	orderCompleteTx chan utils.MessageOrderComplete, newOrderTx chan utils.MessageNewOrder, e *utils.Elevator) {

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

	if toElevatorID == utils.NotDefined && fromElevator.ID != e.ID {

		fmt.Println("Update global order system. Order update for all.")
		// Update global order system
		orders.UpdateGlobalOrderSystem(newOrder, fromElevator, true)
		OrderActive(newOrder, fromElevator, time.Now())
		UpdateElevatorsOnNetwork(fromElevator)

	}

	if e.IsMaster && toElevatorID == e.ID {

		fmt.Println("I am master. I got a new order to delegate")

		// Update global order system locally
		// Find the best elevator for the order
		// Send the order to the best elevator ( if hall order )
		UpdateElevatorsOnNetwork(fromElevator)

		if newOrder.Button == utils.Cab {

			orders.UpdateGlobalOrderSystem(newOrder, e, true)
			OrderActive(newOrder, e, time.Now())

			ProcessElevatorOrders(newOrder, orderCompleteTx, e)

		} else {

			bestElevator := orders.ChooseElevator(newOrder)
			fmt.Println("The best elevator for this order is", bestElevator.ID)

			if bestElevator.ID == e.ID {

				orders.UpdateGlobalOrderSystem(newOrder, e, true)
				OrderActive(newOrder, e, time.Now())
				ProcessElevatorOrders(newOrder, orderCompleteTx, e)

			} else {

				newOrder := utils.MessageNewOrder{
					Type:         "MessageNewOrder",
					NewOrder:     newOrder,
					FromElevator: *e, // Use the correct field name as defined in your ElevatorStatus struct
					ToElevatorID: bestElevator.ID}

				newOrderTx <- newOrder
			}
		}

	} else if !e.IsMaster && toElevatorID == e.ID {

		fmt.Println("New order received: ", newOrder, "from master elevator.")

		UpdateElevatorsOnNetwork(fromElevator)

		orders.UpdateGlobalOrderSystem(newOrder, e, true)
		OrderActive(newOrder, e, time.Now())

		ProcessElevatorOrders(newOrder, orderCompleteTx, e)

	} else if !e.IsMaster && toElevatorID != e.ID && fromElevator.ID != e.ID {

		UpdateElevatorsOnNetwork(fromElevator)
	}
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
	
	var offlineElevators []utils.Elevator
	for i, _ := range utils.Elevators {
		for _, peer := range p.Lost {
			peerID, _ := strconv.Atoi(peer)
			if utils.Elevators[i].ID == peerID {
				utils.Elevators[i].IsActive = false
				offlineElevators = append(offlineElevators, utils.Elevators[i])
			}
		}
	}

	for i, _ := range offlineElevators {
		if offlineElevators[i].ID == e.ID {
			utils.Elevators[i].IsActive = true
		} else {
			orders.RedistributeHallOrders(&offlineElevators[i], newOrderTx, e)
		
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

func DoOrder(order utils.Order, OrderCompleteTx chan utils.MessageOrderComplete,
	e *utils.Elevator) {

	// DoOrder executes the given order for the elevator.
	// It moves the elevator up or down to reach the order's floor.
	// If the order's floor is the same as the current floor of the elevator,
	// it waits for a possible order completion.
	// The function takes the order to be executed, a channel to send order completion messages,
	// and a pointer to the elevator object.

	// Do the order
	if order.Floor > e.CurrentFloor {

		e.GoUp()

	} else if order.Floor < e.CurrentFloor {

		e.GoDown()

	} else {
		// The order is at the current floor
		go WaitForPossibleOrder(order, OrderCompleteTx, e)

	}

}

func WaitForPossibleOrder(order utils.Order, OrderCompleteTx chan utils.MessageOrderComplete,
	e *utils.Elevator) {

	// WaitForPossibleOrder waits for a possible order to be completed.
	// It takes an order, a channel for sending order completion messages, and an elevator as parameters.
	// If the elevator is at the same floor as the order, it stops the elevator and handles the order completion.
	// It also inserts a delay in case cab-orders come in.

	time.Sleep(3 * time.Second) // Insert delay in case cab-orders come in.
	if e.CurrentFloor == order.Floor {
		e.StopElevator()
		HandleElevatorAtFloor(order.Floor, OrderCompleteTx, e)
	}

}
