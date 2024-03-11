package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/crash"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func FSM(e utils.Elevator, DoOrderCh <-chan utils.Order, FloorCh chan int,
	ObstrCh chan bool, LocalStateUpdateCh chan utils.Elevator, PeerTxEnable chan bool,
	IsOnlineCh <-chan bool, LocalLightsCh <-chan [utils.NumButtons - 1][utils.NumFloors]bool,
	LightsRx <-chan utils.MessageLights, ch chan interface{}) {

	Online := true
	this := &e

	const ObstructionTimeout = 5 * time.Second
	const DoorOpenTime = 3 * time.Second
	const MotorLossTime = 5 * time.Second
	const LightsTimeout = 5 * time.Second

	const TimeSinceLastUpdate = 5 * time.Second

	Open := utils.Open
	Close := utils.Close

	lightsTimer := time.NewTimer(LightsTimeout)
	lightsTimer.Stop()

	obstructionTimer := time.NewTimer(ObstructionTimeout)
	obstructionTimer.Stop()

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

			fmt.Println("Current state is: ", this.CurrentState)
			switch e.CurrentState {

			case utils.DoorOpen:

				if ShouldClearOrderAtFloor(*this, floor, int(button)) {
					*this = ClearOrder(*this, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					this.LocalOrderArray[button][floor] = true
				}

			case utils.Still:

				if ShouldClearOrderAtFloor(*this, floor, int(button)) {
					*this = ClearOrder(*this, floor, int(button))
					doorTimer.Reset(DoorOpenTime)
				} else {
					this.LocalOrderArray[button][floor] = true
				}

				this.CurrentDirection, this.CurrentState = GetElevatorDirection(*this)
				fmt.Println("Current direction is: ", this.CurrentDirection, "Current state is: ", this.CurrentState)

				switch this.CurrentState {
				case utils.Moving:
					fmt.Println("Moving...")
					elevio.SetMotorDirection(this.CurrentDirection)
					SetMotorLossTimer(int(this.CurrentDirection), motorLossTimer, MotorLossTime)

				case utils.Still:
					fmt.Println("Still...")
					*this = utils.SetDoorState(Open, *this)
					doorTimer.Reset(DoorOpenTime)
					*this = ClearOrdersAtFloor(*this)
				}

			case utils.Moving:
				this.LocalOrderArray[button][floor] = true

			}

			//utils.PrintLocalOrderArray(e)
			LocalStateUpdateCh <- *this // Update the local elevator instance
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-FloorCh:

			fmt.Println("---ARRIVED AT FLOOR ", floor, "---")

			motorLossTimer.Reset(MotorLossTime)
			e.CurrentFloor = floor
			elevio.SetFloorIndicator(floor)

			if ShouldStop(*this) {

				elevio.SetMotorDirection(elevio.MD_Stop)
				SetMotorLossTimer(int(elevio.MD_Stop), motorLossTimer, MotorLossTime)
				*this = utils.SetDoorState(utils.Open, *this)
				*this = utils.SetState(utils.DoorOpen, *this)
				*this = ClearOrdersAtFloor(*this)
				doorTimer.Reset(DoorOpenTime)
			}

			LocalStateUpdateCh <- *this
			fmt.Println("Local state update sent...")
			fmt.Println("Current state is: ", this.CurrentState)
			//utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-ObstrCh:

			fmt.Println("---OBSTRUCTION DETECTED---")

			if obstruction {
				*this = utils.Obstruction(true, *this)
				doorTimer.Reset(DoorOpenTime)
				obstructionTimer.Reset(ObstructionTimeout)
				PeerTxEnable <- false

				for obstruction {
					select {
					case obstruction = <-ObstrCh:
						if !obstruction {
							*this = utils.Obstruction(false, *this)
							PeerTxEnable <- true
							fmt.Println("---OBSTRUCTION CLEARED---")
							doorTimer.Reset(DoorOpenTime)
							break
						}
					case <-time.After(ObstructionTimeout):
						fmt.Println("Obstruction timeout occurred.")
						crash.Crash(*this)
					}
				}
			} else {
				PeerTxEnable <- true
			}

		case <-doorTimer.C:

			fmt.Println("---DOOR TIMER EXPIRED---")

			*this = utils.SetDoorState(Close, *this)
			*this = utils.SetState(utils.Still, *this)
			//utils.PrintLocalOrderArray(e)
			this.CurrentDirection, this.CurrentState = GetElevatorDirection(*this)
			fmt.Println("Current direction is: ", this.CurrentDirection, "Current state is: ", this.CurrentState)

			motorLossTimer.Reset(MotorLossTime)

			if this.CurrentState == utils.DoorOpen {

				FloorCh <- this.CurrentFloor

			} else {
				elevio.SetMotorDirection(this.CurrentDirection)
				SetMotorLossTimer(int(this.CurrentDirection), motorLossTimer, MotorLossTime)
			}

			LocalStateUpdateCh <- *this
			fmt.Println("Local state update sent...")
			//utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case <-motorLossTimer.C:
			PeerTxEnable <- false
			crash.Crash(*this)

		case <-lastUpdateTimer.C:
			LocalStateUpdateCh <- *this
			lastUpdateTimer.Reset(TimeSinceLastUpdate)
			//fmt.Println("State update timer expired... sending update")

		case update := <-IsOnlineCh:
			Online = update

		case lights := <-LocalLightsCh:
			SetHallLights(lights)
			lightsTimer.Reset(LightsTimeout)

		case l := <-LightsRx:

			if !utils.Master {
				fmt.Println("Received lights from master...")
				SetHallLights(l.Lights)
				lightsTimer.Reset(LightsTimeout)
			}

		case <-lightsTimer.C:
			SetHallLights(GetHallLights(*this))
			lightsTimer.Reset(LightsTimeout)

		}

		SetCabLights(*this)

		if !Online {

			SetHallLights(GetHallLights(*this))
		}

	}

}
