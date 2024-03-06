package elev

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func GlobalOrdersUpdate(channels *utils.Channels, thisElevator *utils.Elevator) {

	for {
		select {

		case elevatorStatus := <-channels.ElevatorStatusRx:

			fmt.Println("---ELEVATOR STATUS RECEIVED---")

			HandleNewElevatorStatus(elevatorStatus, &elevatorStatus.FromElevator, channels.GlobalUpdateCh)

		case GlobalUpdate := <-channels.GlobalUpdateCh:

			fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")

			HandleNewGlobalOrderUpdate(GlobalUpdate, thisElevator)

		case OrderArrayUpdate := <-channels.OrderArraysRx:

			fmt.Println("---ORDER ARRAY UPDATE RECEIVED---")

			HandleNewOrderArrayUpdate(OrderArrayUpdate, thisElevator)

		}
	}
}

func HandleNewGlobalOrderUpdate(GlobalUpdate utils.GlobalOrderUpdate, thisElevator *utils.Elevator) {

	for i, _ := range GlobalUpdate.Orders {

		if GlobalUpdate.FromElevatorID == thisElevator.ID {

			if GlobalUpdate.IsComplete {

				orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], thisElevator.ID, false)
				OrderCompleted(GlobalUpdate.Orders[i], thisElevator.ID)

			} else {

				if !orders.CheckIfGlobalOrderIsActive(GlobalUpdate.Orders[i], thisElevator.ID) {

					orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], thisElevator.ID, true)
					OrderActive(GlobalUpdate.Orders[i], thisElevator.ID, time.Now())
				}

			}

		} else if GlobalUpdate.FromElevatorID != thisElevator.ID {

			if GlobalUpdate.IsComplete {

				orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID, false)
				OrderCompleted(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID)

			} else {

				if !orders.CheckIfGlobalOrderIsActive(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID) {

					orders.UpdateGlobalOrderSystem(GlobalUpdate.Orders[i], GlobalUpdate.FromElevatorID, true)
					OrderActive(GlobalUpdate.Orders[i], thisElevator.ID, time.Now())
				}

			}

		}
	}
}

func HandleOrderArrayUpdate(OrderArrayUpdate utils.MessageOrderArrays, e *utils.Elevator) {

	if OrderArrayUpdate.ToElevatorID == utils.NotDefined { // This is meant for everyone. Hence, we only want to update the global order system.

	} else if OrderArrayUpdate.ToElevatorID == e.ID { // This is meant for me. Hence, I want to update my local order array, as well as the global order system.

	}

}
