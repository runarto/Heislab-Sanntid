package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/crash"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {NullButtons resets all the elevator buttons}
//*

func NullButtons() {

	elevio.SetStopLamp(false)
	for f := 0; f < utils.NumFloors; f++ {
		for b := 0; b < utils.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Initializes the elevator}
//*
//* @return     {Returns the initialized elevator}
//*

func InitializeElevator() utils.Elevator {
	e := crash.CheckCrashDump()
	utils.MasterID = utils.NotDefined

	const MotorLossTime = 5 * time.Second
	motorLossTimer := time.NewTimer(MotorLossTime)
	motorLossTimer.Stop()

	NullButtons()
	fmt.Println("Function: InitializeElevator")

	// Immediate check for the current floor before deciding on the motor direction
	floor := elevio.GetFloor()
	if floor != utils.NotDefined {
		// Elevator is already on a floor, no need to move
		e.CurrentFloor = floor
		e.CurrentDirection = elevio.MD_Stop
		e.CurrentState = utils.Still
		return e
	}

	// Initial direction - default is up, but this could be dynamic based on last known position
	direction := elevio.MD_Up
	maxTime := 2000 // Maximum time in milliseconds to move in one direction before switching
	elevio.SetMotorDirection(direction)
	motorLossTimer.Reset(MotorLossTime)
	startTime := time.Now()

	for floor == utils.NotDefined {
		floor = elevio.GetFloor()
		if floor != utils.NotDefined {
			break // Exit loop if a floor is detected
		}

		// Switch direction if maxTime is exceeded without finding a floor
		if time.Since(startTime).Milliseconds() > int64(maxTime) {
			if direction == elevio.MD_Up {
				direction = elevio.MD_Down
			} else {
				direction = elevio.MD_Up
			}
			elevio.SetMotorDirection(direction)
			startTime = time.Now() // Reset the start time after changing direction
		}

		// Check for motor loss
		select {
		case <-motorLossTimer.C:
			crash.Crash(e)
		default:
			time.Sleep(10 * time.Millisecond) // Sleep to avoid busy looping
		}
	}

	// Stop the elevator and set the current floor
	motorLossTimer.Stop()
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.CurrentFloor = floor
	e.CurrentDirection = elevio.MD_Stop
	e.CurrentState = utils.Still

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Sets the floor indicator light and updates the current floor of the elevator}
//*
//* @param      floor  The floor
//* @param      e      The elevator
//*
//* @return     {Returns the updated elevator state}
//*

func FloorLights(floor int, e utils.Elevator) utils.Elevator {

	// FloorLights sets the floor indicator light and updates the current floor of the elevator.
	// It takes the floor number and a pointer to the elevator as input.
	// The floor number should be between 0 and NumFloors-1.

	if floor >= 0 && floor <= utils.NumFloors-1 {
		elevio.SetFloorIndicator(floor)
		e.CurrentFloor = floor
	}

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {OrdersAbove checks if there are any orders above the current floor}
//*
//* @param      e     The elevator
//*
//* @return     {Returns true if there are any orders above the current floor, and false otherwise}
//*

func OrdersAbove(e utils.Elevator) bool {

	if e.CurrentFloor >= utils.NumFloors-1 {
		return false
	}

	for b := 0; b < utils.NumButtons; b++ {
		for f := e.CurrentFloor + 1; f < utils.NumFloors; f++ {
			if e.LocalOrderArray[b][f] {
				return true
			}
		}
	}
	return false
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {OrdersBelow checks if there are any orders below the current floor}
//*
//* @param      e     The elevator
//*
//* @return     {Returns true if there are any orders below the current floor, and false otherwise}
//*

func OrdersBelow(e utils.Elevator) bool {

	if e.CurrentFloor <= 0 {
		return false
	}

	for b := 0; b < utils.NumButtons; b++ {
		for f := 0; f < e.CurrentFloor; f++ {
			if e.LocalOrderArray[b][f] {
				return true
			}
		}
	}
	return false
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {OrderAtCurrentFloor checks if there are any orders at the current floor}
//*
//* @param      e     The elevator
//*
//* @return     {Returns true if there are any orders at the current floor, and false otherwise}
//*

func OrderAtCurrentFloor(e utils.Elevator) bool {

	for b := 0; b < utils.NumButtons; b++ {
		if e.LocalOrderArray[b][e.CurrentFloor] {
			return true
		}
	}
	return false
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {GetElevatorDirection determines the direction of the elevator based on the current state}
//*
//* @param      e     The elevator
//*
//* @return     {Returns the motor direction and the state of the elevator}
//*

func GetElevatorDirection(e utils.Elevator) (elevio.MotorDirection, utils.State) {

	fmt.Println("Function: GetElevatorDirection")
	fmt.Println("Current direction is: ", e.CurrentDirection)

	switch e.CurrentDirection {
	case utils.Up:
		if OrdersAbove(e) {
			return elevio.MD_Up, utils.Moving
		} else if OrderAtCurrentFloor(e) {
			return elevio.MD_Stop, utils.DoorOpen
		} else if OrdersBelow(e) {
			return elevio.MD_Down, utils.Moving
		} else {
			return elevio.MD_Stop, utils.Still
		}
	case utils.Down:
		if OrdersBelow(e) {
			return elevio.MD_Down, utils.Moving
		} else if OrderAtCurrentFloor(e) {
			return elevio.MD_Up, utils.Still
		} else if OrdersAbove(e) {
			return elevio.MD_Up, utils.Moving
		} else {
			return elevio.MD_Stop, utils.Still
		}
	case utils.Stopped:
		if OrderAtCurrentFloor(e) {
			return elevio.MD_Stop, utils.DoorOpen
		} else if OrdersAbove(e) {
			return elevio.MD_Up, utils.Moving
		} else if OrdersBelow(e) {
			return elevio.MD_Down, utils.Moving
		} else {
			return elevio.MD_Stop, utils.Still
		}
	default:
		return elevio.MD_Stop, utils.Still

	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {ShouldStop checks if the elevator should stop at the current floor}
//*
//* @param      e     The elevator
//*
//* @return     {Returns true if the elevator should stop at the current floor, and false otherwise}
//*

func ShouldStop(e utils.Elevator) bool {

	switch e.CurrentDirection {
	case utils.Up:
		if e.LocalOrderArray[utils.HallUp][e.CurrentFloor] ||
			e.LocalOrderArray[utils.Cab][e.CurrentFloor] ||
			!OrdersAbove(e) {
			return true
		}
		return false

	case utils.Down:
		if e.LocalOrderArray[utils.HallDown][e.CurrentFloor] ||
			e.LocalOrderArray[utils.Cab][e.CurrentFloor] ||
			!OrdersBelow(e) {
			return true
		}
		return false

	case utils.Stopped:
		return true
	default:
		return true
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {SetButtonLamp sets the button lamp for the elevator}
//*
//* @param      b      The button
//* @param      f      The floor
//* @param      value  The value to set the button lamp to
//*

func SetButtonLamp(b int, f int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(b), f, value)
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {ClearOrder clears a specific order at a specific floor for the elevator}
//*
//* @param      e     The elevator
//* @param      f     The floor
//* @param      b     The button
//*
//* @return     {Returns the updated elevator state}
//*

func ClearOrder(e utils.Elevator, f int, b int) utils.Elevator {
	e.LocalOrderArray[b][f] = false
	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {ShouldClearOrderAtFloor checks if the elevator should clear the order at the current floor}
//*
//* @param      e     The elevator
//* @param      f     The floor
//* @param      b     The button
//*
//* @return     {Returns true if the elevator should clear the order at the current floor, and false otherwise}
//*

func ShouldClearOrderAtFloor(e utils.Elevator, f int, b int) bool {

	state := e.CurrentState
	if e.CurrentFloor != f {
		return false // Elevator is not on the requested floor, cannot take the order.
	}

	// Requests from within the cab are always accepted.
	if b == utils.Cab {
		return true
	}

	// Handling hall call buttons when elevator is moving or stopped.
	switch e.CurrentDirection {
	case utils.Up:
		// Accept if the request is to move up from the current floor.
		if b == utils.HallUp {
			return true
		}
		// Special case: If there are no orders above, accept a down request, as the elevator will need to return.
		if b == utils.HallDown && !OrdersAbove(e) {
			return true
		}

	case utils.Down:
		// Accept if the request is to move down from the current floor.
		if b == utils.HallDown {
			return true
		}
		// Special case: If there are no orders below, accept an up request, as the elevator will need to return.
		if b == utils.HallUp && !OrdersBelow(e) {
			return true
		}

	case utils.Stopped:
		// If stopped, the elevator can take any hall call since it can decide its direction freely.
		if b == utils.HallUp || b == utils.HallDown {
			return true
		}
	}

	// If the elevator is idle (no orders above or below), it can accept any hall call.
	if !OrdersAbove(e) && !OrdersBelow(e) {
		return true
	}

	// Handling specific scenarios when the elevator is in 'DoorOpen' or 'Stopped' state.
	if state == utils.DoorOpen || state == utils.Stopped {
		// If there are no orders in the direction of the request, the request can be accepted.
		if !OrdersAbove(e) && e.CurrentDirection == utils.Up && b == utils.HallDown {
			return true
		} else if !OrdersBelow(e) && e.CurrentDirection == utils.Down && b == utils.HallUp {
			return true
		}
	}

	return false // If none of the above conditions are met, the elevator should not take the order.

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {SetMotorLossTimer sets the motor loss timer for the elevator, based off the current direction}
//*
//* @param      direction  The direction
//* @param      timer      The timer
//* @param      duration   The duration
//*

func SetMotorLossTimer(direction int, timer *time.Timer, duration time.Duration) {

	if direction != utils.Stopped {
		timer.Reset(duration)
	} else {
		timer.Stop()
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {ClearOrdersAtFloor clears all orders at the current floor for the elevator}
//*
//* @param      e     The elevator
//*
//* @return     {Returns the updated elevator state}
//*

func ClearOrdersAtFloor(e utils.Elevator) utils.Elevator {

	e = Clear(e, utils.Cab)

	switch e.CurrentDirection {
	case utils.Up:
		if !OrdersAbove(e) && !e.LocalOrderArray[utils.HallUp][e.CurrentFloor] {
			e = Clear(e, utils.HallDown)
		}
		e = Clear(e, utils.HallUp)

	case utils.Down:
		if !OrdersBelow(e) && !e.LocalOrderArray[utils.HallDown][e.CurrentFloor] {
			e = Clear(e, utils.HallUp)
		}
		e = Clear(e, utils.HallDown)

	case utils.Stopped:
		e = Clear(e, utils.HallDown)
		e = Clear(e, utils.HallUp)

	}

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Clear clears an orderr at the current floor for the elevator}
//*
//* @param      e     The elevator
//* @param      b     The button
//*
//* @return     {Returns the updated elevator state}
//*

func Clear(e utils.Elevator, b int) utils.Elevator {
	e.LocalOrderArray[b][e.CurrentFloor] = false
	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {SetCabLights sets the cab lights for the elevator}
//*
//* @param      e     The elevator
//*

func SetCabLights(e utils.Elevator) {
	for f := 0; f < utils.NumFloors; f++ {
		if e.LocalOrderArray[utils.Cab][f] {
			SetButtonLamp(utils.Cab, f, true)
		} else {
			SetButtonLamp(utils.Cab, f, false)
		}
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief {SetHallLights sets the hall lights for the elevator}
//*
//* @param      lights  The lights to set
//*

func SetHallLights(lights [2][utils.NumFloors]bool) {

	for f := 0; f < utils.NumFloors; f++ {
		if lights[utils.HallUp][f] {
			SetButtonLamp(utils.HallUp, f, true)
		} else {
			SetButtonLamp(utils.HallUp, f, false)
		}
		if lights[utils.HallDown][f] {
			SetButtonLamp(utils.HallDown, f, true)
		} else {
			SetButtonLamp(utils.HallDown, f, false)
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {GetHallLights gets the hall lights for the elevator}
//*
//* @param      e     The elevator
//*
//* @return     {Returns the hall lights for the elevator}
//*

func GetHallLights(e utils.Elevator) [2][utils.NumFloors]bool {

	var lights [2][utils.NumFloors]bool

	for f := 0; f < utils.NumFloors; f++ {
		lights[utils.HallUp][f] = e.LocalOrderArray[utils.HallUp][f]
		lights[utils.HallDown][f] = e.LocalOrderArray[utils.HallDown][f]
	}

	return lights
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {ExecuteOrder executes a new order for the elevator}
//*
//* @param      newOrder              The new order
//* @param      e                     The elevator
//* @param      doorTimer             The door timer
//* @param      motorLossTimer        The motor loss timer
//* @param      DoorOpenTime          The door open time
//* @param      MotorLossTime         The motor loss time
//* @param      messageHandler        Channel for sending orders to the network
//* @param      Online                Indicates if online
//* @param      OfflineOrderCompleteCh Channel for offline order completion
//*
//* @return     {Returns the updated elevator state}
//*

func ExecuteOrder(newOrder utils.Order, e utils.Elevator, doorTimer *time.Timer,
	motorLossTimer *time.Timer, DoorOpenTime time.Duration, MotorLossTime time.Duration,
	messageHandler chan utils.Message, Online bool, OfflineOrderCompleteCh chan utils.Order) utils.Elevator {

	floor := newOrder.Floor
	button := newOrder.Button

	fmt.Println("---DO ORDER RECEIVED---")

	fmt.Println("New order for FSM: ", newOrder)

	switch e.CurrentState {

	case utils.DoorOpen:

		if ShouldClearOrderAtFloor(e, floor, int(button)) {
			prev := e
			fmt.Println("Clearing order at floor: ", floor, " and button: ", button)
			e = ClearOrder(e, floor, int(button))
			CheckOrdersDone(messageHandler, e, prev, Online, OfflineOrderCompleteCh)

			doorTimer.Reset(DoorOpenTime)
		} else {
			e.LocalOrderArray[button][floor] = true
		}

	case utils.Still:

		if ShouldClearOrderAtFloor(e, floor, int(button)) {
			prev := e
			e = ClearOrder(e, floor, int(button))
			fmt.Println("Clearing order at floor: ", floor, " and button: ", button)
			CheckOrdersDone(messageHandler, e, prev, Online, OfflineOrderCompleteCh)
			doorTimer.Reset(DoorOpenTime)

		} else {
			e.LocalOrderArray[button][floor] = true
		}

		e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)

		switch e.CurrentState {
		case utils.Moving:
			fmt.Println("Moving...")
			elevio.SetMotorDirection(e.CurrentDirection)
			SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)

		case utils.Still:
			fmt.Println("Still...")
			prev := e
			e = utils.SetDoorState(utils.Open, e)
			doorTimer.Reset(DoorOpenTime)
			e = ClearOrdersAtFloor(e)
			CheckOrdersDone(messageHandler, e, prev, Online, OfflineOrderCompleteCh)
		}

	case utils.Moving:
		e.LocalOrderArray[button][floor] = true

	}

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {HandleArrivalAtFloor handles the arrival at a floor for the elevator}
//*
//* @param      floor         The floor
//* @param      e             The elevator
//* @param      motorLossTimer The motor loss timer
//* @param      doorTimer     The door timer
//* @param      DoorOpenTime  The door open time
//* @param      MotorLossTime The motor loss time
//*
//* @return     {Returns the updated elevator state}
//*

func HandleArrivalAtFloor(floor int, e utils.Elevator, motorLossTimer *time.Timer, doorTimer *time.Timer,
	DoorOpenTime time.Duration, MotorLossTime time.Duration) utils.Elevator {

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

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {DoorTimerExpired handles the door timer expiration for the elevator}
//*
//* @param      e             The elevator
//* @param      doorTimer     The door timer
//* @param      DoorOpenTime  The door open time
//* @param      motorLossTimer The motor loss timer
//* @param      MotorLossTime The motor loss time
//* @param      FloorSensorCh The floor sensor channel
//*
//* @return     {Returns the updated elevator state}
//*

func DoorTimerExpired(e utils.Elevator, doorTimer *time.Timer, DoorOpenTime time.Duration,
	motorLossTimer *time.Timer, MotorLossTime time.Duration, FloorSensorCh chan int) utils.Elevator {

	e = utils.SetDoorState(utils.Close, e)
	e = utils.SetState(utils.Still, e)
	//utils.PrintLocalOrderArray(e)
	e.CurrentDirection, e.CurrentState = GetElevatorDirection(e)

	motorLossTimer.Reset(MotorLossTime)

	if e.CurrentState == utils.DoorOpen {

		FloorSensorCh <- e.CurrentFloor

	} else {
		elevio.SetMotorDirection(e.CurrentDirection)
		SetMotorLossTimer(int(e.CurrentDirection), motorLossTimer, MotorLossTime)
	}

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Obstruction handles the obstruction for the elevator}
// *
// * @param      obstruction        Indicates if there is an obstruction
// * @param      e                  The elevator
// * @param      doorTimer          The door timer
// * @param      DoorOpenTime       The door open time
// * @param      ObstructionTimeout The obstruction timeout
// * @param      obstructionTimer   The obstruction timer
// * @param      ObstrCh            The obstruction channel
// * @param      PeerTxEnable       The peer transmission enable channel
// *
// * @return     {Returns the updated elevator state}
// *

func Obstruction(obstruction bool, e utils.Elevator, doorTimer *time.Timer, DoorOpenTime time.Duration, ObstructionTimeout time.Duration,
	obstructionTimer *time.Timer, ObstrCh <-chan bool, PeerTxEnable chan bool) utils.Elevator {

	prevDirection := e.CurrentDirection

	if obstruction {
		e = utils.Obstruction(true, e)
		doorTimer.Reset(DoorOpenTime)
		obstructionTimer.Reset(ObstructionTimeout)
		PeerTxEnable <- false
		elevio.SetMotorDirection(elevio.MD_Stop)

		for obstruction {
			select {
			case obstruction = <-ObstrCh:
				if !obstruction {
					e = utils.Obstruction(false, e)
					PeerTxEnable <- true
					fmt.Println("---OBSTRUCTION CLEARED---")
					doorTimer.Reset(DoorOpenTime)
					elevio.SetMotorDirection(prevDirection)
					return e
				}
			case <-time.After(ObstructionTimeout):
				fmt.Println("Obstruction timeout occurred.")
				crash.Crash(e)
			}
		}
	} else {
		PeerTxEnable <- true
		e = utils.Obstruction(false, e)

	}

	return e
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//* @brief      {CheckOrdersDone checks if the orders are done for the elevator}
// *
// * @param      messageHandler  Channel for sending orders to the network
// * @param      e               The elevator
// * @param      prev            The previous elevator state
// * @param      Online          Indicates if online
// * @param      OfflineOrderCompleteCh The offline order complete channel
// *

func CheckOrdersDone(messageHandler chan utils.Message, e utils.Elevator, prev utils.Elevator, Online bool, OfflineOrderCompleteCh chan utils.Order) {
	for b := 0; b < utils.NumButtons; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if !e.LocalOrderArray[b][f] && prev.LocalOrderArray[b][f] {
				if Online {
					msg := utils.PackMessage("MessageOrderComplete", utils.MasterID, utils.ID, utils.Order{Floor: f, Button: elevio.ButtonType(b)})
					messageHandler <- msg
				} else {
					OfflineOrderCompleteCh <- utils.Order{Floor: f, Button: elevio.ButtonType(b)}
				}
			}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
