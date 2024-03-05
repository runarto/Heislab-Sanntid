package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func OrderCompleted(order utils.Order, e *utils.Elevator) {
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

func OrderActive(order utils.Order, e *utils.Elevator) {
	button := order.Button
	floor := order.Floor

	if button == utils.Cab {
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Active = true
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Completed = false
		utils.OrderWatcher.CabOrderArray[e.ID][floor].Time = time.Now()

	} else {
		utils.OrderWatcher.HallOrderArray[button][floor].Active = true
		utils.OrderWatcher.HallOrderArray[button][floor].Completed = false
		utils.OrderWatcher.HallOrderArray[button][floor].Time = time.Now()
	}
}

func CheckIfOrderIsComplete(e *utils.Elevator, newOrderTx chan utils.MessageNewOrder) {
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

			orders.UpdateLocalOrderSystem(ordersToBeReAssigned[i], e)

		} else {

			newOrderTx <- newOrder

		}

		OrderActive(ordersToBeReAssigned[i], e)

	}

	utils.OrderWatcher.HallOrderArray = HallOrderArray

}

func DetermineMaster(e *utils.Elevator) {
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
