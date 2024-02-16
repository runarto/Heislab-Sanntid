package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
	"fmt"
)

func (e *Elevator) InitLocalOrderSystem() {
	for i := 0; i < elevio._numFloors; i++ {
		for j := 0; j < elevio._numButtons; j++ {
			e.LocalOrderArray[i][j] = False // 
		}
	}  

	// Initialies the order system with all zeros. 
}

func (e *Elevator) CheckAmountOfActiveOrders() int {
	ActiveOrders := 0
	for i := 0; i < elevio._numFloors; i++ {
		for j := 0; j < elevio._numButtons; j++ {
			if e.LocalOrderArray[i][j] == True { // If there is an active order
				ActiveOrders++ // Increment the number of active orders
			}
		}
	}
	return ActiveOrders // Return the number of active orders
}


func (e *Elevator) ChooseBestOrder() Order {

	if e.CurrentDirection == Up {
		if e.CheckAbove(e.CurrentFloor).Floor != NotDefined { // If there are any orders above the elevator
			return e.CheckAbove(e.CurrentFloor) // Return the closest order
		}
		else if e.CheckBelow(e.CurrentFloor).Floor != NotDefined { //If there are no orders above, check below
			return e.CheckBelow(e.CurrentFloor) // Return the closest order
		}
	}

	if e.CurrentDirection == Down {
		if e.CheckBelow(e.CurrentFloor).Floor != NotDefined {
			return e.CheckBelow(e.CurrentFloor)
		}
		else if e.CheckAbove(e.CurrentFloor).Floor != NotDefined {
			return e.CheckAbove(e.CurrentFloor)
		}
	}

}

func (e *Elevator) CheckAbove(floor int) Order {
	// Check if there are any orders above the elevator
	bestOrder = Order{NotDefined, 2} // Initialize the best order
	for button := 0; button < elevio._numButtons; button++ {
		for floorOrder := floor; floorOrder < elevio._numFloors; i++ {
			if e.LocalOrderArray[button][floorOrder] == True && button != HallDown { 
				if abs(floorOrder - floor) <= abs(bestOrder.Floor - floor) {
					Order = Order{floorOrder, button} // Return the order
				}
			}
		}
	}
	return bestOrder // Return the best order

	// Potential issue: If all orders are active at an ideal floor, the elevator will do the cab first. 
	// How can we assert that it does the hall orders next?


}

func (e *Elevator) CheckBelow(floor int) Order {
	// Check if there are any orders below the elevator
	bestOrder = Order{NotDefined, 2} // Initialize the best order
	for button := 0; button < elevio._numButtons; button++ {
		for floorOrder := floor; floorOrder >= 0; i-- {
			if e.LocalOrderArray[button][floorOrder] == True && button != HallUp {
				if abs(floorOrder - floor) <= abs(bestOrder.Floor - floor) {
					Order = Order{floorOrder, button} // Return the order
				}
			}
		}
	}
	return bestOrder // Return the best order
}





func (e *Elevator) UpdateOrderSystem(order Order) {
	floor := order.Floor // The floor the order is at 
	button := order.Button // Type of order (Up, Down, Cab)

	if e.LocalOrderArray[floor][button] == True { // If the order is already in the local order array
		e.LocalOrderArray[floor][button] = False // Remove the order from the local order array
		elevio.SetButtonLamp(button, floor, False) // Turn off the button lamp
	}
	else {
		e.LocalOrderArray[floor][button] =  True // Add the order to the local order array
		elevio.SetButtonLamp(button, floor, True) // Turn on the button lamp
	}


}

func (e* Elevator) DoOrder(order Order)  {
	// Do the order
	if order.Floor > e.CurrentFloor {
		e.GoUp()
	}
	else if order.Floor < e.CurrentFloor {
		e.GoDown()
	}
	else {
		e.StopElevator()
	}
	
}