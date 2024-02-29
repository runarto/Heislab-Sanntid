package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/runarto/Heislab-Sanntid/Network/bcast"
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
)

func main() {

	// Initialize the elevator
	var port = flag.String("port", "15657", "define the port number")
	flag.Parse()
	elevio.Init("localhost:"+*port, numFloors)

	var myElevator Elevator = Elevator{
		CurrentState:     Still, // Assuming Still is a defined constant in the State type
		CurrentDirection: elevio.MD_Stop,
		GeneralDirection: Stopped,             // Example, use a valid value from elevio.MotorDirection
		CurrentFloor:     elevio.GetFloor(),   // Starts at floor 0
		doorOpen:         false,               // Door starts closed
		Obstruction:      false,               // No obstruction initially
		stopButton:       false,               // Stop button not pressed initially
		LocalOrderArray:  [3][numFloors]int{}, // Initialize with zero values
		isMaster:         false,               // Not master initially
		ID:               0,                   // Set to the ID of the elevator
		isActive:         true,                // Elevator is active initially
	}

	Elevators = append(Elevators, myElevator) // Add the elevator to the list of active elevators
	myElevator.InitLocalOrderSystem()         // Initialize the local order system
	myElevator.InitElevator()                 // Initialize the elevator

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
	go bcast.Receiver(_ListeningPort, newOrderRx, orderCompleteRx, elevatorStatusRx, orderArraysRx)    // You can add more channels as needed
	go BroadcastElevatorStatus(myElevator, elevatorStatusTx)                                           // Start broadcasting the elevator status

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
				UpdateElevatorsOnNetwork(elevator)                     // Update the active elevators
				myElevator.DetermineMaster()                           // Determine the master elevator
				// fromElevator := elevatorStatus.E

				// if fromElevator.ID != myElevator.ID {
				//     UpdateElevatorsOnNetwork(fromElevator) // Update the active elevators
				//     // Get the elevator status from the received message
				//     fmt.Println("Received elevator status: ", fromElevator.ID) // Update the elevator status
				//     myElevator.DetermineMaster()

				// }
			}

		case p := <-peerUpdateCh:

			myElevator.HandlePeersUpdate(p, elevatorStatusTx)

		case Order := <-newOrderRx:

			newOrder := Order.NewOrder
			fromElevator := Order.E
			toElevatorID := Order.ToElevatorID

			fmt.Println("Received order from elevator", fromElevator.ID)

			myElevator.HandleNewOrder(newOrder, fromElevator, toElevatorID, orderCompleteTx, newOrderTx)

		case orderComplete := <-orderCompleteRx:

			orders := orderComplete.Orders
			fromElevatorID := orderComplete.FromElevatorID

			if fromElevatorID != myElevator.ID {
				// Update the elevator status
				fmt.Println("Order completed: ", orders, "by elevator", fromElevatorID)
				// Update global order system
			}

			if myElevator.CheckAmountOfActiveOrders() > 0 {

				bestOrder = myElevator.ChooseBestOrder() // Choose the best order
				fmt.Println("Best order: ", bestOrder)

				if bestOrder.Floor == myElevator.CurrentFloor {

					myElevator.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor

				} else {

					myElevator.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
				}
			} else {

				myElevator.StopElevator()

			}

		case btn := <-drv_buttons:

			floor := btn.Floor
			button := btn.Button
			newOrder := Order{floor, button}
			fmt.Println("New local order: ", newOrder)

			myElevator.HandleButtonEvent(newOrderTx, orderCompleteTx, newOrder)

		case floor := <-drv_floors:

			fmt.Println("Arrived at floor: ", floor)

			myElevator.floorLights(floor)                            // Update the floor lights
			myElevator.HandleElevatorAtFloor(floor, orderCompleteTx) // Handle the elevator at the floor

		case obstr := <-drv_obstr:
			myElevator.isObstruction(obstr)

		case stop := <-drv_stop:
			myElevator.StopButton(stop)

		}
	}
}
