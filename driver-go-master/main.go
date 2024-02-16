package main

import (
    "fmt"
    "github.com/runarto/Heislab-Sanntid/elevio"
    "sync"
)



func main() {

    numFloors := 4

    // Initialize the elevator
    elevio.Init("localhost:15657", numFloors)
  
    var myElevator Elevator = Elevator {
        CurrentState:     Still, 
        CurrentDirection: elevio.MD_Stop,
        doorOpen:         false,
        Obstruction:      false,
        stopButton:       false,
        ActiveOrders:     []activeOrder{}, 
        NetworkAdress:    "192.168.1.100",
        IsMaster:         true,
    }

    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator
    myelevator.NullButtons() // Initialize the master

    // Create channels for handling events
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    // Main event loop
    for {
        select {
        case btn := <-drv_buttons:

            myElevator.UpdateOrderSystem(btn) // Update the local order array
            order := myElevator.ChooseBestOrder() // Choose the best order
            MyElevator.DoOrder(order) // Move the elevator to the best order
            

        case floor := <-drv_floors:

            myElevator.floorLights(floor) // Update the floor lights
            
            if myElevator.ElevatorAtFloor(floor) { // Check for active orders at floor
                myElevator.StopElevator() // Stop the elevator
                myElevator.SetDoorState(Open) // Open the door
                time.sleep(1000 * time.Millisecond) // Wait for a second
                myElevator.SetDoorState(Close) // Close the door
                if myElevator.CheckAmountOfActiveOrders() > 0 {
                    myElevator.ChooseBestOrder() // Choose the best order
                    // DoOrder(order) // Move the elevator to the best order (pseudocode function to move the elevator to the best order
                } else {
                    myElevator.SetState(Still) // If no orders, set the state to still
                }
            }




        case obstr := <-drv_obstr:
            while obstr {
                e.SetDoorState(Open) // Open the door
            }
            e.SetDoorState(Close) // Close the door


        case stop := <-drv_stop:
            if stop {
                e.StopElevator() // Stop the elevator
            }
          
        }
        
    }
}
