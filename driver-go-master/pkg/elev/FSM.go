package elev

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func FSM(channels *utils.Channels, thisElevator *utils.Elevator) {

	for {
		select {

		case btn := <-channels.ButtonCh:

			fmt.Println("---BUTTON PRESSED---")

			floor := btn.Floor
			button := btn.Button

			newOrder := utils.Order{
				Floor:  floor,
				Button: button}
			fmt.Println("New local order: ", newOrder)

			HandleButtonEvent(newOrder, thisElevator, channels)

		case floor := <-channels.FloorCh:

			fmt.Println("---ARRIVED AT NEW FLOOR---")

			fmt.Println("Arrived at floor: ", floor)

			FloorLights(floor, thisElevator)                     // Update the floor lights
			HandleElevatorAtFloor(floor, channels, thisElevator) // Handle the elevator at the floor

		case obstr := <-channels.ObstrCh:
			thisElevator.Obstruction(obstr)

		case stop := <-channels.StopCh:
			thisElevator.StopBtnPressed(stop)
			//StopButton(stop)

		}
	}

}
