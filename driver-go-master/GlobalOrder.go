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
	for i, _ := range elevators {
		if Elevators[i].isActive {
			cost := CalculateCost(elevator, order)
			if cost <= lowestCost {
				lowestCost = cost
				bestElevator = elevator
			}
		}
	}

	return bestElevator
}


func CheckIfGlobalOrderIsActive(order Order) bool {

	if order.Button == Cab {
		if globalOrderArray.CabOrderArray[order.Floor] == True {
			return true
		} else {
			return false
		}
	} else {
		if globalOrderArray.HallOrderArray[order.Button][order.Floor] == True {
			return true
		} else {
			return false
		}
	}

}


func UpdateGlobalOrderSystem(order Order, e Elevator, value bool) {

	HallOrderArray := globalOrderArray.HallOrderArray
	CabOrderArray := globalOrderArray.CabOrderArray

	if value {
		if order.Button == Cab {
			CabOrderArray[e.ID][order.Floor] = True
		} else {
			HallOrderArray[order.Button][order.Floor] = True
			e.SetButtonLamp(order.Button, order.Floor, true)
		}
	} else {
		if order.Button == Cab {
			CabOrderArray[e.ID][order.Floor] = False
		} else {
			HallOrderArray[order.Button][order.Floor] = False
			e.SetButtonLamp(order.Button, order.Floor, false)
		}
	}
	

}


func (e *Elevator) RedistributeHallOrders(offlineElevator Elevator) {

	for button := 0; button < numButtons-1; button++ { // Don't change Cab-orders
		for floor := 0; floor < numFloors; floor++ {
			if offlineElevator.LocalOrderArray[button][floor] == True {

				Order = Order{floor, elevio.ButtonType(button)}

				UpdateGlobalOrderSystem(Order, e, false)
				offlineElevator.LocalOrderArray[button][floor] = False

				bestElevator := chooseElevator(Elevators, Order)
				
				UpdateGlobalOrderSystem(Order, bestElevator, true)

				if bestElevator.ID == e.ID {

					e.LocalOrderArray[button][floor] = True

				} else {

					MessageNewOrder = MessageNewOrder{
						Type: "MessageNewOrder",
						NewOrder: Order,
						E: bestElevator,
						ToElevatorID: bestElevator.ID,}

					newOrderTx <- MessageNewOrder
				}

			}
		}
	}
}
