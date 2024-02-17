package main

import (
    "github.com/runarto/Heislab-Sanntid/elevio"
    "time"
    "fmt"
)



func main() {

    // Initialize the elevator
    elevio.Init("localhost:15657", numFloors)
  
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
        
            myElevator.UpdateOrderSystem(newOrder) // Update the local order array
            myElevator.PrintLocalOrderSystem()
            order := myElevator.ChooseBestOrder() // Choose the best order
            fmt.Println("Best order: ", order)
            myElevator.DoOrder(order) // Move the elevator to the best order
            

        case floor := <-drv_floors:

            fmt.Println("Arrived at floor: ", floor)

            myElevator.floorLights(floor) // Update the floor lights
            
            if myElevator.ElevatorAtFloor(floor) { // Check for active orders at floor
                myElevator.StopElevator() // Stop the elevator
                myElevator.SetDoorState(Open) // Open the door
                time.Sleep(1000 * time.Millisecond) // Wait for a second
                myElevator.SetDoorState(Close) // Close the door
                fmt.Println("Ordersystem: ")
                myElevator.PrintLocalOrderSystem()
                amountOfOrders := myElevator.CheckAmountOfActiveOrders() // Check the amount of active orders
                fmt.Println("Amount of active orders: ", amountOfOrders)
                if amountOfOrders > 0 {
                    order := myElevator.ChooseBestOrder() // Choose the best order
                    myElevator.DoOrder(order)
                    // DoOrder(order) // Move the elevator to the best order (pseudocode function to move the elevator to the best order
                } else {
                    myElevator.SetState(Still) // If no orders, set the state to still
                }
            }




        case obstr := <-drv_obstr:
            myElevator.isObstruction(obstr)


        case stop := <-drv_stop:
            myElevator.StopButton(stop)
          
        }
        
    }
}
