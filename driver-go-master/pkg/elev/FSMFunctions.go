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

func InitializeElevator(thisElevator *utils.Elevator) {

	NullButtons()

	fmt.Println("Function: InitializeElevator")

	floor := elevio.GetFloor()
	direction := utils.Up // 1 for up, -1 for down
	maxTime := 2000       // maximum time to move in one direction

	// Start moving up
	thisElevator.GoUp()

	// Start a timer
	startTime := time.Now()

	for floor == utils.NotDefined {
		floor = elevio.GetFloor()

		// If we've been moving in one direction for more than maxTime milliseconds
		// without finding a floor, switch direction
		if time.Since(startTime).Milliseconds() > int64(maxTime) {
			if direction == 1 {
				thisElevator.GoDown()
				direction = -1
			} else {
				thisElevator.GoUp()
				direction = 1
			}

			// Reset the timer
			startTime = time.Now()
		}
	}

	// Stop the elevator when a floor is found
	thisElevator.StopElevator()
}

func FloorLights(floor int, thisElevator *utils.Elevator) {

	// FloorLights sets the floor indicator light and updates the current floor of the elevator.
	// It takes the floor number and a pointer to the elevator as input.
	// The floor number should be between 0 and NumFloors-1.

	if floor >= 0 && floor <= utils.NumFloors-1 {
		elevio.SetFloorIndicator(floor)
		thisElevator.CurrentFloor = floor
	}
}

