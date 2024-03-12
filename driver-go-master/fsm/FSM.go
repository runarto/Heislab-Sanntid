package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/crash"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func FSM(e utils.Elevator, DoOrderCh <-chan utils.Order, FloorCh chan int,
	ObstrCh chan bool, LocalStateUpdateCh chan utils.Elevator, PeerTxEnable chan bool,
	IsOnlineCh <-chan bool, SetLights <-chan [2][utils.NumFloors]bool, ch chan interface{}) {

	Online := true

	const ObstructionTimeout = 5 * time.Second
	const DoorOpenTime = 3 * time.Second
	const MotorLossTime = 5 * time.Second
	const LightsTimeout = 5 * time.Second

	const TimeSinceLastUpdate = 5 * time.Second

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

			e = ExecuteOrder(newOrder, e, doorTimer, motorLossTimer, DoorOpenTime, MotorLossTime)

			//utils.PrintLocalOrderArray(e)
			LocalStateUpdateCh <- e // Update the local elevator instance
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-FloorCh:

			fmt.Println("---ARRIVED AT FLOOR ", floor, "---")

			e = HandleArrivalAtFloor(floor, e, motorLossTimer, doorTimer, DoorOpenTime, MotorLossTime)

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			fmt.Println("Current state is: ", e.CurrentState)
			//utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-ObstrCh:

			fmt.Println("---OBSTRUCTION DETECTED---")

			e = Obstruction(obstruction, e, doorTimer, DoorOpenTime, ObstructionTimeout, obstructionTimer, ObstrCh, PeerTxEnable)
			LocalStateUpdateCh <- e

		case <-doorTimer.C:

			fmt.Println("---DOOR TIMER EXPIRED---")

			e = DoorTimerExpired(e, doorTimer, DoorOpenTime, motorLossTimer, MotorLossTime, FloorCh)

			LocalStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			//utils.PrintLocalOrderArray(e)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case <-motorLossTimer.C:
			PeerTxEnable <- false
			crash.Crash(e)

		case <-lastUpdateTimer.C:
			LocalStateUpdateCh <- e
			lastUpdateTimer.Reset(TimeSinceLastUpdate)
			//fmt.Println("State update timer expired... sending update")

		case update := <-IsOnlineCh:
			Online = update

		case lights := <-SetLights:

			fmt.Println("Received lights from master...")
			SetHallLights(lights)
			lightsTimer.Reset(LightsTimeout)

		case <-lightsTimer.C:
			SetHallLights(GetHallLights(e))
			lightsTimer.Reset(LightsTimeout)

		}

		SetCabLights(e)

		if !Online {

			SetHallLights(GetHallLights(e))
		}

	}

}
