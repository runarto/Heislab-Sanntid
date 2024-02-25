package main

import (
    "github.com/runarto/Heislab-Sanntid/elevio"
    "github.com/runarto/Heislab-Sanntid/Network/bcast"
    "github.com/runarto/Heislab-Sanntid/Network/peers"
    "fmt"
    "strconv"
    "flag"
)



func main() {

    // Initialize the elevator
    var port = flag.String("port", "15658", "define the port number")
    flag.Parse()
    elevio.Init("localhost:"+*port, numFloors)
  
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
        ID:                0,                        // Set to the ID of the elevator
        isActive:         true,                     // Elevator is active initially
    }

    Elevators = append(Elevators, myElevator) // Add the elevator to the list of active elevators
    myElevator.InitLocalOrderSystem() // Initialize the local order system
    myElevator.InitElevator() // Initialize the elevator

    peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

    go peers.Transmitter(_ListeningPort+1, strconv.Itoa(myElevator.ID), peerTxEnable)
	go peers.Receiver(_ListeningPort+1, peerUpdateCh)

    newOrderTx := make(chan MessageNewOrder)
	newOrderRx := make(chan MessageNewOrder)

	orderCompleteTx := make(chan MessageOrderComplete)
	orderCompleteRx := make(chan MessageOrderComplete)

    elevatorStatusTx := make(chan ElevatorStatus) // Channel to transmit elevator status
    elevatorStatusRx := make(chan ElevatorStatus) // Channel to receive elevator status (if needed)

    orderArraysTx := make(chan MessageOrderArrays) // Channel to transmit global order array
    orderArraysRx := make(chan MessageOrderArrays) // Channel to receive global order array (if needed)

    go bcast.Transmitter(_ListeningPort, newOrderTx, orderCompleteTx, elevatorStatusTx, orderArraysTx) // You can add more channels as needed
	go bcast.Receiver(_ListeningPort, newOrderRx, orderCompleteRx, elevatorStatusRx, orderArraysRx) // You can add more channels as needed
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

        case orderArrays := <-orderArraysRx:

            toElevatorID := orderArrays.ToElevatorID

            if toElevatorID == myElevator.ID {

                fmt.Println("Received order arrays")
                // Overwrite existing global order array
                //CompareAndOverwriteLocalOrrderArray() // Compare and overwrite the local order array
            }

        case elevatorStatus := <-elevatorStatusRx:

            elevator := elevatorStatus.E

            if elevator.ID != myElevator.ID {
        
                fmt.Println("Received elevator status: ", elevator.ID) // Update the elevator status
                UpdateElevatorsOnNetwork(elevator) // Update the active elevators
                myElevator.DetermineMaster() // Determine the master elevator
                // fromElevator := elevatorStatus.E

                // if fromElevator.ID != myElevator.ID { 
                //     UpdateElevatorsOnNetwork(fromElevator) // Update the active elevators
                //     // Get the elevator status from the received message
                //     fmt.Println("Received elevator status: ", fromElevator.ID) // Update the elevator status
                //     myElevator.DetermineMaster()

                // }
            }


        case p := <-peerUpdateCh:

            fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

            outer:
            for i, _ := range Elevators {
                if Elevators[i].isActive {
                    fmt.Println("Elevator is active: ", Elevators[i].ID)
                    continue
                }

                for _, peer := range p.Peers {
                    peerID, _ := strconv.Atoi(peer) 
                    fmt.Println("Peer: ", peerID)
                    if Elevators[i].ID == peerID {
                        UpdateElevatorsOnNetwork(Elevators[i])
                        Elevators[i].isActive = true
                        continue outer
                    }
                }
            }

            for i, _ := range Elevators {
                for _, peer := range p.Lost {
                    peerID, _ := strconv.Atoi(peer) 
                    if Elevators[i].ID == peerID {
                        Elevators[i].isActive = false
                        UpdateElevatorsOnNetwork(Elevators[i])
                    }
                }
            }

            myElevator.DetermineMaster() // Determine the master elevator


           
              
            

        case Order := <-newOrderRx:

            newOrder := Order.NewOrder
            toElevatorID := Order.ToElevatorID

            if myElevator.isMaster {
                // Update global order system locally
                // Find the best elevator for the order
                // Send the order to the best elevator ( if hall order )
                //  newOrder := MessageNewOrder{
                    //    Type:     "MessageNewOrder",
                    //    NewOrder: newOrder,
                    //    E: myElevator, // Use the correct field name as defined in your ElevatorStatus struct
                    //    ToElevatorID: bestElevatorID,}
            }


            if toElevatorID == myElevator.ID {

                fmt.Println("New order received: ", newOrder)
                
                myElevator.UpdateOrderSystem(newOrder) // Update the local order array
                myElevator.PrintLocalOrderSystem()
                bestOrder = myElevator.ChooseBestOrder() // Choose the best order
                fmt.Println("Best order: ", bestOrder)
    
                if bestOrder.Floor == myElevator.CurrentFloor {
                    myElevator.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
                } else {
                    myElevator.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
                }
            }
        

        
        case orderComplete := <-orderCompleteRx:

            orders := orderComplete.Orders
            fromElevatorID := orderComplete.FromElevatorID

            if fromElevatorID != myElevator.ID {
                 // Update the elevator status
                fmt.Println("Order completed: ", orders)
                // Update global order system
            }

            bestOrder = myElevator.ChooseBestOrder() // Choose the best order
            fmt.Println("Best order: ", bestOrder)

            if bestOrder.Floor == myElevator.CurrentFloor {
                myElevator.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
            } else {
                myElevator.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
            }


            

        case btn := <-drv_buttons:

            floor := btn.Floor
            button := btn.Button
            newOrder := Order{floor, button}
            fmt.Println("New order: ", newOrder)

            newOrderTx <- MessageNewOrder{Type: "MessageNewOrder", NewOrder: newOrder, E: myElevator, ToElevatorID: myElevator.ID} // Send the new order to the network

            if button == Cab {
                if myElevator.CheckIfOrderIsActive(newOrder) { // Check if the order is active
                    if bestOrder.Floor == myElevator.CurrentFloor {
                        myElevator.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
                    } else {
                        myElevator.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
                    }
                    
                } else {

                    // All this can be removed and implemented under newOrderRx



                    // if myElevator.isMaster -> update global order system locally
                    // else, send order to master

                    //newOrderToSend := MessageNewOrder{newOrder, myElevator} // Create a new order message
                    //SendOrder(masterAddress, newOrderToSend) // Send the order to master

                    // SendOrder(address, newOrder) // Send the order to master

                    bestOrder = myElevator.ChooseBestOrder() // Choose the best order
                    fmt.Println("Best order: ", bestOrder)
        
                    if bestOrder.Floor == myElevator.CurrentFloor {
                        myElevator.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
                    } else {
                        myElevator.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
                    }

                }
            }
            

        case floor := <-drv_floors:

            fmt.Println("Arrived at floor: ", floor)

            myElevator.floorLights(floor) // Update the floor lights
            myElevator.HandleElevatorAtFloor(floor, orderCompleteTx) // Handle the elevator at the floor




        case obstr := <-drv_obstr:
            myElevator.isObstruction(obstr)


        case stop := <-drv_stop:
            myElevator.StopButton(stop)
          
        }
    }
}
