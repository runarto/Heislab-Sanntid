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
    var ordersDone []Order // Number of orders done

    
    for button := 0; button < numButtons; button++ {
        if e.localOrderArray[button][floor] == True { // If there is an active order at the floor

            if e.CurrentDirection == Up && button == HallUp {
                Order = Order{floor, HallUp}
                ordersDone = append(ordersDone, Order)
                // HallUp order, and the elevator is going up (take order)
                continue 
            } 

            if (e.CurrentDirection == Up && button == HallDown) && (e.localOrderArray[HallUp][floor] == False) {
                check := e.CheckAbove(floor)
                if check.Floor == NotDefined { // There are no orders above the current floor
                    Order = Order{floor, HallDown}
                    ordersDone = append(ordersDone, Order) // Update the local order array
                    // HallDown order, and the elevator is going up (take order)
                    continue
                }
            }

            if e.CurrentDirection == Down && button == HallDown {
                Order = Order{floor, HallDown}
                ordersDone = append(ordersDone, Order) // Update the local order array
                // HallDown order, and the elevator is going down (take order)
                continue
            }

            if (e.CurrentDirection == Down && button == HallUp) && (e.localOrderArray[HallDown][floor] == False) {
                check := e.CheckBelow(floor)
                if check.Floor == NotDefined { // There are no orders below the current floor
                    Order = Order{floor, HallUp}
                    ordersDone = append(ordersDone, Order) // Update the local order array
                    ordersDone++
                    // HallUp order, and the elevator is going down (take order)
                    continue
                }
            }

            if button == Cab {
                Order = Order{floor, Cab}
                ordersDone = append(ordersDone, Order) // Update the local order array
                // Cab order (take order)
                continue
            }


        }  
    }
    if len(ordersDone) > 0 {
        for i := 0; i < len(ordersDone); i++ {
            e.UpdateOrderSystem(ordersDone[i]) // Update the local order array
        }
        return true // There are active orders at the floor
    } else {
        return false // There are no active orders at the floor
    }

}