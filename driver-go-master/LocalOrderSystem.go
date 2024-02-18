package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
	"fmt"
)

func (e *Elevator) PrintLocalOrderSystem() {
	for i := 0; i < numButtons; i++ {
		for j := 0; j < numFloors; j++ {
			fmt.Print(e.LocalOrderArray[i][j], " ")
		}
		fmt.Println()
	}

}

func (e *Elevator) InitLocalOrderSystem() {
	for i := 0; i < numButtons; i++ {
		for j := 0; j < numFloors; j++ {
			e.LocalOrderArray[i][j] = False // 
		}
	}  

	// Initialies the order system with all zeros. 
}

func (e *Elevator) CheckAmountOfActiveOrders() int {
	ActiveOrders := 0
	for i := 0; i < numButtons; i++ {
		for j := 0; j < numFloors; j++ {
			if e.LocalOrderArray[i][j] == True { // If there is an active order
				ActiveOrders++ // Increment the number of active orders
			}
		}
	}
	return ActiveOrders // Return the number of active orders
}


func (e *Elevator) ChooseBestOrder() Order {

	orderAbove := e.CheckAbove(e.CurrentFloor)
	orderBelow := e.CheckBelow(e.CurrentFloor)

	if e.CurrentDirection == Up {
		if e.CurrentFloor == 3 {
			fmt.Println("Check below")
			return orderBelow
		} else {
			fmt.Println("Check above")
			if orderAbove.Floor == NotDefined {
				return orderBelow
			}
			return orderAbove
		}
	} else {
		if e.CurrentFloor == 0 {
			fmt.Println("Check above")
			return orderAbove
		} else {
			fmt.Println("Check below")
			if orderBelow.Floor == NotDefined {
				return orderAbove
			}
			return orderBelow
		}
	}

}

func (e *Elevator) CheckAbove(floor int) Order {
	// Check if there are any orders above the elevator
	var CurrentBestOrder = Order{NotDefined, 2} // Initialize the best order
	var CurrentSecondBestOrder = Order{NotDefined, 2}

	for button := 0; button < numButtons; button++ {
		for floorOrder := floor; floorOrder < numFloors; floorOrder++ { // Iterate over LocalOrderArray
			if e.LocalOrderArray[button][floorOrder] == True { // If there is an active order

				fmt.Println("Button: ", button, "Floor: ", floorOrder)

				if button == HallUp || button == Cab { // If the order is an up order or a cab order
					if CurrentBestOrder.Floor == NotDefined { // If the best order is not defined
						Order := Order{floorOrder, elevio.ButtonType(button)} 
						CurrentBestOrder = Order // Set the best order to the current order
						continue // Continue to the next iteration
					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
						}
					} else {
						if CurrentSecondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if CurrentBestOrder.Floor != NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}


func (e *Elevator) CheckBelow(floor int) Order {
	// Check if there are any orders above the elevator
	var CurrentBestOrder = Order{NotDefined, 2} // Initialize the best order
	var CurrentSecondBestOrder = Order{NotDefined, 2}

	for button := 0; button < numButtons; button++ {
		fmt.Println("Button: ", button)
		for floorOrder := 0; floorOrder <= floor; floorOrder++ {
			fmt.Println("Floor: ", floorOrder)
			if e.LocalOrderArray[button][floorOrder] == True { 
				if button == HallDown || button == Cab {
					if CurrentBestOrder.Floor == NotDefined {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
						}
					} else {
						if CurrentSecondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if CurrentBestOrder.Floor != NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}



func (e *Elevator) CheckHallOrdersAbove(floor int) Order {
	// Check if there are any orders above the elevator
	var CurrentBestOrder = Order{NotDefined, 2} // Initialize the best order
	var CurrentSecondBestOrder = Order{NotDefined, 2}

	for button := 0; button < numButtons-1; button++ {
		for floorOrder := floor; floorOrder < numFloors; floorOrder++ { // Iterate over LocalOrderArray
			if e.LocalOrderArray[button][floorOrder] == True { // If there is an active order

				fmt.Println("Button: ", button, "Floor: ", floorOrder)

				if button == HallUp { // If the order is an up order or a cab order
					if CurrentBestOrder.Floor == NotDefined { // If the best order is not defined
						Order := Order{floorOrder, elevio.ButtonType(button)} 
						CurrentBestOrder = Order // Set the best order to the current order
						continue // Continue to the next iteration
					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
						}
					} else {
						if CurrentSecondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if CurrentBestOrder.Floor != NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}

func (e *Elevator) CheckHallOrdersBelow(floor int) Order {
	// Check if there are any orders above the elevator
	var CurrentBestOrder = Order{NotDefined, 2} // Initialize the best order
	var CurrentSecondBestOrder = Order{NotDefined, 2}

	for button := 0; button < numButtons-1; button++ {
		fmt.Println("Button: ", button)
		for floorOrder := 0; floorOrder <= floor; floorOrder++ {
			fmt.Println("Floor: ", floorOrder)
			if e.LocalOrderArray[button][floorOrder] == True { 
				if button == HallDown {
					if CurrentBestOrder.Floor == NotDefined {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
					} else if AbsValue(floorOrder, floor) <= AbsValue(CurrentBestOrder.Floor, floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						CurrentBestOrder = Order
						continue
						}
					} else {
						if CurrentSecondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						} else if AbsValue(floorOrder, floor) >= AbsValue(CurrentSecondBestOrder.Floor, floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							CurrentSecondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if CurrentBestOrder.Floor != NotDefined {
		return CurrentBestOrder
	} else {
		return CurrentSecondBestOrder
	}
}




func (e *Elevator) UpdateOrderSystem(order Order) {
	floor := order.Floor // The floor the order is at 
	button := order.Button // Type of order (Up, Down, Cab)

	if e.LocalOrderArray[button][floor] == True { // If the order is already in the local order array
		e.LocalOrderArray[button][floor] = False // Remove the order from the local order array
		elevio.SetButtonLamp(button, floor, false) // Turn off the button lamp
	} else {
		e.LocalOrderArray[button][floor] =  True // Add the order to the local order array
		elevio.SetButtonLamp(button, floor, true) // Turn on the button lamp
	}


}

func (e* Elevator) DoOrder(order Order)  {
	// Do the order
	if order.Floor > e.CurrentFloor {
		if e.CurrentDirection == Up {
			return
		} else {
			e.GoUp()
		}
	} else if order.Floor < e.CurrentFloor {
		if e.CurrentDirection == Down {
			return
		} else {
			e.GoDown()
		}
	} else {
		e.StopElevator()
	}
	
}