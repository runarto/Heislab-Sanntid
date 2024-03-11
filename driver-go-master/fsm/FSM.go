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

			fmt.Println("---DO ORDER RECEIVED---")

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
					fmt.Println("Moving...")
					elevio.SetMotorDirection(e.CurrentDirection)
					SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)

				case utils.Still:
					fmt.Println("Still...")
					e = utils.SetDoorState(Open, e)
					doorTimer.Reset(DoorOpenTime)
					e = ClearOrdersAtFloor(e)
				}

			case utils.Moving:
				e.LocalOrderArray[button][floor] = true

			}

			utils.PrintLocalOrderArray(e)
			LocalStateUpdateCh <- e // Update the local elevator instance
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-FloorCh:

			fmt.Println("---ARRIVED AT FLOOR ", floor, "---")

			motorLossTimer.Reset(MotorLossTime)
			e.CurrentFloor = floor
			elevio.SetFloorIndicator(floor)

			if ShouldStop(e) {

				elevio.SetMotorDirection(elevio.MD_Stop)
				SetMotorLossTimer(int(elevio.MD_Stop), motorLossTimer, MotorLossTime)
				e = utils.SetDoorState(utils.Open, e)
				e = utils.SetState(utils.DoorOpen, e)
				e = ClearOrdersAtFloor(e)
				doorTimer.Reset(DoorOpenTime)
			}

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			fmt.Println("Current state is: ", e.CurrentState)
			utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-ObstrCh:

			fmt.Println("---OBSTRUCTION DETECTED---")

			if obstruction {
				e = utils.Obstruction(true, e)
				doorTimer.Reset(DoorOpenTime)
				PeerTxEnable <- false
			} else {
				e = utils.Obstruction(false, e)
				PeerTxEnable <- true
			}

		case <-doorTimer.C:

			fmt.Println("---DOOR TIMER EXPIRED---")

			e = utils.SetDoorState(Close, e)
			e = utils.SetState(utils.Still, e)
			utils.PrintLocalOrderArray(e)
			e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)
			fmt.Println("Current direction is: ", e.CurrentDirection, "Current state is: ", e.CurrentState)

			motorLossTimer.Reset(MotorLossTime)

			if e.CurrentState == utils.DoorOpen {

				FloorCh <- e.CurrentFloor

			} else {
				elevio.SetMotorDirection(e.CurrentDirection)
				SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)
			}

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		//case <-motorLossTimer.C:

		//PeerTxEnable <- false

		case <-lastUpdateTimer.C:
			LocalStateUpdateCh <- e
			lastUpdateTimer.Reset(TimeSinceLastUpdate)
			fmt.Println("State update timer expired... sending update")

		case update := <-IsOnlineCh:
			Online = update

		case lights := <-LocalLightsCh:
			SetHallLights(lights)

		case l := <-LightsRx:

			if !utils.Master {
				fmt.Println("Received lights from master...")
				SetHallLights(l.Lights)
			}

		}

		SetCabLights(e)

		if !Online {

			SetHallLights(GetHallLights(e))
		}

	}

}
