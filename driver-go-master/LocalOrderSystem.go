package main

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
	"math"
)

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
		if e.CheckAbove(e.CurrentFloor).Floor != NotDefined { // If there are any orders above the elevator
			return e.CheckAbove(e.CurrentFloor) // Return the closest order
		} else if e.CheckBelow(e.CurrentFloor).Floor != NotDefined { //If there are no orders above, check below
			return e.CheckBelow(e.CurrentFloor) // Return the closest order
		}
	}

	if e.CurrentDirection == Down {
		if e.CheckBelow(e.CurrentFloor).Floor != NotDefined {
			return e.CheckBelow(e.CurrentFloor)
		} else if e.CheckAbove(e.CurrentFloor).Floor != NotDefined {
			return e.CheckAbove(e.CurrentFloor)
		}
	}

	Order := Order{0,0}
	return Order

}

func (e *Elevator) CheckAbove(floor int) Order {
	// Check if there are any orders above the elevator
	var bestOrder = Order{NotDefined, 2} // Initialize the best order
	for button := 0; button < numButtons; button++ {
		for floorOrder := floor; floorOrder < numFloors; floorOrder++ {
			if e.LocalOrderArray[button][floorOrder] == True && button != HallDown { 
				if math.Abs(float64(floorOrder - floor)) <= math.Abs(float64(bestOrder.Floor - floor)) {
					Order := Order{floorOrder, elevio.ButtonType(button)} // Return the order
					bestOrder = Order
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
	bestOrder := Order{NotDefined, 2} // Initialize the best order
	for button := 0; button < numButtons; button++ {
		for floorOrder := floor; floorOrder >= 0; floorOrder-- {
			if e.LocalOrderArray[button][floorOrder] == True && button != HallUp {
				if math.Abs(float64(floorOrder - floor)) <= math.Abs(float64(bestOrder.Floor - floor)) {
					Order := Order{floorOrder, elevio.ButtonType(button)} // Return the order
					bestOrder = Order
				}
			}
		}
	}
	return bestOrder // Return the best order
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