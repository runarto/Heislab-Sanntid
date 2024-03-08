package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func FSM(c *utils.Channels, e utils.Elevator) {

	Online := false

	const DoorOpenTime = 3 * time.Second
	const MotorLossTime = 5 * time.Second

	const TimeSinceLastUpdate = 5 * time.Second

	Open := utils.Open
	Close := utils.Close
	doorTimer := time.NewTimer(DoorOpenTime)
	doorTimer.Stop()

	motorLossTimer := time.NewTimer(MotorLossTime)
	motorLossTimer.Stop()

	lastUpdateTimer := time.NewTimer(TimeSinceLastUpdate)

	for {

		select {

		case newOrder := <-c.DoOrderCh:

			fmt.Println("New order to do received")

			floor := newOrder.Floor
			button := newOrder.Button

			switch e.CurrentState {

			case utils.DoorOpen:

				if ShouldClearOrderAtFloor(e, floor, int(button)) {
					e = ClearOrder(e, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					e.LocalOrderArray[floor][button] = true
				}

			case utils.Still:

				if ShouldClearOrderAtFloor(e, floor, int(button)) {
					e = ClearOrder(e, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					e.LocalOrderArray[floor][button] = true
				}

				e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)

				switch e.CurrentState {
				case utils.Moving:
					elevio.SetMotorDirection(e.CurrentDirection)
					SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)

				case utils.Still:
					e.SetDoorState(Open)
					doorTimer.Reset(DoorOpenTime)
					e = ClearOrdersAtFloor(e)
				}

			case utils.Moving:
				e.LocalOrderArray[floor][button] = true

			}

			c.LocalStateUpdateCh <- e // Update the local elevator instance
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-c.FloorCh:

			fmt.Println("Arrived at floor", floor)

			motorLossTimer.Reset(MotorLossTime)
			e.CurrentFloor = floor
			elevio.SetFloorIndicator(floor)

			if ShouldStop(e) {

				elevio.SetMotorDirection(elevio.MD_Stop)
				SetMotorLossTimer(int(elevio.MD_Stop), motorLossTimer, MotorLossTime)
				e.SetDoorState(Open)
				e = ClearOrdersAtFloor(e)
				doorTimer.Reset(DoorOpenTime)
			}

			c.LocalStateUpdateCh <- e
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-c.ObstrCh:

			if obstruction {
				e.Obstruction(true)
				doorTimer.Reset(DoorOpenTime)
				c.PeerTxEnable <- false
			} else {
				e.Obstruction(false)
				c.PeerTxEnable <- true
			}

		case <-doorTimer.C:

			fmt.Println("Door timer expired")

			e.SetDoorState(Close)
			e.CurrentState = utils.Still
			e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)
			motorLossTimer.Reset(MotorLossTime)

			if e.CurrentState == utils.DoorOpen {

				c.FloorCh <- e.CurrentFloor

			}

			c.LocalStateUpdateCh <- e
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case <-motorLossTimer.C:

			c.PeerTxEnable <- false

		case <-lastUpdateTimer.C:
			c.LocalStateUpdateCh <- e

		case update := <-c.IsOnlineCh:
			Online = update

		case lights := <-c.LocalLightsCh:
			if utils.Master {
				SetHallLights(lights)
			}

		case l := <-c.LightsRx:

			lights := l.Lights

			if l.FromElevatorID != e.ID {
				SetHallLights(lights)
				utils.CreateAndSendMessage(c, "MessageLightsConfirmed", true, e.ID)
			}

		}

		SetCabLights(e)

		if !Online {

			SetHallLights(GetHallLights(e))
		}

	}

}
