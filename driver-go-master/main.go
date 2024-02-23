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
        CurrentDirection: elevio.MD_Stop,
        GeneralDirection // Example, use a valid value from elevio.MotorDirection
        CurrentFloor:     elevio.GetFloor(), // Starts at floor 0
        doorOpen:         false, // Door starts closed
        Obstruction:      false, // No obstruction initially
        stopButton:       false, // Stop button not pressed initially
        LocalOrderArray:  [3][numFloors]int{}, // Initialize with zero values
        isMaster:         false, // Not master initially
        ElevatorIP:       "localhost"+_ListeningPort, // Set to the IP of the elevator
        ElevatorID:       0, // Set to the ID of the elevator
        isActive:         true, // Elevator is active initially
    }


    if err != nil {
        fmt.Println("Error setting up broadcast listener: ", err)
    }
    conn, err := SetUpBroadcastListener() // Set up the broadcast listener
    drv_Message := make(chan Data)

    time.Sleep(3 * time.Second) // Wait for 3 seconds

    go HandleMessages(conn, drv_NewOrder, drv_OrderComplete) // Start handling messages in a separate goroutine
    msg := MessageElevator{myElevator} // Create a new elevator instance message
    SendMessage(_broadcastAddr, msg) // Broadcast the elevator instance message 



    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator
   


    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)


    for {
        select {

        case msg := <-drv_Message:
            messageType := msg.Message[0]
            messageBytes := msg.Message[1:]
            conn = msg.Address

            myElevator.MessageType(messageType, messageBytes, conn) // Handle the message type

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
