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
        isMaster:         false, // Not master initially
        ElevatorIP:       "localhost:20000", // Set to the IP of the elevator
        ElevatorID:       0, // Set to the ID of the elevator
    }

    // Assumption: All elevators know each others IP at start-up
    // Need a general function for deciding which elevator is master. 
    // This function should be called by all elevators at start-up.
    // Idea: All elevators broadcast their Elevator-instance to all other elevators.
    // Each elevator compares the n values with their own, and the elevator with the highest value is master.
    // There likely is not a need to confirm this by sending a message to all other elevators,
    // because each elevator will have the same result.



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
    drv_OrderComplete := make(chan MessageOrderComplete)
    drv_NewOrder := make(chan MessageNewOrder)
    drv_watchdog := make(chan bool)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    go incrementCounter(drv_watchdog) // Only for slave
    myElevator.BroadcastElevatorInstance()

    // for address, conn := range connections {

    //     go HandleMessage(conn, drv_OrderComplete, drv_NewOrder) // Only for master

    // }

    // if myElevator.isMaster { 
    //    Message := MessageGlobalOrder{globalOrderSystem}
    //    globalOrdersSys := Message.Serialize()
    //    go BroadcastGlobalOrderSystem(globalOrdersSys) }

    // while len(Elevator != numOfElevators) {
    //    wait
    //}

    // Main event loop
    for {
        select {

        case <-drv_watchdog:

            // Check if master is still alive
            // If master is not alive, elect new master
            // ElectNewMaster()
            // If master is alive, start a new timer
            // go incrementCounter(drv_watchdog)

        case newOrder := <-drv_NewOrder:

            fmt.Println("New order: ", newOrder)

            // Update global order system
            // Check if cab or hall order
            // If hall, determine best elevator for order
            // Send order to best elevator


        case orderComplete := <-drv_OrderComplete:
    
            fmt.Println("Order complete: ", orderComplete)
            // Update global order system

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

                //newOrderToSend := MessageNewOrder{newOrder, myElevator} // Create a new order message
                //SendOrder(masterAddress, newOrderToSend) // Send the order to master

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
