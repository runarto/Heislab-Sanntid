package main

import (
	"flag"
	"strconv"

	"github.com/runarto/Heislab-Sanntid/Network/bcast"
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/fsm"
	"github.com/runarto/Heislab-Sanntid/net"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func main() {

	bufferSize := 100

	// Initialize the elevator
	var port = flag.String("port", "15657", "define the port number")
	flag.Parse()
	elevio.Init("localhost:"+*port, utils.NumFloors)

	var thisElevator utils.Elevator = utils.Elevator{
		CurrentState:     utils.Still, // Assuming Still is a defined constant in the State type
		CurrentDirection: elevio.MD_Stop,
		CurrentFloor:     elevio.GetFloor(),                         // Starts at floor 0
		LocalOrderArray:  [utils.NumButtons][utils.NumFloors]bool{}, // Initialize with zero values
		ID:               1,                                         // Set to the ID of the elevator
		IsActive:         true,                                      // Elevator is active initially
	}

	fsm.NullButtons()
	fsm.InitializeElevator(&thisElevator)

	c := utils.Channels{
		PeerTxEnable: make(chan bool, bufferSize),
		PeerUpdateCh: make(chan peers.PeerUpdate, bufferSize),

		NewOrderTx: make(chan utils.MessageNewOrder, bufferSize),
		NewOrderRx: make(chan utils.MessageNewOrder, bufferSize),

		OrderCompleteTx: make(chan utils.MessageOrderComplete, bufferSize),
		OrderCompleteRx: make(chan utils.MessageOrderComplete, bufferSize),

		OrderArraysTx: make(chan utils.MessageGlobalOrderArrays, bufferSize),
		OrderArraysRx: make(chan utils.MessageGlobalOrderArrays, bufferSize),

		ElevatorStatusTx: make(chan utils.MessageElevatorStatus, bufferSize),
		ElevatorStatusRx: make(chan utils.MessageElevatorStatus, bufferSize),

		MasterOrderWatcherTx: make(chan utils.MessageOrderWatcher, bufferSize),
		MasterOrderWatcherRx: make(chan utils.MessageOrderWatcher, bufferSize),

		AckTx: make(chan utils.MessageOrderConfirmed, bufferSize),
		AckRx: make(chan utils.MessageOrderConfirmed, bufferSize),

		LightsTx: make(chan utils.MessageLights, bufferSize),
		LightsRx: make(chan utils.MessageLights, bufferSize),

		GlobalUpdateCh: make(chan utils.GlobalOrderUpdate, bufferSize),
		ButtonCh:       make(chan elevio.ButtonEvent, bufferSize),
		FloorCh:        make(chan int, bufferSize),
		ObstrCh:        make(chan bool, bufferSize),
		StopCh:         make(chan bool, bufferSize),
		BestOrderCh:    make(chan utils.Order, bufferSize),
	}

	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(thisElevator.ID), c.PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, c.PeerUpdateCh)

	go bcast.Transmitter(utils.ListeningPort, c.NewOrderTx, c.OrderCompleteTx,
		c.ElevatorStatusTx, c.OrderArraysTx, c.MasterOrderWatcherTx, c.AckTx) // You can add more channels as needed
	go bcast.Receiver(utils.ListeningPort, c.NewOrderRx, c.OrderCompleteRx,
		c.ElevatorStatusRx, c.OrderArraysRx, c.MasterOrderWatcherRx, c.AckRx) // You can add more channels as needed

	go net.BroadcastElevatorStatus(&thisElevator, &c)
	go net.BroadcastMasterOrderWatcher(&thisElevator, &c)

	// Start polling functions in separate goroutines
	go elevio.PollButtons(c.ButtonCh)
	go elevio.PollFloorSensor(c.FloorCh)
	go elevio.PollObstructionSwitch(c.ObstrCh)
	go elevio.PollStopButton(c.StopCh)

	go fsm.FSM(&c, thisElevator)
	go orders.OrderHandler(&c, thisElevator)
	go updater.Updater(&c, thisElevator)

	select {}

}
