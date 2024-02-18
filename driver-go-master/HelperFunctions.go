package main

import (
	"math"
	"fmt"
	"time"
)

func AbsValue(x int, y int) int { 
	return int(math.Abs(float64(x - y)))
}


func (e *Elevator) HandleElevatorAtFloor(floor int) {

	if e.ElevatorAtFloor(floor) { // Check for active orders at floor
		e.StopElevator() // Stop the elevator
		e.SetDoorState(Open) // Open the door
		time.Sleep(1000 * time.Millisecond) // Wait for a second
		e.SetDoorState(Close) // Close the door
		fmt.Println("Order system: ")
		e.PrintLocalOrderSystem()
		amountOfOrders := e.CheckAmountOfActiveOrders() // Check the amount of active orders
		fmt.Println("Amount of active orders: ", amountOfOrders)
		if amountOfOrders > 0 {
			bestOrder = e.ChooseBestOrder() // Choose the best order
			fmt.Println("Best order: ", bestOrder)
			e.DoOrder(bestOrder)
			// DoOrder(order) // Move the elevator to the best order (pseudocode function to move the elevator to the best order
		} else {
			e.SetState(Still) // If no orders, set the state to still
		}
	}
}

func (e *Elevator) CheckIfOrderIsActive(order Order) bool {

	if e.LocalOrderArray[order.Button][order.Floor] == True {
		return true
	} else {
		return false
	}
}