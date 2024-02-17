package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
	"math"
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

	if e.CurrentDirection == Up {
		return e.CheckAbove(e.CurrentFloor)
	} else {
		return e.CheckBelow(e.CurrentFloor)
	}

}

func (e *Elevator) CheckAbove(floor int) Order {
	// Check if there are any orders above the elevator
	var bestOrder = Order{NotDefined, 2} // Initialize the best order
	var secondBestOrder = Order{NotDefined, 2}
	for button := 0; button < numButtons; button++ {
		for floorOrder := floor; floorOrder < numFloors; floorOrder++ {
			if e.LocalOrderArray[button][floorOrder] == True { 
				if button == HallUp || button == Cab {
					if bestOrder.Floor == NotDefined {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						bestOrder = Order
						continue
					} else if Abs(order.Floor - floor) <= Abs(bestOrder.Floor - floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						bestOrder = Order
						continue
						}
					} else {
						if secondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							secondBestOrder = Order
							continue
						} else if Abs(order.Floor - floor) >= Abs(secondBestOrder.Floor - floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							secondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if bestOrder.Floor != NotDefined {
		return bestOrder
	} else {
		return secondBestOrder
	}
}


func (e *Elevator) CheckBelow(floor int) Order {
	// Check if there are any orders above the elevator
	var bestOrder = Order{NotDefined, 2} // Initialize the best order
	var secondBestOrder = Order{NotDefined, 2}

	for button := 0; button < numButtons; button++ {
		for floorOrder := floor; floorOrder < numFloors; floorOrder++ {
			if e.LocalOrderArray[button][floorOrder] == True { 
				if button == HallDown || button == Cab {
					if bestOrder.Floor == NotDefined {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						bestOrder = Order
						continue
					} else if Abs(order.Floor - floor) =< Abs(bestOrder.Floor - floor) {
						Order := Order{floorOrder, elevio.ButtonType(button)}
						bestOrder = Order
						continue
						}
					} else {
						if secondBestOrder.Floor == NotDefined {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							secondBestOrder = Order
							continue
						} else if Abs(order.Floor - floor) >= Abs(secondBestOrder.Floor - floor) {
							Order := Order{floorOrder, elevio.ButtonType(button)}
							secondBestOrder = Order
							continue
						}


				}
			}
		} 
	}
	if bestOrder.Floor != NotDefined {
		return bestOrder
	} else {
		return secondBestOrder
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
		e.GoUp()
	} else if order.Floor < e.CurrentFloor {
		e.GoDown()
	} else {
		e.StopElevator()
	}
	
}