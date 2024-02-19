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
        isMaster:         true, // Not master initially
        ElevatorIP:       "localhost:20000", // Set to the IP of the elevator
        ElevatorID:       0 // Set to the ID of the elevator
    }
    // Food for thought: Cannot initialize as master. Must be elected master by the other elevators.
    // If the elevator is initialized as master and goes offline, another elevator must be elected master.
    // However if it comes back online, it would be initialized as master again, hence we would have two masters. 
    // This is a problem. We need to solve this.
    // Hence, we need to broadcast a message to each of the elevators, figure out which one is master, and let the elevators know. 




    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator

    // Check if elevator is initialized as master or slave
    //  if myElevator.CheckIfMaster()

    // If it master, broadcast message to other elevators letting them know that.
    // Message should contain the master's IP address, and the port number it is listening on.
    // myElevator.BroadcastMaster() 

    // Create channels for handling events
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)
    drv_UDP := make(chan Order)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)
    go ReadOrder(_ListeningPort, drv_UDP) // Read from the UDP port

    // if myElevator.isMaster { 
    //    Message = MessageGlobalOrder{globalOrderSystem}
    //    globalOrdersSys := Message.Serialize()
    //    go BroadcastGlobalOrderSystem(globalOrdersSys) }

    // Main event loop
    for {
        select {

        case newOrder := <-drv_UDP:
            fmt.Println("New order: ", newOrder)

            // check if master, if not, update local order system

            // if master, update global order system
            // if order is a HallOrder, pull local order systems from all elevators
            // Choose best elevator for order
            // Send order to best elevator 




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
                // if myElevator.isMaster -> update global order system locally
                // else, send order to master

                // SendOrder(address, newOrder) // Send the order to master

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
