package main

import (
    "github.com/runarto/Heislab-Sanntid/elevio"
    "fmt"
)



func main() {

    // Initialize the elevator
    elevio.Init("localhost:15658", numFloors)
  
    var myElevator Elevator = Elevator{
        CurrentState:     Still, // Assuming Still is a defined constant in the State type
        CurrentDirection: elevio.MD_Stop, // Example, use a valid value from elevio.MotorDirection
        CurrentFloor:     elevio.GetFloor(), // Starts at floor 0
        doorOpen:         false, // Door starts closed
        Obstruction:      false, // No obstruction initially
        stopButton:       false, // Stop button not pressed initially
        LocalOrderArray:  [3][numFloors]int{}, // Initialize with zero values
    }

    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator

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

            floor := btn.Floor
            button := btn.Button
            newOrder := Order{floor, button}
            fmt.Println("New order: ", newOrder)

            if myElevator.CheckIfOrderIsActive(newOrder) { // Check if the order is active
                if bestOrder.Floor == myElevator.CurrentFloor {
                    myElevator.HandleElevatorAtFloor(bestOrder.Floor) // Handle the elevator at the floor
                } else {
                    myElevator.DoOrder(bestOrder) // Move the elevator to the best order
                }
                
            } else {
                myElevator.UpdateOrderSystem(newOrder) // Update the local order array
                myElevator.PrintLocalOrderSystem()
                bestOrder = myElevator.ChooseBestOrder() // Choose the best order
                fmt.Println("Best order: ", bestOrder)
    
                if bestOrder.Floor == myElevator.CurrentFloor {
                    myElevator.HandleElevatorAtFloor(bestOrder.Floor) // Handle the elevator at the floor
                } else {
                    myElevator.DoOrder(bestOrder) // Move the elevator to the best order
                }

            }
            

        case floor := <-drv_floors:

            fmt.Println("Arrived at floor: ", floor)

            myElevator.floorLights(floor) // Update the floor lights
            myElevator.HandleElevatorAtFloor(floor) // Handle the elevator at the floor




        case obstr := <-drv_obstr:
            myElevator.isObstruction(obstr)


        case stop := <-drv_stop:
            myElevator.StopButton(stop)
          
        }
        
    }
}
