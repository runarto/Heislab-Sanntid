package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/bcast"
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/pkg/elev"
	"github.com/runarto/Heislab-Sanntid/pkg/orders"
	"github.com/runarto/Heislab-Sanntid/pkg/utils"
)

func main() {

	// Initialize the elevator
	var port = flag.String("port", "15657", "define the port number")
	flag.Parse()
	elevio.Init("localhost:"+*port, utils.NumFloors)

	var myElevator utils.Elevator = utils.Elevator{
		CurrentState:     utils.Still, // Assuming Still is a defined constant in the State type
		CurrentDirection: elevio.MD_Stop,
		GeneralDirection: utils.Stopped,                            // Example, use a valid value from elevio.MotorDirection
		CurrentFloor:     elevio.GetFloor(),                        // Starts at floor 0
		DoorOpen:         false,                                    // Door starts closed
		Obstructed:       false,                                    // No obstruction initially
		StopButton:       false,                                    // Stop button not pressed initially
		LocalOrderArray:  [utils.NumButtons][utils.NumFloors]int{}, // Initialize with zero values
		IsMaster:         false,                                    // Not master initially
		ID:               0,                                        // Set to the ID of the elevator
		IsActive:         true,                                     // Elevator is active initially
	}

	utils.Elevators = append(utils.Elevators, myElevator) // Add the elevator to the list of active elevators
	orders.InitLocalOrderSystem(&myElevator)              // Initialize the local order system
	elev.InitializeElevator(&myElevator)                  // Initialize the elevator

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(myElevator.ID), peerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, peerUpdateCh)

	newOrderTx := make(chan utils.MessageNewOrder)
	newOrderRx := make(chan utils.MessageNewOrder)

	orderCompleteTx := make(chan utils.MessageOrderComplete)
	orderCompleteRx := make(chan utils.MessageOrderComplete)

	elevatorStatusTx := make(chan utils.ElevatorStatus) // Channel to transmit elevator status
	elevatorStatusRx := make(chan utils.ElevatorStatus) // Channel to receive elevator status (if needed)

	orderArraysTx := make(chan utils.MessageOrderArrays) // Channel to transmit global order array
	orderArraysRx := make(chan utils.MessageOrderArrays) // Channel to receive global order array (if needed)

	ackStructTx := make(chan utils.AckMatrix) // Channel to transmit the Ack struct
	ackStructRx := make(chan utils.AckMatrix) // Channel to receive the Ack struct (if needed)

	go bcast.Transmitter(utils.ListeningPort, newOrderTx, orderCompleteTx, elevatorStatusTx, orderArraysTx, ackStructTx) // You can add more channels as needed
	go bcast.Receiver(utils.ListeningPort, newOrderRx, orderCompleteRx, elevatorStatusRx, orderArraysRx, ackStructRx)    // You can add more channels as needed

	go elev.BroadcastElevatorStatus(&myElevator, elevatorStatusTx)
	go elev.BroadcastAckMatrix(&myElevator, ackStructTx)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	// Start polling functions in separate goroutines
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go func() {

		for {
			elev.CheckIfOrderIsComplete(&myElevator, newOrderTx, orderCompleteTx)
			time.Sleep(3 * time.Second) // sleep for a while before checking again
		}

	}()

	fmt.Println("Elevator initialized")

	for {
		select {

		case ackStruct := <-ackStructRx:

			if ackStruct.FromElevatorID != myElevator.ID {

				fmt.Println("---ACK STRUCT RECEIVED---")

				val := ackStruct.OrderWatcher
				utils.OrderWatcher = val

			}

		case orderArrays := <-orderArraysRx:

			toElevatorID := orderArrays.ToElevatorID
			fromElevator := orderArrays.FromElevator

			if toElevatorID == myElevator.ID {

				fmt.Println("---ORDER ARRAY RECEIVED---")

				fmt.Println("Received order arrays")

				myElevator.LocalOrderArray = orderArrays.LocalOrderArray
				orders.PrintLocalOrderSystem(&myElevator)

				utils.GlobalOrders = orderArrays.GlobalOrders

				HallOrders := utils.GlobalOrders.HallOrderArray
				CabOrders := utils.GlobalOrders.CabOrderArray

				for button := 0; button < utils.NumButtons-1; button++ {
					for floor := 0; floor < utils.NumFloors; floor++ {
						if HallOrders[button][floor] == utils.True {

							Order := utils.Order{
								Floor:  floor,
								Button: elevio.ButtonType(button)}

							orders.UpdateGlobalOrderSystem(Order, &fromElevator, true)

						} else {

							Order := utils.Order{
								Floor:  floor,
								Button: elevio.ButtonType(button)}

							orders.UpdateGlobalOrderSystem(Order, &fromElevator, false)
						}
					}
				}

				for floor := 0; floor < utils.NumFloors; floor++ {

					if CabOrders[myElevator.ID][floor] == utils.True {
						Order := utils.Order{
							Floor:  floor,
							Button: elevio.ButtonType(utils.Cab)}

						orders.UpdateLocalOrderSystem(Order, &myElevator)
					}
				}

				if orders.CheckAmountOfActiveOrders(&myElevator) > 0 {

					utils.BestOrder = orders.ChooseBestOrder(&myElevator) // Choose the best order
					fmt.Println("Best order: ", utils.BestOrder)

					if utils.BestOrder.Floor == myElevator.CurrentFloor {

						elev.HandleElevatorAtFloor(utils.BestOrder.Floor, orderCompleteTx, &myElevator) // Handle the elevator at the floor

					} else {

						elev.DoOrder(utils.BestOrder, orderCompleteTx, &myElevator) // Move the elevator to the best order
					}
				} else {

					myElevator.StopElevator()

				}
			}

		case elevatorStatus := <-elevatorStatusRx:

			elevator := elevatorStatus.FromElevator

			if elevator.ID != myElevator.ID {
				fmt.Println("---ELEVATOR STATUS RECEIVED---")

				fmt.Println("Local order system from elevator ", elevator.ID)
				orders.PrintLocalOrderSystem(&elevator)

				fmt.Println("Received elevator status: ", elevator.ID) // Update the elevator status
				elev.UpdateElevatorsOnNetwork(&elevator)               // Update the active elevators
				elev.DetermineMaster(&myElevator)                      // Determine the master elevator

			}

		case p := <-peerUpdateCh:

			fmt.Println("---PEER UPDATE RECEIVED---")

			elev.HandlePeersUpdate(p, elevatorStatusTx, orderArraysTx, newOrderTx, &myElevator) // Handle the peer update

		case Order := <-newOrderRx:

			newOrder := Order.NewOrder
			fromElevator := Order.FromElevator
			toElevatorID := Order.ToElevatorID

			fmt.Println("---NEW ORDER RECEIVED---")

			fmt.Println("Received order from elevator", fromElevator.ID, "meant for ", toElevatorID)

			fmt.Println("I am elevator", myElevator.ID)

			elev.HandleNewOrder(newOrder, &fromElevator, toElevatorID, orderCompleteTx, newOrderTx, &myElevator) // Handle the new orde

		case orderComplete := <-orderCompleteRx:

			completedOrders := orderComplete.Orders
			fromElevatorID := orderComplete.FromElevatorID
			fromElevator := orderComplete.FromElevator

			if fromElevatorID != myElevator.ID {

				fmt.Println("---ORDER COMPLETE RECEIVED---")
				// Update the elevator status
				fmt.Println("Order completed: ", completedOrders, "by elevator", fromElevatorID)
				elev.UpdateElevatorsOnNetwork(&fromElevator)

				for i, _ := range completedOrders {
					value := orders.CheckIfGlobalOrderIsActive(completedOrders[i], &myElevator)
					fmt.Println("Value: ", value)
					orders.UpdateGlobalOrderSystem(completedOrders[i], &myElevator, false)
					elev.OrderCompleted(completedOrders[i], &fromElevator)
				}
			}

			if orders.CheckAmountOfActiveOrders(&myElevator) > 0 {

				utils.BestOrder = orders.ChooseBestOrder(&myElevator) // Choose the best order
				fmt.Println("Best order: ", utils.BestOrder)

				if utils.BestOrder.Floor == myElevator.CurrentFloor {

					elev.HandleElevatorAtFloor(utils.BestOrder.Floor, orderCompleteTx, &myElevator) // Handle the elevator at the floor

				} else {

					elev.DoOrder(utils.BestOrder, orderCompleteTx, &myElevator) // Move the elevator to the best order
				}
			} else {

				myElevator.StopElevator()

			}

		case btn := <-drv_buttons:

			fmt.Println("---BUTTON PRESSED---")

			floor := btn.Floor
			button := btn.Button

			newOrder := utils.Order{
				Floor:  floor,
				Button: button}
			fmt.Println("New local order: ", newOrder)

			elev.HandleButtonEvent(newOrderTx, orderCompleteTx, newOrder, &myElevator)

		case floor := <-drv_floors:

			fmt.Println("---ARRIVED AT NEW FLOOR---")

			fmt.Println("Arrived at floor: ", floor)

			elev.FloorLights(floor, &myElevator)                            // Update the floor lights
			elev.HandleElevatorAtFloor(floor, orderCompleteTx, &myElevator) // Handle the elevator at the floor

		case obstr := <-drv_obstr:
			myElevator.Obstruction(obstr)

		case stop := <-drv_stop:
			myElevator.StopBtnPressed(stop)
			//StopButton(stop)

		}
	}
}
