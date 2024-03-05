package orders

import (
	"fmt"
	"math"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func AbsValue(x int, y int) int {

	// AbsValue calculates the absolute difference between two integers.
	// It takes two integers, x and y, and returns the absolute difference as an integer.

	return int(math.Abs(float64(x - y)))
}

func PrintLocalOrderSystem(e *utils.Elevator) {

	// PrintLocalOrderSystem prints the local order system of an elevator.
	// It takes a pointer to an Elevator struct as a parameter.
	// The function iterates over the local order array of the elevator and prints its contents.
	// Each element of the array represents the status of a button press on a specific floor.
	// The function does not return any value.

	for i := 0; i < utils.NumButtons; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			fmt.Print(e.LocalOrderArray[i][j], " ")
		}
		fmt.Println()
	}

}

func InitLocalOrderSystem(e *utils.Elevator) {

	// InitLocalOrderSystem initializes the local order system for the elevator.
	// It sets all elements in the LocalOrderArray of the elevator to false.

	for i := 0; i < utils.NumButtons; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			e.LocalOrderArray[i][j] = utils.False //
		}
	}

	// Initialies the order system with all zeros.
}

func CheckAmountOfActiveOrders(e *utils.Elevator) int {

	// CheckAmountOfActiveOrders calculates the number of active orders in the given elevator.
	// It iterates through the LocalOrderArray of the elevator and counts the number of orders that are marked as active.
	// The function returns the total number of active orders.

	ActiveOrders := 0
	for i := 0; i < utils.NumButtons; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			if e.LocalOrderArray[i][j] == utils.True { // If there is an active order
				ActiveOrders++ // Increment the number of active orders
			}
		}
	}
	return ActiveOrders // Return the number of active orders
}

func ChooseBestOrder(e *utils.Elevator) utils.Order {

	// ChooseBestOrder selects the best order for the given elevator.
	// It considers the current floor and direction of the elevator to determine the best order.
	// If the elevator is moving up, it checks for orders above the current floor and returns the order with the highest floor.
	// If there are no orders above, it checks for orders below and returns the order with the lowest floor.
	// If the elevator is moving down, it checks for orders below the current floor and returns the order with the lowest floor.
	// If there are no orders below, it checks for orders above and returns the order with the highest floor.
	// If there are no orders in either direction, it returns an order with the floor set to utils.NotDefined.

	orderAbove := CheckAbove(e.CurrentFloor, e)
	orderBelow := CheckBelow(e.CurrentFloor, e)

	if e.CurrentDirection == utils.Up {
		if e.CurrentFloor == 3 {
			fmt.Println("Check below (Up)")
			return orderBelow
		} else {
			fmt.Println("Check above (Up)")
			if orderAbove.Floor == utils.NotDefined {
				return orderBelow
			}
			return orderAbove
		}

	} else {
		if e.CurrentFloor == 0 {
			fmt.Println("Check above (Down)")
			return orderAbove
		} else {
			fmt.Println("Check below (Down)")
			if orderBelow.Floor == utils.NotDefined {
				return orderAbove
			}
			return orderBelow
		}
	}

}

func CheckAbove(floor int, e *utils.Elevator) utils.Order {

	// CheckAbove checks if there are any orders above the given floor in the elevator system.
	// It iterates over the local order array and finds the best and second best orders based on the given floor.
	// The best order is determined by the closest floor above the given floor, while the second best order is determined by the farthest floor above the given floor.
	// If there are no orders above the given floor, it returns the second best order.
	// If there are no orders at all, it returns an order with floor and button set to utils.NotDefined.

	fmt.Println("Function: CheckAbove")
	// Check if there are any orders above the elevator
	CurrentBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	CurrentSecondBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	for button := 0; button < utils.NumButtons; button++ {
		for floorOrder := floor; floorOrder < utils.NumFloors; floorOrder++ { // Iterate over LocalOrderArray
			if e.LocalOrderArray[button][floorOrder] == utils.True { // If there is an active order

				if button == utils.HallUp || button == utils.Cab { // If the order is an up order or a cab order
					if CurrentBestOrder.Floor == utils.NotDefined { // If the best order is not defined
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order // Set the best order to the current order
						continue                 // Continue to the next iteration

					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order
						continue
					}
				} else {

					if CurrentSecondBestOrder.Floor == utils.NotDefined {

						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}
						CurrentSecondBestOrder = Order
						continue

					} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}
						CurrentSecondBestOrder = Order
						continue
					}

				}
			}
		}
	}
	if CurrentBestOrder.Floor != utils.NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}

