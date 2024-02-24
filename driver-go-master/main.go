package main

import (
    "github.com/runarto/Heislab-Sanntid/elevio"
    "github.com/runarto/Heislab-Sanntid/network/bcast"
    "github.com/runarto/Heislab-Sanntid/network/peers"
    "fmt"
    "strconv"
)



func main() {

    // Initialize the elevator
    elevio.Init("localhost:15658", numFloors)
  
    var myElevator Elevator = Elevator{
        CurrentState:     Still,            // Assuming Still is a defined constant in the State type
        CurrentDirection: elevio.MD_Stop,
        GeneralDirection: Stopped,            // Example, use a valid value from elevio.MotorDirection
        CurrentFloor:     elevio.GetFloor(), // Starts at floor 0
        doorOpen:         false,               // Door starts closed
        Obstruction:      false,                // No obstruction initially
        stopButton:       false,                // Stop button not pressed initially
        LocalOrderArray:  [3][numFloors]int{},  // Initialize with zero values
        isMaster:         false,                    // Not master initially
        ID:       0,                        // Set to the ID of the elevator
        isActive:         true,                     // Elevator is active initially
    }


    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator

    peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
    go peers.Transmitter(_ListeningPort, strconv.Itoa(myElevator.ID), peerTxEnable)
	go peers.Receiver(_ListeningPort, peerUpdateCh)

    newOrderTx := make(chan MessageNewOrder)
	newOrderRx := make(chan MessageNewOrder)

	orderCompleteTx := make(chan MessageOrderComplete)
	orderCompleteRx := make(chan MessageOrderComplete)

    elevatorStatusTx := make(chan ElevatorStatus) // Channel to transmit elevator status
    elevatorStatusRx := make(chan ElevatorStatus) // Channel to receive elevator status (if needed)

    go bcast.Transmitter(_ListeningPort, newOrderTx, orderCompleteTx, elevatorStatusTx) // You can add more channels as needed
	go bcast.Receiver(_ListeningPort, newOrderRx, orderCompleteRx, elevatorStatusRx) // You can add more channels as needed
    go BroadcastElevatorStatus(myElevator, elevatorStatusTx) // Start broadcasting the elevator status

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    fmt.Println("Elevator initialized")

    for {
        select {

        case elevatorStatus := <-elevatorStatusRx:
            fromElevator := elevatorStatus.E // Get the elevator status from the received message
            fmt.Println("Received elevator status: ", fromElevator.ID)
            UpdateActiveElevators(fromElevator) // Update the elevator status
            DetermineMaster()


        case p := <-peerUpdateCh:
            fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

            for _, peer := range p.Lost {
                for _, elevator := range Elevators {
                    if peer == strconv.Itoa(myElevator.ID) {
                        elevator.isActive = false
                    }
                }
            }
            DetermineMaster() // Re-evaluate the master elevator
            

        case Order := <-newOrderRx:

            newOrder := Order.NewOrder
            // fromElevator := Order.E
            toElevatorID := Order.ToElevatorID

            if toElevatorID != myElevator.ID {
                fmt.Println("New order received: ", newOrder)
                // if newOrder == Cab { add to global order system}
                // else, if myElevator.isMaster -> {
                    // Find best elevator for order
                    //    newOrder := MessageNewOrder{
                    //    Type:     "MessageNewOrder",
                    //    NewOrder: newOrder,
                    //    E: myElevator, // Use the correct field name as defined in your ElevatorStatus struct
                    //    ToElevatorID: bestElevatorID,
                    // }
                    //
                    // newOrderTx <- newOrder  (only if bestElevatorID != myElevator.ID)
            }
                //}

                // if not master, check if ToElevatorID == myElevator.ID
                // if true, add to local order system
        

        
        case orderComplete := <-orderCompleteRx:

            order := orderComplete.Order
            fromElevator := orderComplete.E
            fromElevatorID := orderComplete.FromElevatorID

            if fromElevatorID != myElevator.ID {
                UpdateActiveElevators(fromElevator) // Update the elevator status
                fmt.Println("Order completed: ", order)
                // Update global order system
            }


            

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
