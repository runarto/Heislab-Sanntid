package elev

import (
	"fmt"
	"time"

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
	e *utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate) bool {

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

		orders := utils.GlobalOrderUpdate{
			Orders:         ordersDone,
			FromElevatorID: e.ID,
			IsComplete:     true,
			IsNew:          false}

		GlobalUpdateCh <- orders

		ordersComplete := utils.MessageOrderComplete{
			Type:           "MessageOrderComplete",
			Orders:         ordersDone,
			ToElevatorID:   utils.NotDefined,
			FromElevatorID: e.ID}

		OrderCompleteTx <- ordersComplete

		return true

	} else {

		fmt.Println("Function HandleOrdersAtFloor: false")
		return false // There are no active orders at the floor
	}

}

func HandleElevatorAtFloor(floor int, OrderCompleteTx chan utils.MessageOrderComplete, e *utils.Elevator, bestOrderCh chan utils.Order) {

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

			BestOrder := orders.ChooseBestOrder(e) // Choose the best order

			fmt.Println("Best order: ", utils.BestOrder)

			bestOrderCh <- BestOrder

		} else {

			fmt.Println("No orders, stopped elevator.")
			e.SetState(utils.Still) // If no orders, set the state to still
			e.GeneralDirection = utils.Stopped
		}
	}
}

func HandleButtonEvent(newOrderTx chan utils.MessageNewOrder, orderCompleteTx chan utils.MessageOrderComplete, newOrder utils.Order, e *utils.Elevator, bestOrderCh chan utils.Order) {

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
				ProcessElevatorOrders(newOrder, orderCompleteTx, e, bestOrderCh)
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
					ProcessElevatorOrders(newOrder, orderCompleteTx, e, bestOrderCh)

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

func ProcessElevatorOrders(newOrder utils.Order, orderCompleteTx chan utils.MessageOrderComplete, e *utils.Elevator, bestOrderCh chan utils.Order) {

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

		BestOrder := orders.ChooseBestOrder(e) // Choose the best order
		fmt.Println("Best order: ", utils.BestOrder)

		if utils.BestOrder.Floor == e.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
			HandleElevatorAtFloor(BestOrder.Floor, orderCompleteTx, e, bestOrderCh) // Handle the elevator at the floor
		} else {

			bestOrderCh <- BestOrder
		}
	} else {
		e.StopElevator()
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
