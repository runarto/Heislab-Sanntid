package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func OrderCompleted(order utils.Order, e *utils.Elevator) {

	// OrderCompleted updates the status of a completed order in the OrderWatcher.
	// It takes an order and an elevator as input parameters.
	// If the order is a cab order, it marks the corresponding cab order as completed,
	// sets the completion time, and marks it as inactive.
	// If the order is a hall order, it marks the corresponding hall order as completed,
	// sets the completion time, and marks it as inactive.

	fmt.Println("Function: OrderCompleted")

	button := order.Button
	floor := order.Floor

	if button == utils.Cab {
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Completed = true
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Time = time.Now()
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Active = false
	} else {
		utils.OrderWatcher.HallOrderArray[button][floor].Completed = true
		utils.OrderWatcher.HallOrderArray[button][floor].Time = time.Now()
		utils.OrderWatcher.HallOrderArray[button][floor].Active = false
	}
}

func OrderActive(order utils.Order, e *utils.Elevator, time time.Time) {

	// OrderActive updates the status of an order in the OrderWatcher based on the given order and elevator.
	// If the order is a cab order, it sets the corresponding cab order as active and updates the time.
	// If the order is a hall order, it sets the corresponding hall order as active and updates the time.
	// Parameters:
	// - order: The order to be updated.
	// - e: The elevator associated with the order.

	fmt.Println("Function: OrderActive")

	button := order.Button
	floor := order.Floor

	if button == utils.Cab {
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Active = true
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Completed = false
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Time = time

	} else {
		utils.OrderWatcher.HallOrderArray[button][floor].Active = true
		utils.OrderWatcher.HallOrderArray[button][floor].Completed = false
		utils.OrderWatcher.HallOrderArray[button][floor].Time = time
	}
}

func CheckIfOrderIsComplete(e *utils.Elevator, newOrderTx chan utils.MessageNewOrder, orderCompleteTx chan utils.MessageOrderComplete) {

	// CheckIfOrderIsComplete checks if any hall or cab orders have not been completed within 15 seconds.
	// If an order is not completed, it reassigns the order to the best available elevator.
	// It also updates the order status and sends new order messages if necessary.
	// Finally, it updates the hall and cab order arrays.

	fmt.Println("Function: CheckIfOrderIsComplete")

	currentTime := time.Now()
	var ordersToBeReAssigned []utils.Order

	HallOrderArray := utils.OrderWatcher.HallOrderArray

	for button := 0; button < 2; button++ {
		for floor := 0; floor < utils.NumFloors; floor++ {
			if HallOrderArray[button][floor].Active == true && HallOrderArray[button][floor].Completed == false {
				if currentTime.Sub(HallOrderArray[button][floor].Time) > 15*time.Second {

					fmt.Println("Order", button, "at floor", floor, "is not completed. Reassigning order.")

					ordersToBeReAssigned = append(ordersToBeReAssigned, utils.Order{
						Floor:  floor,
						Button: elevio.ButtonType(button)})

					HallOrderArray[button][floor].Active = false
				}
			}
		}
	}

	CabOrderArray := utils.OrderWatcher.CabOrderArray

	for i, _ := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			for floor := 0; floor < utils.NumFloors; floor++ {
				if CabOrderArray[utils.Elevators[i].ID][floor].Active == true && CabOrderArray[utils.Elevators[i].ID][floor].Completed == false {
					if currentTime.Sub(CabOrderArray[utils.Elevators[i].ID][floor].Time) > 15*time.Second {

						fmt.Println("Elevator", utils.Elevators[i].ID, "did not complete cab order at floor", floor, ". Resending order.")

						newOrder := utils.MessageNewOrder{
							Type: "MessageNewOrder",
							NewOrder: utils.Order{
								Floor:  floor,
								Button: utils.Cab},
							FromElevator: *e,
							ToElevatorID: utils.Elevators[i].ID,
						}

						CabOrderArray[utils.Elevators[i].ID][floor].Active = false

						newOrderTx <- newOrder

					}
				}
			}
		}

	}

	for i, _ := range ordersToBeReAssigned {

		BestElevator := orders.ChooseElevator(ordersToBeReAssigned[i])

		fmt.Print("Reassigning order to elevator", BestElevator.ID)

		newOrder := utils.MessageNewOrder{
			Type:         "MessageNewOrder",
			NewOrder:     ordersToBeReAssigned[i],
			FromElevator: *e,
			ToElevatorID: BestElevator.ID,
		}

		if BestElevator.ID == e.ID {

			ProcessElevatorOrders(utils.Order{
				Floor:  ordersToBeReAssigned[i].Floor,
				Button: ordersToBeReAssigned[i].Button,
			}, orderCompleteTx, e)

			HallOrderArray[ordersToBeReAssigned[i].Button][ordersToBeReAssigned[i].Floor].Active = true

		} else {

			newOrderTx <- newOrder
			HallOrderArray[ordersToBeReAssigned[i].Button][ordersToBeReAssigned[i].Floor].Active = true

		}

		OrderActive(ordersToBeReAssigned[i], e, time.Now())

	}

	utils.OrderWatcher.HallOrderArray = HallOrderArray
	utils.OrderWatcher.CabOrderArray = CabOrderArray

}

func DetermineMaster(e *utils.Elevator) {

	// DetermineMaster determines the master elevator among the available elevators.
	// It sets the first active elevator as the initial master candidate and then compares it with other active elevators.
	// The elevator with the lowest ID becomes the master candidate.
	// Finally, it updates the local state and broadcasts the master election result if necessary.

	fmt.Println("Function: DetermineMaster")

	if len(utils.Elevators) == 0 {
		return // No elevators available
	}

	fmt.Println("Determining master")

	masterCandidate := utils.Elevators[0] // Set the first elevator as the master candidate

	for i, _ := range utils.Elevators {
		if utils.Elevators[i].IsActive {
			masterCandidate = utils.Elevators[i] // Set the first active elevator as the master candidate
			break
		}
	}

	for i, _ := range utils.Elevators {
		fmt.Println("Elevator: ", utils.Elevators[i].ID, "isActive: ", utils.Elevators[i].IsActive)
		if utils.Elevators[i].IsActive {
			if utils.Elevators[i].ID < masterCandidate.ID { // If the elevator ID is less than the master candidate ID
				masterCandidate = utils.Elevators[i] // Set the elevator as the master candidate
			}
		}
	}

	for i, _ := range utils.Elevators {
		if masterCandidate.ID != utils.Elevators[i].ID {
			if e.IsMaster {
				e.IsMaster = false
			}
		}
	}

	// Set the masterCandidate as the master and update the local state as needed
	// This is a simplified representation; actual implementation may require additional synchronization and communication
	masterCandidate.IsMaster = true
	fmt.Println("The master now is elevator: ", masterCandidate.ID)
	if masterCandidate.ID == e.ID {
		e.IsMaster = true
		fmt.Println("I am the master")
	}

	utils.MasterElevatorID = masterCandidate.ID // Set the master elevator ID

	// Broadcast or communicate the master election result as needed
	// This could involve sending a message to all elevators or updating a shared state
}

// To-Do function for handling motor stop?
// In the case of it happening, redistribute all active hall orders?
