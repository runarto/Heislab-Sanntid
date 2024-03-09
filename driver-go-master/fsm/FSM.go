package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func FSM(e utils.Elevator, DoOrderCh <-chan utils.Order, FloorCh chan int,
	ObstrCh chan bool, LocalStateUpdateCh chan utils.Elevator, PeerTxEnable chan bool,
	IsOnlineCh <-chan bool, LocalLightsCh <-chan [utils.NumButtons - 1][utils.NumFloors]bool,
	LightsRx <-chan utils.MessageLights, ch chan interface{}) {

	Online := true

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

		case newOrder := <-DoOrderCh:

			fmt.Println("New order to do received")

			floor := newOrder.Floor
			button := newOrder.Button

			fmt.Println("Current state is: ", e.CurrentState)
			switch e.CurrentState {

			case utils.DoorOpen:

				if ShouldClearOrderAtFloor(e, floor, int(button)) {
					e = ClearOrder(e, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					e.LocalOrderArray[button][floor] = true
				}

			case utils.Still:

				if ShouldClearOrderAtFloor(e, floor, int(button)) {
					e = ClearOrder(e, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					e.LocalOrderArray[button][floor] = true
				}

				e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)
				fmt.Println("Current direction is: ", e.CurrentDirection, "Current state is: ", e.CurrentState)

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
				e.LocalOrderArray[button][floor] = true

			}

			LocalStateUpdateCh <- e // Update the local elevator instance
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-FloorCh:

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

				e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)
				fmt.Println("Current direction is: ", e.CurrentDirection, "Current state is: ", e.CurrentState)

				if e.CurrentState == utils.Still {
					e.SetDoorState(Open)
					e = ClearOrdersAtFloor(e)
					doorTimer.Reset(DoorOpenTime)
				} else {
					elevio.SetMotorDirection(e.CurrentDirection)
					SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)
				}

			}

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-ObstrCh:

			if obstruction {
				e.Obstruction(true)
				doorTimer.Reset(DoorOpenTime)
				PeerTxEnable <- false
			} else {
				e.Obstruction(false)
				PeerTxEnable <- true
			}

		case <-doorTimer.C:

			fmt.Println("Door timer expired...")

			e.SetDoorState(Close)
			e.SetState(utils.Still)
			e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)
			fmt.Println("Current direction is: ", e.CurrentDirection, "Current state is: ", e.CurrentState)

			motorLossTimer.Reset(MotorLossTime)

			if e.CurrentState == utils.DoorOpen {

				FloorCh <- e.CurrentFloor

			}

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		//case <-motorLossTimer.C:

		//PeerTxEnable <- false

		case <-lastUpdateTimer.C:
			LocalStateUpdateCh <- e
			fmt.Println("State update timer expired... sending update")

		case update := <-IsOnlineCh:
			Online = update

		case lights := <-LocalLightsCh:
			SetHallLights(lights)

		case l := <-LightsRx:

			lights := l.Lights

			if l.FromElevatorID != e.ID {
				SetHallLights(lights)
				msg := utils.PackMessage("MessageLightsAck", true, e.ID)
				ch <- msg
			}

		}

		SetCabLights(e)

		if !Online {

			SetHallLights(GetHallLights(e))
		}

	}

}
