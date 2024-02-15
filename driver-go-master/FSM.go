package main


import (
    "fmt"
    "github.com/runarto/Heislab-Sanntid/elevio"
    "time"
)


func NullButtons() { // Turns off all buttons
    elevio.SetStopLamp(Off)
    for f := 0; f < numFloors; f++ {
        for b := 0; b < numButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, Off)
        }
    }
}


func (e *Elevator) InitElevator() {
    NullButtons()
    e.SetDoorState(Close) // Close the door
    
    for floor := elevio.GetFloor(); floor != 0; floor = elevio.GetFloor() {
        if floor > 0 || floor == -1 {
            e.GoDown()
        }
        time.Sleep(100 * time.Millisecond)
    }
    e.StopElevator()
    e.CurrentFloor = elevio.GetFloor()
    fmt.Println("Elevator is ready for use")
}



func (e *Elevator) floorLights(floor int) {
    if (floor >= 0 && floor <= 3) {
        elevio.SetFloorIndicator(floor);
        LastDefinedFloor = currentFloor
        e.CurrentFloor = floor
    }
}

func (e *Elevator) ElevatorAtFloor(floor int) bool {
    e.CurrentFloor = floor; // Update the current floor
    ordersDone := 0 // Number of orders done
    
    for i := 0; i < elevio._numButtons; i++ {
        if e.localOrderArray[i][floor] == True { // If there is an active order at the floor
            if e.CurrentDirection == Up && i == 0 {
                Order = Order{floor, HallUp}
                e.UpdateOrderSystem(Order) // Update the local order array
                ordersDone++
                // HallUp order, and the elevator is going up (take order)
                continue 
            }
            else if e.CurrentDirection == Down && i == 1 {
                Order = Order{floor, HallDown}
                e.UpdateOrderSystem(Order) // Update the local order array
                ordersDone++
                // HallDown order, and the elevator is going down (take order)
                continue 
            }
            else if i == 2 {
                Order = Order{floor, Cab}
                e.UpdateOrderSystem(Order) // Update the local order array
                ordersDone++
                // Cab order
                continue 
            }

        }  
    }
    if ordersDone > 0 {
        return true // There are active orders at the floor
    } else {
        return false // There are no active orders at the floor
    }

}