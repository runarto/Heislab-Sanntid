package main

import (
	"flag"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/bcast"
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/fsm"
	"github.com/runarto/Heislab-Sanntid/net"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//TODO: Have three elevators running. Disconnect the master, and see what happens.
// If the master is disconnected, the other elevators should elect a new master, and
// reallocate the actice orders of the master.

const bufferSize = 216

func main() {

	// Initialize the elevator
	var port = flag.String("port", "15657", "define the port number")
	var id_val = flag.Int("id", 0, "define the elevator id")

	flag.Parse()
	utils.ID = *id_val
	elevio.Init("localhost:"+*port, utils.NumFloors)

	var e utils.Elevator
	fsm.NullButtons()
	e = fsm.InitializeElevator()
	utils.Elevators = append(utils.Elevators, e)

	//---------------------------------------------------------------------------

	// General message sending and distributing channels channels
	messageSender := make(chan net.Packet, bufferSize)
	messageReceiver := make(chan net.Packet, bufferSize)
	messageHandler := make(chan utils.Message, bufferSize)
	messageDistributor := make(chan utils.Message, bufferSize)

	//Event channels
	ButtonPressCh := make(chan elevio.ButtonEvent, bufferSize)
	PeerUpdateCh := make(chan peers.PeerUpdate)
	OrderHandlerNetworkUpdateCh := make(chan utils.Message, bufferSize)
	OfflineOrderCompleteCh := make(chan utils.Order, bufferSize)

	//Update channels
	AllOrdersCh := make(chan utils.GlobalOrderUpdate, bufferSize)
	OrderWatcherCh := make(chan utils.OrderWatcher, bufferSize)
	MasterUpdateCh := make(chan int, bufferSize)
	ActiveElevatorUpdateCh := make(chan utils.Status, bufferSize)
	LocalElevatorStateUpdateCh := make(chan utils.Elevator)
	ArrayUpdater := make(chan utils.Message, bufferSize)
	SlaveUpdateCh := make(chan utils.Order, bufferSize)
	ActiveOrdersCh := make(chan map[int][3][utils.NumFloors]bool)

	//FSM channels
	DoOrderCh := make(chan utils.Order, bufferSize)
	IsOnlineCh := make(chan bool)

	//Message sending channels
	MasterID_Tx := make(chan int)
	LightsTx := make(chan utils.MessageLights)
	ElevatorStateTx := make(chan utils.MessageElevatorStatus)
	MasterOrderWatcherTx := make(chan utils.MessageOrderWatcher)
	PeerTxEnable := make(chan bool)

	// Message receiving channels
	MasterID_Rx := make(chan int)
	LightsRx := make(chan utils.MessageLights)
	ElevatorStateRx := make(chan utils.MessageElevatorStatus)
	SetLights := make(chan [2][utils.NumFloors]bool)
	MasterOrderWatcherRx := make(chan utils.MessageOrderWatcher)

	//---------------------------------------------------------------------------

	// Broadcasting -----------------------------------------------------------
	go bcast.Transmitter(utils.ListeningPort, ElevatorStateTx, LightsTx, MasterOrderWatcherTx, MasterID_Tx, messageSender)
	go bcast.Receiver(utils.ListeningPort, ElevatorStateRx, LightsRx, MasterOrderWatcherRx, MasterID_Rx, messageReceiver)

	// Peer handling -----------------------------------------------------------
	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(e.ID), PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, PeerUpdateCh)

	time.Sleep(1000 * time.Millisecond)

	// Button polling ----------------------------------------------------------------
	go elevio.PollButtons(ButtonPressCh)

	// Message sending, receiving and distributing------------------------------------------------
	go net.NetworkHandler(messageHandler, messageReceiver, messageSender, messageDistributor, OrderHandlerNetworkUpdateCh, SlaveUpdateCh)
	//go net.MessageDistributor(messageDistributor, OrderHandlerNetworkUpdate, SetLights)
	go net.BroadcastMaster(MasterID_Tx)
	go net.BroadcastElevatorState(LocalElevatorStateUpdateCh, ElevatorStateTx, &e)
	go net.BroadcastLights(LightsTx)
	go net.BroadcastOrderWatcher(MasterOrderWatcherTx, ActiveOrdersCh)
	go net.ReceiveBroadcasts(MasterOrderWatcherRx, ElevatorStateRx, LightsRx, ArrayUpdater, SetLights, MasterID_Rx, MasterUpdateCh)

	// FSM, Order-handling and variable-updaters ------------------------------------------------
	go fsm.FSM(e, DoOrderCh, LocalElevatorStateUpdateCh, PeerTxEnable, IsOnlineCh, SetLights, messageHandler, OfflineOrderCompleteCh)

	go orders.OrderHandler(e, ButtonPressCh, AllOrdersCh, OrderHandlerNetworkUpdateCh, PeerUpdateCh,
		DoOrderCh, LocalElevatorStateUpdateCh, messageHandler, IsOnlineCh, ActiveElevatorUpdateCh, OfflineOrderCompleteCh)

	go updater.Updater(e, AllOrdersCh, OrderWatcherCh, LocalElevatorStateUpdateCh, messageHandler, ActiveElevatorUpdateCh, ArrayUpdater, SlaveUpdateCh, ActiveOrdersCh, ButtonPressCh)

	select {}

}
