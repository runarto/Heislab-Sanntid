package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/crash"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

const bufferSize = 1000

func FSM(e utils.Elevator, DoOrderCh <-chan utils.Order, LocalElevatorStateUpdateCh chan utils.Elevator, PeerTxEnable chan bool,
	IsOnlineCh <-chan bool, SetLights chan [2][utils.NumFloors]bool, messageHandler chan utils.Message, OfflineOrderCompleteCh chan utils.Order) {

	Online := false
	var prev utils.Elevator

	const ObstructionTimeout = 15 * time.Second
	const DoorOpenTime = 3 * time.Second
	const MotorLossTime = 5 * time.Second

	const TimeSinceLastUpdate = 5 * time.Second

	obstructionTimer := time.NewTimer(ObstructionTimeout)
	obstructionTimer.Stop()

	doorTimer := time.NewTimer(DoorOpenTime)
	doorTimer.Stop()

	motorLossTimer := time.NewTimer(MotorLossTime)
	motorLossTimer.Stop()

	lastUpdateTimer := time.NewTimer(TimeSinceLastUpdate)

	FloorSensorCh := make(chan int, bufferSize)
	ObstrCh := make(chan bool, bufferSize)
	StopCh := make(chan bool, bufferSize)

	go elevio.PollFloorSensor(FloorSensorCh)
	go elevio.PollObstructionSwitch(ObstrCh)
	go elevio.PollStopButton(StopCh)

	for {

		select {

		case newOrder := <-DoOrderCh:

			if e.LocalOrderArray[newOrder.Button][newOrder.Floor] {
				continue
			}

			prev = e

			e = ExecuteOrder(newOrder, e, doorTimer, motorLossTimer, DoorOpenTime, MotorLossTime,
				messageHandler, Online, OfflineOrderCompleteCh)

			crash.SaveCabOrders(e)
			LocalElevatorStateUpdateCh <- e // Update the local elevator instance
			CheckOrdersDone(messageHandler, e, prev, Online, OfflineOrderCompleteCh)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case floor := <-FloorSensorCh:

			fmt.Println("---ARRIVED AT FLOOR ", floor, "---")

			prev = e

			e = HandleArrivalAtFloor(floor, e, motorLossTimer, doorTimer, DoorOpenTime, MotorLossTime)

			LocalElevatorStateUpdateCh <- e
			crash.SaveCabOrders(e)
			CheckOrdersDone(messageHandler, e, prev, Online, OfflineOrderCompleteCh)
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case obstruction := <-ObstrCh:

			fmt.Println("---OBSTRUCTION DETECTED---")

			e = Obstruction(obstruction, e, doorTimer, DoorOpenTime, ObstructionTimeout, obstructionTimer, ObstrCh, PeerTxEnable)
			LocalElevatorStateUpdateCh <- e

		case <-doorTimer.C:

			fmt.Println("---DOOR TIMER EXPIRED---")

			e = DoorTimerExpired(e, doorTimer, DoorOpenTime, motorLossTimer, MotorLossTime, FloorSensorCh)

			LocalElevatorStateUpdateCh <- e
			fmt.Println("Local state update sent...")
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case <-motorLossTimer.C:
			PeerTxEnable <- false
			crash.Crash(e)

		case <-lastUpdateTimer.C:
			LocalElevatorStateUpdateCh <- e
			lastUpdateTimer.Reset(TimeSinceLastUpdate)

		case update := <-IsOnlineCh:
			Online = update
			fmt.Println("Online status updated: ", update)

		case lights := <-SetLights:

			SetHallLights(lights)

		}

		SetCabLights(e)
		LocalElevatorStateUpdateCh <- e

		if !Online {

			SetHallLights(GetHallLights(e))

		}

	}

}
