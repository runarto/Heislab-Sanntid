package main

import (
	"flag"
	"fmt"
	"strconv"

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

	var thisElevator utils.Elevator = utils.Elevator{
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

	utils.Elevators = append(utils.Elevators, thisElevator) // Add the elevator to the list of active elevators
	orders.InitLocalOrderSystem(&thisElevator)              // Initialize the local order system
	elev.InitializeElevator(&thisElevator)                  // Initialize the elevator

	channels := &utils.Channels{

		PeerUpdateCh: make(chan peers.PeerUpdate),
		PeerTxEnable: make(chan bool),

		NewOrderTx: make(chan utils.MessageNewOrder),
		NewOrderRx: make(chan utils.MessageNewOrder),

		OrderCompleteTx: make(chan utils.MessageOrderComplete),
		OrderCompleteRx: make(chan utils.MessageOrderComplete),

		OrderArraysTx: make(chan utils.MessageOrderArrays),
		OrderArraysRx: make(chan utils.MessageOrderArrays),

		ElevatorStatusTx: make(chan utils.ElevatorStatus),
		ElevatorStatusRx: make(chan utils.ElevatorStatus),

		MasterOrderWatcherTx: make(chan utils.MessageOrderWatcher),
		MasterOrderWatcherRx: make(chan utils.MessageOrderWatcher),

		AckTx: make(chan utils.OrderConfirmed),
		AckRx: make(chan utils.OrderConfirmed),

		GlobalUpdateCh: make(chan utils.GlobalOrderUpdate),
		BestOrderCh:    make(chan utils.Order),
		ButtonCh:       make(chan elevio.ButtonEvent),
		FloorCh:        make(chan int),
		ObstrCh:        make(chan bool),
		StopCh:         make(chan bool),
	}

	fmt.Println("lessgoo")

	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(thisElevator.ID), channels.PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, channels.PeerUpdateCh)

	go bcast.Transmitter(utils.ListeningPort, channels.NewOrderTx, channels.OrderCompleteTx,
		channels.ElevatorStatusTx, channels.OrderArraysTx, channels.MasterOrderWatcherTx, channels.AckTx) // You can add more channels as needed
	go bcast.Receiver(utils.ListeningPort, channels.NewOrderRx, channels.OrderCompleteRx,
		channels.ElevatorStatusRx, channels.OrderArraysRx, channels.MasterOrderWatcherRx, channels.AckRx) // You can add more channels as needed

	go elev.BroadcastElevatorStatus(&thisElevator, channels)
	go elev.BroadcastMasterOrderWatcher(&thisElevator, channels.MasterOrderWatcherTx)

	go elev.Bark(&thisElevator, channels)
	go elev.Watchdog(channels, &thisElevator)

	// Start polling functions in separate goroutines
	go elevio.PollButtons(channels.ButtonCh)
	go elevio.PollFloorSensor(channels.FloorCh)
	go elevio.PollObstructionSwitch(channels.ObstrCh)
	go elevio.PollStopButton(channels.StopCh)

	go elev.GlobalUpdates(channels, &thisElevator)

	go elev.NetworkUpdate(channels, &thisElevator)

	go elev.FSM(channels, &thisElevator)

	select {}

}
