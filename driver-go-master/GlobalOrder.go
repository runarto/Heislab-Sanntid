package main

// Idea: calculate the cost of each elevator and choose the one with the lowest cost

func CalculateCost(elevator Elevator, order Order) int {

	cost := elevator.CheckAmountOfActiveOrders()
	return cost

}

// Function to find the best elevator for a given order
func chooseElevator(elevators []Elevator, order Order) Elevator {
	//Initiate variables
	var bestElevator Elevator
	lowestCost := int(^uint(0) >> 1) // Sets "lowestCost" to max int value

	//Iterate through all elevators and calculate the cost for each. Update bestElevator if a lower cost is found
	for _, elevator := range elevators {
		cost := CalculateCost(elevator, order)
		if cost <= lowestCost {
			lowestCost = cost
			bestElevator = elevator
		}
	}

	return bestElevator
}
