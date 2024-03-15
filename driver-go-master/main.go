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
	messageSender := make(chan interface{}, bufferSize*100)
	messageDistributor := make(chan interface{}, bufferSize*1000)

	//Event channels
	ButtonPressCh := make(chan elevio.ButtonEvent, bufferSize)
	OrderComplete := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrder := make(chan utils.MessageNewOrder, bufferSize)
	PeerUpdateCh := make(chan peers.PeerUpdate, bufferSize)
	LocalOrdersCh := make(chan [utils.NumButtons][utils.NumFloors]bool, bufferSize)

	//Update channels
	GlobalUpdateCh := make(chan utils.GlobalOrderUpdate, bufferSize)
	OrderWatcher := make(chan utils.OrderWatcher, bufferSize)
	MasterUpdateCh := make(chan int, bufferSize)
	ActiveElevatorUpdateCh := make(chan utils.Status, bufferSize)
	ElevStatus := make(chan utils.MessageElevatorStatus, bufferSize)
	OrdersUpdate := make(chan map[int][utils.NumButtons][utils.NumFloors]bool)
	LocalStateUpdateCh := make(chan utils.Elevator, bufferSize)

	//FSM channels
	DoOrderCh := make(chan utils.Order, bufferSize)
	IsOnlineCh := make(chan bool, bufferSize)

	//Message sending channels
	LightsTx := make(chan utils.MessageLights, bufferSize)
	ElevStatusTx := make(chan utils.MessageElevatorStatus, bufferSize)
	OrderCompleteTx := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrderTx := make(chan utils.MessageNewOrder, bufferSize)
	MasterOrderWatcherTx := make(chan utils.MessageOrderWatcher, bufferSize)
	PeerTxEnable := make(chan bool, bufferSize)
	AckTx := make(chan utils.MessageConfirmed, bufferSize)

	// Message receiving channels
	LightsRx := make(chan utils.MessageLights, bufferSize)
	SendLights := make(chan [2][utils.NumFloors]bool)
	SetLights := make(chan [2][utils.NumFloors]bool)
	OrderCompleteRx := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrderRx := make(chan utils.MessageNewOrder, bufferSize)
	ElevStatusRx := make(chan utils.MessageElevatorStatus, bufferSize)
	MasterOrderWatcherRx := make(chan utils.MessageOrderWatcher, bufferSize)
	AckRx := make(chan utils.MessageConfirmed, bufferSize*10)

	//---------------------------------------------------------------------------

	// Broadcasting -----------------------------------------------------------
	go bcast.Transmitter(utils.ListeningPort, NewOrderTx, OrderCompleteTx, ElevStatusTx, LightsTx, MasterOrderWatcherTx, AckTx)
	go bcast.Receiver(utils.ListeningPort, NewOrderRx, OrderCompleteRx, ElevStatusRx, LightsRx, MasterOrderWatcherRx, AckRx)

	// Broadcasting -----------------------------------------------------------
	go bcast.Transmitter(utils.ListeningPort, NewOrderTx, OrderCompleteTx, ElevStatusTx, LightsTx, MasterOrderWatcherTx, AckTx)
	go bcast.Receiver(utils.ListeningPort, NewOrderRx, OrderCompleteRx, ElevStatusRx, LightsRx, MasterOrderWatcherRx, AckRx)

	// Button polling ----------------------------------------------------------------
	go elevio.PollButtons(ButtonPressCh)

	// Message sending, receiving and distributing------------------------------------------------
	go net.MessagePasser(messageSender, OrderCompleteTx, NewOrderTx, ElevStatusTx, MasterOrderWatcherTx, LightsTx, OrderWatcher, AckRx, DoOrderCh)
	go net.MessageReceiver(NewOrderRx, OrderCompleteRx, ElevStatusRx, AckTx, messageDistributor, LightsRx)
	go net.MessageDistributor(messageDistributor, OrderComplete, ElevStatus, NewOrder, SendLights)
	go net.LightsReceiver(LightsRx, SetLights)

	// FSM, Order-handling and variable-updaters ------------------------------------------------
	go fsm.FSM(e, DoOrderCh, LocalStateUpdateCh, PeerTxEnable, IsOnlineCh, SetLights, messageSender)

	go orders.OrderHandler(e, ButtonPressCh, GlobalUpdateCh, NewOrderRx, OrderComplete, PeerUpdateCh,
		DoOrderCh, LocalStateUpdateCh, MasterUpdateCh, messageSender, IsOnlineCh, ActiveElevatorUpdateCh, OrderWatcher, LocalOrdersCh)

	go updater.LocalUpdater(e, GlobalUpdateCh, OrderWatcher, LocalStateUpdateCh, messageSender, ButtonPressCh, IsOnlineCh, ActiveElevatorUpdateCh,
		DoOrderCh, MasterUpdateCh, OrdersUpdate, LocalOrdersCh, SendLights)

	go updater.GlobalUpdater(ElevStatus, MasterOrderWatcherRx)

	// Broadcasting lights and OrderWatcher -----------------------------------------------------------
	go net.BroadcastLights(SendLights, LightsTx)
	go updater.BroadcastMasterOrderWatcher(messageSender)

	// Peer handling -----------------------------------------------------------
	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(e.ID), PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, PeerUpdateCh)

	select {}

}
