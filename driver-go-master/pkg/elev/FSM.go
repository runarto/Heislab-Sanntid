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

func NullButtons() { // Turns off all buttons
	elevio.SetStopLamp(false)
	for f := 0; f < utils.NumFloors; f++ {
		for b := 0; b < utils.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func InitElevator(e *utils.Elevator) {
	NullButtons()
	e.SetDoorState(utils.Close) // utils.Close the door

	for floor := elevio.GetFloor(); floor != 0; floor = elevio.GetFloor() {
		if floor > 0 || floor == -1 {
			e.GoDown()
		}
		time.Sleep(100 * time.Millisecond)
	}
	e.StopElevator()
	e.CurrentFloor = elevio.GetFloor()
	fmt.Println("Elevator is ready for use")

}

func FloorLights(floor int, e *utils.Elevator) {
	if floor >= 0 && floor <= 3 {
		elevio.SetFloorIndicator(floor)
		e.CurrentFloor = floor
	}
}

func HandleOrdersAtFloor(floor int, OrderCompleteTx chan utils.MessageOrderComplete,
	e *utils.Elevator) bool {

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
			orders.UpdateOrderSystem(ordersDone[i], e)              // Update the local order array
			orders.UpdateGlobalOrderSystem(ordersDone[i], e, false) // Update the global order system
			OrderCompleted(ordersDone[i], e)                        // Update the ackStruct

		}

		OrderCompleteTx <- utils.MessageOrderComplete{Type: "OrderComplete",
			Orders:         ordersDone,
			E:              *e,
			FromElevatorID: e.ID}

		fmt.Println("Function HandleOrdersAtFloor: true")

		return true // There are active orders at the floor

	} else {

		fmt.Println("Function HandleOrdersAtFloor: false")
		return false // There are no active orders at the floor
	}

}

func HandleElevatorAtFloor(floor int, OrderCompleteTx chan utils.MessageOrderComplete, e *utils.Elevator) {
	fmt.Println("Function: HandleElevatorAtFloor")

	if HandleOrdersAtFloor(floor, OrderCompleteTx, e) { // If true, orders have been handled at the floor

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
	fmt.Println("Function: HandleButtonEvent")

	if !orders.CheckIfGlobalOrderIsActive(newOrder, e) { // Check if the order is already active

		orders.UpdateGlobalOrderSystem(newOrder, e, true) // Update the global order system

		button := newOrder.Button
		//floor := newOrder.Floor

		if button == utils.Cab {

			fmt.Println("Cab order")

			newOrderTx <- utils.MessageNewOrder{Type: "MessageNewOrder", NewOrder: newOrder, E: *e, ToElevatorID: utils.NotDefined}

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

				OrderActive(newOrder, e)


				fmt.Println("This is the master")
				// Handle order locally (remember lights)
				bestElevator := orders.ChooseElevator(newOrder)

				if bestElevator.ID == e.ID {

					fmt.Println("Handling locally")

					ProcessElevatorOrders(newOrder, orderCompleteTx, e)

					newOrder := utils.MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     newOrder,
						E:            *e, // Use the correct field name as defined in your ElevatorStatus struct
						ToElevatorID: utils.NotDefined}

					fmt.Println("Sending order")
					newOrderTx <- newOrder

				} else {
					fmt.Println("Sending order")

					newOrder := utils.MessageNewOrder{
						Type:         "MessageNewOrder",
						NewOrder:     newOrder,
						E:            *e, // Use the correct field name as defined in your ElevatorStatus struct
						ToElevatorID: bestElevator.ID}

					newOrderTx <- newOrder

					if orders.CheckAmountOfActiveOrders(e) > 0 {
						DoOrder(utils.BestOrder, orderCompleteTx, e)
					}

				}

			} else {

				// Set lights.
				fmt.Println("Sending order to master.")

				newOrderTx <- utils.MessageNewOrder{Type: "MessageNewOrder", NewOrder: newOrder, E: *e, ToElevatorID: utils.MasterElevatorID}

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

	fmt.Println("Function: ProcessElevatorOrders")

	orders.UpdateOrderSystem(newOrder, e)

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

	fmt.Println("Function: HandleNewOrder")

	 // Check if the order is already active

	orders.UpdateGlobalOrderSystem(newOrder, e, true) // Update the global order system
	OrderActive(newOrder, e)                          // Update the ackStruct

	if toElevatorID == utils.NotDefined && fromElevator.ID != e.ID {

		fmt.Println("Update global order system. Order update for all.")
		// Update global order system
		UpdateElevatorsOnNetwork(fromElevator)

	}

	if e.IsMaster && toElevatorID == e.ID {

		OrderActive(newOrder, e)

		fmt.Println("I am master. I got a new order to delegate")

		// Update global order system locally
		// Find the best elevator for the order
		// Send the order to the best elevator ( if hall order )
		UpdateElevatorsOnNetwork(fromElevator)
		bestElevator := orders.ChooseElevator(newOrder)
		fmt.Println("The best elevator for this order is", bestElevator.ID)

		if bestElevator.ID == e.ID {

			ProcessElevatorOrders(newOrder, orderCompleteTx, e)

		} else {

			newOrder := utils.MessageNewOrder{
				Type:         "MessageNewOrder",
				NewOrder:     newOrder,
				E:            *e, // Use the correct field name as defined in your ElevatorStatus struct
				ToElevatorID: bestElevator.ID}

			newOrderTx <- newOrder
		}

	} else if !e.IsMaster && toElevatorID == e.ID {

		fmt.Println("New order received: ", newOrder, "from master elevator.")

		UpdateElevatorsOnNetwork(fromElevator)

		ProcessElevatorOrders(newOrder, orderCompleteTx, e)

	} else if !e.IsMaster && toElevatorID != e.ID && fromElevator.ID != e.ID {
		// Update global order system locally
		// Remember to set lights (if hall order)
		UpdateElevatorsOnNetwork(fromElevator)
	}
}

func HandlePeersUpdate(p peers.PeerUpdate, elevatorStatusTx chan utils.ElevatorStatus,
	orderArraysTx chan utils.MessageOrderArrays, newOrderTx chan utils.MessageNewOrder, e *utils.Elevator) {

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
				Type: "ElevatorStatus",
				E:    *e,}
			}

			time.Sleep(1 * time.Second)
		}

	for i, _ := range utils.Elevators {
		for _, peer := range p.Lost {
			peerID, _ := strconv.Atoi(peer)
			if utils.Elevators[i].ID == peerID {
				utils.Elevators[i].IsActive = false
				orders.RedistributeHallOrders(&utils.Elevators[i], newOrderTx, e)
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

			MessageOrderArrays := utils.MessageOrderArrays{
				Type:            "MessageOrderArrays",
				GlobalOrders:    utils.GlobalOrders,
				LocalOrderArray: LocalOrders,
				ToElevatorID:    peerID}

			orderArraysTx <- MessageOrderArrays

			time.Sleep(1 * time.Second)
		}
	}

	DetermineMaster(e) // Determine the master elevator
}

func DoOrder(order utils.Order, OrderCompleteTx chan utils.MessageOrderComplete,
	e *utils.Elevator) {

	fmt.Println("Function: DoOrder")
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

	time.Sleep(3 * time.Second) // Insert delay in case cab-orders come in.
	if e.CurrentFloor == order.Floor {
		e.StopElevator()
		HandleElevatorAtFloor(order.Floor, OrderCompleteTx, e)
	}

}
