package fsm

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func NullButtons() {

	// NullButtons turns off all elevator buttons and the stop lamp.

	elevio.SetStopLamp(false)
	for f := 0; f < utils.NumFloors; f++ {
		for b := 0; b < utils.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func InitializeElevator() utils.Elevator {

	if utils.ID == 0 {
		utils.Master = true
		utils.MasterID = 0
	}
	NullButtons()
	fmt.Println("Function: InitializeElevator")
	floor := elevio.GetFloor()
	direction := utils.Up
	maxTime := 2000
	elevio.SetMotorDirection(elevio.MD_Up)
	startTime := time.Now()

	for floor == utils.NotDefined {
		floor = elevio.GetFloor()

		if time.Since(startTime).Milliseconds() > int64(maxTime) {
			if direction == 1 {
				elevio.SetMotorDirection(elevio.MD_Up)
				direction = -1
			} else {
				elevio.SetMotorDirection(elevio.MD_Down)
				direction = 1
			}
			startTime = time.Now()
		}
	}

	return utils.Elevator{
		CurrentFloor:     floor,
		CurrentDirection: elevio.MD_Stop,
		CurrentState:     utils.Still,
		LocalOrderArray:  [utils.NumButtons][utils.NumFloors]bool{},
		ID:               utils.ID,
	}
}

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

func OrderAtCurrentFloor(e utils.Elevator) bool {

	for b := 0; b < utils.NumButtons; b++ {
		if e.LocalOrderArray[b][e.CurrentFloor] {
			return true
		}
	}
	return false
}

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
		fallthrough
	default:
		return true
	}
}

func SetButtonLamp(b int, f int, value bool) {
	elevio.SetButtonLamp(elevio.ButtonType(b), f, value)
}

func OpenAndCloseDoor() {
	elevio.SetDoorOpenLamp(true)
	time.Sleep(utils.DoorOpenTime * time.Second)
	elevio.SetDoorOpenLamp(false)
}

func ClearOrder(e utils.Elevator, f int, b int) utils.Elevator {
	e.LocalOrderArray[b][f] = false
	return e
}

func ShouldClearOrderAtFloor(e utils.Elevator, f int, b int) bool {

	return e.CurrentFloor == f && ((e.CurrentDirection == utils.Up && b == utils.HallUp) ||
		(e.CurrentDirection == utils.Down && b == utils.HallDown) ||
		(e.CurrentDirection == utils.Stopped || b == utils.Cab))
}

func SetMotorLossTimer(direction int, timer *time.Timer, duration time.Duration) {

	if direction != utils.Stopped {
		timer.Reset(duration)
	} else {
		timer.Stop()
	}
}

func ClearOrdersAtFloor(e utils.Elevator) utils.Elevator {

	e = ClearFloor(e, utils.Cab)

	switch e.CurrentDirection {
	case utils.Up:
		if !OrdersAbove(e) && !e.LocalOrderArray[utils.HallUp][e.CurrentFloor] {
			e = ClearFloor(e, utils.HallDown)
		}
		e = ClearFloor(e, utils.HallUp)

	case utils.Down:
		if !OrdersBelow(e) && !e.LocalOrderArray[utils.HallDown][e.CurrentFloor] {
			e = ClearFloor(e, utils.HallUp)
		}
		e = ClearFloor(e, utils.HallDown)

	case utils.Stopped:
		e = ClearFloor(e, utils.HallDown)
		e = ClearFloor(e, utils.HallUp)

	}

	return e
}

func ClearFloor(e utils.Elevator, b int) utils.Elevator {
	e.LocalOrderArray[b][e.CurrentFloor] = false
	return e
}

func SetCabLights(e utils.Elevator) {
	for f := 0; f < utils.NumFloors; f++ {
		if e.LocalOrderArray[utils.Cab][f] {
			SetButtonLamp(utils.Cab, f, true)
		} else {
			SetButtonLamp(utils.Cab, f, false)
		}
	}

}

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

func GetHallLights(e utils.Elevator) [2][utils.NumFloors]bool {

	var lights [2][utils.NumFloors]bool

	for f := 0; f < utils.NumFloors; f++ {
		lights[utils.HallUp][f] = e.LocalOrderArray[utils.HallUp][f]
		lights[utils.HallDown][f] = e.LocalOrderArray[utils.HallDown][f]
	}

	return lights
}
