package main


import (
    "fmt"
    "Heislab-Sanntid/elevio"
    "time"
)


func nullButtons() {
    elevio.SetStopLamp(Off)
    for f := 0; f < numFloors; f++ {
        for b := 0; b < numButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, Off)

        }
    }
}


func initElevator() {
    //nullButtons()
    elevatorDoorState(Close)
    
    for floor := elevio.GetFloor(); floor != 0; floor = elevio.GetFloor() {
        if floor > 0 || floor == -1 {
            moveElevator(elevio.MD_Down)
            CurrentDirection = elevio.MD_Down
        }
        time.Sleep(100 * time.Millisecond)
    }
    elevatorStill()
    currentFloor = elevio.GetFloor()
    LastDefinedFloor = NotDefined
    fmt.Println("Elevator is ready for use")
}


func elevatorDoorState(state int) {
    if state == Open {
        elevio.SetDoorOpenLamp(On)
        IsDoorOpen = Open
    } else {
        elevio.SetDoorOpenLamp(Off)
        IsDoorOpen = Close
    }

}

func stopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
    nullButtons()
    if elevio.GetFloor() != -1 {
        elevatorDoorState(Open)
    }
    CurrentState = Stop
}

func elevatorStill() {
	elevio.SetMotorDirection(elevio.MD_Stop)
    CurrentState = Still
}


func floorLights(floor int) {
    if (floor >= 0 && floor <= 3) {
        elevio.SetFloorIndicator(floor);
        LastDefinedFloor = currentFloor
        currentFloor = floor
    }
    
}

func moveElevator(Direction elevio.MotorDirection) {
    if Direction == elevio.MD_Down {
        elevio.SetMotorDirection(elevio.MD_Down)
        CurrentDirection = ElevDown
        CurrentDirectionAlt = Down
        CurrentState = Moving
    } else if Direction == elevio.MD_Up {
        elevio.SetMotorDirection(elevio.MD_Up)
        CurrentDirection = ElevUp
        CurrentDirectionAlt = Up
        CurrentState = Moving
    } else {
        elevio.SetMotorDirection(elevio.MD_Stop)
        CurrentState = Still
    }
}


func elevatorAtFloor() {
    elevatorStill()
    elevatorDoorState(Open)
    
    go func() {
        time.Sleep(3 * time.Second) // Non-blocking sleep in a separate goroutine
        elevatorDoorState(Close)    // Close the door after the sleep period
        moveElevator(elevatorDirection()) // Determine and move to the next direction
    }()
}