func CheckBelow(floor int, e *utils.Elevator) utils.Order {

	// CheckBelow checks for orders below the given floor in the local order system.
	// It iterates through the local order array and finds the best and second best orders
	// based on their proximity to the given floor. The best order is the one with the smallest
	// absolute difference between its floor and the given floor, while the second best order
	// is the one with the largest absolute difference. If there are no orders below the given floor,
	// it returns the second best order with floor set to utils.NotDefined.

	fmt.Println("Function: CheckBelow")
	// Check if there are any orders above the elevator
	CurrentBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	CurrentSecondBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	for button := 0; button < utils.NumButtons; button++ {
		for floorOrder := 0; floorOrder <= floor; floorOrder++ {
			if e.LocalOrderArray[button][floorOrder] == utils.True {
				if button == utils.HallDown || button == utils.Cab {
					if CurrentBestOrder.Floor == utils.NotDefined {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order
						continue
					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {

						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order
						continue
					}
				} else {

					if CurrentSecondBestOrder.Floor == utils.NotDefined {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue

					} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue
					}

				}
			}
		}
	}
	if CurrentBestOrder.Floor != utils.NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}

func CheckHallOrdersAbove(floor int, e *utils.Elevator) utils.Order {

	// CheckHallOrdersAbove checks for hall orders above the given floor in the elevator system.
	// It returns the best order (up order or cab order) that is closest to the given floor.
	// If there are no orders above the floor, it returns the second best order.

	fmt.Println("Function: CheckHallOrdersAbove")
	// Check if there are any orders above the elevator
	CurrentBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined} // Initialize the best order

	CurrentSecondBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	for button := 0; button < utils.NumButtons-1; button++ {
		for floorOrder := floor; floorOrder < utils.NumFloors; floorOrder++ { // Iterate over LocalOrderArray
			if e.LocalOrderArray[button][floorOrder] == utils.True { // If there is an active order

				if button == utils.HallUp { // If the order is an up order or a cab order

					if CurrentBestOrder.Floor == utils.NotDefined { // If the best order is not defined

						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order // Set the best order to the current order
						continue                 // Continue to the next iteration

					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
					}
				} else {

					if CurrentSecondBestOrder.Floor == utils.NotDefined {

						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue

					} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue
					}

				}
			}
		}
	}
	if CurrentBestOrder.Floor != utils.NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}

func CheckHallOrdersBelow(floor int, e *utils.Elevator) utils.Order {

	// CheckHallOrdersBelow checks for hall orders below the given floor in the elevator system.
	// It iterates through the local order array and finds the best and second best order based on the distance from the current floor.
	// If there are no orders below the given floor, it returns the second best order.
	// If there are no orders at all, it returns an order with NotDefined values.

	fmt.Println("Function: CheckHallOrdersBelow")
	// Check if there are any orders above the elevator
	CurrentBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	CurrentSecondBestOrder := utils.Order{
		Floor:  utils.NotDefined,
		Button: utils.NotDefined}

	for button := 0; button < utils.NumButtons-1; button++ {
		for floorOrder := 0; floorOrder <= floor; floorOrder++ {
			if e.LocalOrderArray[button][floorOrder] == utils.True {
				if button == utils.HallDown {

					if CurrentBestOrder.Floor == utils.NotDefined {

						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue

					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentBestOrder = Order
						continue
					}
				} else {

					if CurrentSecondBestOrder.Floor == utils.NotDefined {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue

					} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
						Order := utils.Order{
							Floor:  floorOrder,
							Button: elevio.ButtonType(button)}

						CurrentSecondBestOrder = Order
						continue
					}

				}
			}
		}
	}
	if CurrentBestOrder.Floor != utils.NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}

func UpdateLocalOrderSystem(order utils.Order, e *utils.Elevator) {

	// UpdateLocalOrderSystem updates the local order system based on the given order and elevator state.
	// It toggles the order in the local order array and turns on/off the corresponding button lamp.
	// If the order is already in the local order array, it removes the order and turns off the button lamp.
	// If the order is not in the local order array, it adds the order and turns on the button lamp.

	floor := order.Floor   // The floor the order is at
	button := order.Button // Type of order (Up, Down, Cab)

	if e.LocalOrderArray[button][floor] == utils.True { // If the order is already in the local order array

		e.LocalOrderArray[button][floor] = utils.False // Remove the order from the local order array
		elevio.SetButtonLamp(button, floor, false)     // Turn off the button lamp

	} else {

		e.LocalOrderArray[button][floor] = utils.True // Add the order to the local order array
		elevio.SetButtonLamp(button, floor, true)     // Turn on the button lamp

	}

}

func CheckIfOrderIsActive(order utils.Order, e *utils.Elevator) bool {

	// CheckIfOrderIsActive checks if the specified order is active for the given elevator.
	// It returns true if the order is active, otherwise false.

	if e.LocalOrderArray[order.Button][order.Floor] == utils.True {
		return true
	} else {
		return false
	}
}