func HandleOrdersAtFloor(floor int, channels *utils.Channels, thisElevator *utils.Elevator) bool {

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
		if thisElevator.LocalOrderArray[button][floor] == utils.True { // If there is an active order at the floor

			if floor == utils.BestOrder.Floor {

				if thisElevator.CurrentDirection == utils.Up && button == utils.HallUp || thisElevator.CurrentDirection == utils.Stopped && button == utils.HallUp {

					fmt.Println("HandleOrdersAtFloor: HallUp order at floor: ", floor)

					Order := utils.Order{
						Floor:  floor,
						Button: utils.HallUp}

					ordersDone = append(ordersDone, Order)
					// HallUp order, and the elevator is going up (take order)
					continue
				}

				if (thisElevator.CurrentDirection == utils.Up && button == utils.HallDown) && (thisElevator.LocalOrderArray[utils.HallUp][floor] == utils.False) {

					check := orders.CheckHallOrdersAbove(floor, thisElevator)

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

				if thisElevator.CurrentDirection == utils.Down && button == utils.HallDown || thisElevator.CurrentDirection == utils.Stopped && button == utils.HallDown {

					fmt.Println("HandleOrdersAtFloor: HallDown order at floor: ", floor)

					Order := utils.Order{
						Floor:  floor,
						Button: utils.HallDown}
					ordersDone = append(ordersDone, Order) // Update the local order array
					// HallDown order, and the elevator is going down (take order)
					continue
				}

				if (thisElevator.CurrentDirection == utils.Down && button == utils.HallUp) && (thisElevator.LocalOrderArray[utils.HallDown][floor] == utils.False) {

					check := orders.CheckHallOrdersBelow(floor, thisElevator)

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

		fmt.Println("Function HandleOrdersAtFloor: true")

		for i, _ := range ordersDone {
			orders.UpdateLocalOrderSystem(ordersDone[i], thisElevator)
		}

		go func() {

			orders := utils.GlobalOrderUpdate{
				Orders:         ordersDone,
				FromElevatorID: thisElevator.ID,
				IsComplete:     true,
				IsNew:          false}

			channels.GlobalUpdateCh <- orders
		}()

		go func() {

			channels.OrderWatcher <- utils.OrderWatcher{
				Orders:        ordersDone,
				ForElevatorID: thisElevator.ID,
				IsComplete:    true,
				IsNew:         false}
		}()

		go func() {
			ordersComplete := utils.MessageOrderComplete{
				Type:           "MessageOrderComplete",
				Orders:         ordersDone,
				ToElevatorID:   utils.NotDefined,
				FromElevatorID: thisElevator.ID}

			fmt.Println("Sending order complete message...")
			channels.OrderCompleteTx <- ordersComplete
			fmt.Println("Order complete message sent.")
		}()

		return true

	} else {

		fmt.Println("Function HandleOrdersAtFloor: false")
		return false // There are no active orders at the floor
	}

}

func HandleElevatorAtFloor(floor int, channels *utils.Channels, thisElevator *utils.Elevator) {

	// HandleElevatorAtFloor handles the elevator's behavior when it reaches a specific floor.
	// It handles the orders at the floor, stops the elevator, opens the door, waits for a second,
	// closes the door, prints the order system, checks the amount of active orders, chooses the best order,
	// and sets the elevator in the direction of the best order if there are active orders.
	// If there are no orders, it stops the elevator and sets the state to still.

	fmt.Println("Function: HandleElevatorAtFloor")

	if HandleOrdersAtFloor(floor, channels, thisElevator) { // If true, orders have been handled at the floor

		fmt.Println("Here")

		thisElevator.StopElevator()            // Stop the elevator
		thisElevator.SetDoorState(utils.Open)  // utils.Open the door
		time.Sleep(1500 * time.Millisecond)    // Wait for a second
		thisElevator.SetDoorState(utils.Close) // utils.Close the door

		fmt.Println("Order system: ")
		orders.PrintLocalOrderSystem(thisElevator)

		amountOfOrders := orders.CheckAmountOfActiveOrders(thisElevator) // Check the amount of active orders

		fmt.Println("Amount of active orders: ", amountOfOrders)

		if amountOfOrders > 0 {

			utils.BestOrder = orders.ChooseBestOrder(thisElevator) // Choose the best order

			fmt.Println("Best order: ", utils.BestOrder)

			if thisElevator.CurrentState == utils.Still {

				DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order

			}

		} else {

			fmt.Println("No orders, stopped elevator.")
			thisElevator.SetState(utils.Still) // If no orders, set the state to still
			thisElevator.GeneralDirection = utils.Stopped
		}
	}
}

func HandleButtonEvent(newOrder utils.Order, thisElevator *utils.Elevator, channels *utils.Channels) {

	// HandleButtonEvent handles a button event by processing the new order and updating the global order system.
	// It takes in the following parameters:
	// - newOrderTx: a channel for sending a new order message
	// - orderCompleteTx: a channel for sending an order complete message
	// - newOrder: the new order to be processed
	// - e: a pointer to the elevator object

	fmt.Println("Function: HandleButtonEvent")

	if !orders.CheckIfGlobalOrderIsActive(newOrder, thisElevator.ID) { // Check if the order is already active

		channels.GlobalUpdateCh <- utils.GlobalOrderUpdate{
			Orders:         []utils.Order{newOrder},
			FromElevatorID: thisElevator.ID,
			IsComplete:     false,
			IsNew:          true}

		button := newOrder.Button
		//floor := newOrder.Floor

		if button == utils.Cab {

			fmt.Println("Cab order")

			order := utils.MessageNewOrder{
				Type:           "MessageNewOrder",
				NewOrder:       newOrder,
				ToElevatorID:   utils.NotDefined,
				FromElevatorID: thisElevator.ID}

			channels.NewOrderTx <- order

			if !thisElevator.IsMaster {


				utils.SlaveOrderWatcher.CabOrderArray[thisElevator.ID][newOrder.Floor].Time = time.Now()
				utils.SlaveOrderWatcher.CabOrderArray[thisElevator.ID][newOrder.Floor].Confirmed = false
				utils.SlaveOrderWatcher.CabOrderArray[thisElevator.ID][newOrder.Floor].Active = true

				go WaitForAck(channels.AckRx, utils.Timeout, newOrder, thisElevator)
			}

			if orders.CheckIfLocalOrderIsActive(newOrder, thisElevator) { // Check if the order is active
				if utils.BestOrder.Floor == thisElevator.CurrentFloor && elevio.GetFloor() != utils.NotDefined {
					HandleElevatorAtFloor(utils.BestOrder.Floor, channels, thisElevator) // Handle the elevator at the floor
				} else {

					fmt.Println("Best order is", utils.BestOrder)
					DoOrder(utils.BestOrder, thisElevator, channels) // Move the elevator to the best order
				}

			} else {

				fmt.Println("Going into ProcessElevatorOrders")
				ProcessElevatorOrders(newOrder, thisElevator, channels)

			}
		} else {

			fmt.Println("Hall order")

			if thisElevator.IsMaster {

				fmt.Println("This is the master")
				// Handle order locally (remember lights)
				bestElevator := orders.ChooseElevator(newOrder)

				if bestElevator.ID == thisElevator.ID {

					fmt.Println("Handling locally")

					go ProcessElevatorOrders(newOrder, thisElevator, channels)

					go func() {

						newOrder := utils.MessageNewOrder{
							Type:           "MessageNewOrder",
							NewOrder:       newOrder,
							ToElevatorID:   utils.NotDefined, // Use the correct field name as defined in your ElevatorStatus struct
							FromElevatorID: thisElevator.ID}

						fmt.Println("Sending order")

						channels.NewOrderTx <- newOrder

					}()

					go func() {

						newOrder := utils.OrderWatcher{
							Orders:        []utils.Order{newOrder},
							ForElevatorID: thisElevator.ID,
							IsComplete:    false,
							IsNew:         true}

						channels.OrderWatcher <- newOrder

					}()

				} else {

					fmt.Println("Sending order")

					go func() {

						order := utils.MessageNewOrder{
							Type:           "MessageNewOrder",
							NewOrder:       newOrder,
							ToElevatorID:   bestElevator.ID, // Use the correct field name as defined in your ElevatorStatus struct
							FromElevatorID: thisElevator.ID}

						channels.NewOrderTx <- order

					}()

					go func() {
						channels.OrderWatcher <- utils.OrderWatcher{
							Orders:        []utils.Order{newOrder},
							ForElevatorID: bestElevator.ID,
							IsComplete:    false,
							IsNew:         true}
					}()

					if orders.CheckAmountOfActiveOrders(thisElevator) > 0 {
						DoOrder(utils.BestOrder, thisElevator, channels)
					}

				}

			} else {

				fmt.Println("Sending order to master...")
				// Send order to master
				order := utils.MessageNewOrder{
					Type:           "MessageNewOrder",
					NewOrder:       newOrder,
					ToElevatorID:   utils.MasterElevatorID,
					FromElevatorID: thisElevator.ID}

				channels.NewOrderTx <- order

				utils.SlaveOrderWatcher.HallOrderArray[newOrder.Button][newOrder.Floor].Time = time.Now()
				utils.SlaveOrderWatcher.HallOrderArray[newOrder.Button][newOrder.Floor].Confirmed = false
				utils.SlaveOrderWatcher.HallOrderArray[thisElevator.ID][newOrder.Floor].Active = true


				go WaitForAck(channels.AckRx, utils.Timeout, newOrder, thisElevator)

				if orders.CheckAmountOfActiveOrders(thisElevator) > 0 {

					DoOrder(utils.BestOrder, thisElevator, channels)

				} else {

					thisElevator.StopElevator()

				}

			}
		}
	}
}

func ProcessElevatorOrders(newOrder utils.Order, thisElevator *utils.Elevator, channels *utils.Channels) {

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

	orders.UpdateLocalOrderSystem(newOrder, thisElevator)

	orders.PrintLocalOrderSystem(thisElevator)

	amountOfOrders := orders.CheckAmountOfActiveOrders(thisElevator)

	fmt.Println("Amount of active orders: ", amountOfOrders)

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
}

func DoOrder(order utils.Order, thisElevator *utils.Elevator, channels *utils.Channels) {

	// DoOrder executes the given order for the elevator.
	// It moves the elevator up or down to reach the order's floor.
	// If the order's floor is the same as the current floor of the elevator,
	// it waits for a possible order completion.
	// The function takes the order to be executed, a channel to send order completion messages,
	// and a pointer to the elevator object.

	// Do the order
	if order.Floor > thisElevator.CurrentFloor {

		thisElevator.GoUp()

	} else if order.Floor < thisElevator.CurrentFloor {

		thisElevator.GoDown()

	} else {
		// The order is at the current floor
		go WaitForPossibleOrder(order, thisElevator, channels)

	}

}

func WaitForPossibleOrder(order utils.Order, thisElevator *utils.Elevator, channels *utils.Channels) {

	// WaitForPossibleOrder waits for a possible order to be completed.
	// It takes an order, a channel for sending order completion messages, and an elevator as parameters.
	// If the elevator is at the same floor as the order, it stops the elevator and handles the order completion.
	// It also inserts a delay in case cab-orders come in.

	time.Sleep(3 * time.Second) // Insert delay in case cab-orders come in.
	if thisElevator.CurrentFloor == order.Floor {
		thisElevator.StopElevator()
		HandleElevatorAtFloor(order.Floor, channels, thisElevator)
	}

}
