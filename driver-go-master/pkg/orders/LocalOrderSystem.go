package orders

import (
	"fmt"
	"math"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func AbsValue(x int, y int) int {
	return int(math.Abs(float64(x - y)))
}

func PrintLocalOrderSystem(e *utils.Elevator) {
	for i := 0; i < utils.NumButtons; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			fmt.Print(e.LocalOrderArray[i][j], " ")
		}
		fmt.Println()
	}

}

func InitLocalOrderSystem(e *utils.Elevator) {
	for i := 0; i < utils.NumButtons; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			e.LocalOrderArray[i][j] = utils.False //
		}
	}

	// Initialies the order system with all zeros.
}

func CheckAmountOfActiveOrders(e *utils.Elevator) int {
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
	fmt.Println("Function: CheckAbove")
	// Check if there are any orders above the elevator
	var CurrentBestOrder utils.Order
	var CurrentSecondBestOrder utils.Order

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

	fmt.Println("Function: CheckBelow")
	// Check if there are any orders above the elevator
	var CurrentBestOrder utils.Order // Initialize the best order
	var CurrentSecondBestOrder utils.Order

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
	fmt.Println("Function: CheckHallOrdersAbove")
	// Check if there are any orders above the elevator
	var CurrentBestOrder utils.Order // Initialize the best order
	var CurrentSecondBestOrder utils.Order

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
	fmt.Println("Function: CheckHallOrdersBelow")
	// Check if there are any orders above the elevator
	var CurrentBestOrder utils.Order // Initialize the best order
	var CurrentSecondBestOrder utils.Order

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

func UpdateOrderSystem(order utils.Order, e *utils.Elevator) {

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

	if e.LocalOrderArray[order.Button][order.Floor] == utils.True {
		return true
	} else {
		return false
	}
}
