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

const bufferSize = 100

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

	ch := make(chan interface{}, bufferSize)
	ButtonPressCh := make(chan elevio.ButtonEvent, bufferSize)
	FloorSensorCh := make(chan int, bufferSize)
	ObstrCh := make(chan bool, bufferSize)
	StopCh := make(chan bool, bufferSize)
	GlobalUpdateCh := make(chan utils.GlobalOrderUpdate, bufferSize)
	OrderWatcher := make(chan utils.OrderWatcher, bufferSize)
	MasterUpdateCh := make(chan int, bufferSize)
	IsOnlineCh := make(chan bool, bufferSize)
	LocalLightsCh := make(chan [2][utils.NumFloors]bool, bufferSize)
	ActiveElevatorUpdateCh := make(chan utils.Status, bufferSize)

	DoOrderCh := make(chan utils.Order, bufferSize)
	LocalStateUpdateCh := make(chan utils.Elevator, bufferSize)
	PeerUpdateCh := make(chan peers.PeerUpdate, bufferSize)
	PeerTxEnable := make(chan bool, bufferSize)
	LightsRx := make(chan utils.MessageLights, bufferSize)
	LightsTx := make(chan utils.MessageLights, bufferSize)
	LightsConfirmedTx := make(chan utils.MessageLightsConfirmed, bufferSize)
	OrderConfirmed := make(chan utils.MessageOrderConfirmed, bufferSize)
	GlobalOrderArrayTx := make(chan utils.MessageGlobalOrderArrays, bufferSize)
	OrderCompleteTx := make(chan utils.MessageOrderComplete, bufferSize)
	OrderCompleteRx := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrderTx := make(chan utils.MessageNewOrder, bufferSize)
	NewOrderRx := make(chan utils.MessageNewOrder, bufferSize)
	ElevStatusTx := make(chan utils.MessageElevatorStatus, bufferSize)
	ElevStatusRx := make(chan utils.MessageElevatorStatus, bufferSize)
	MasterOrderWatcherTx := make(chan utils.MessageOrderWatcher, bufferSize)
	MasterOrderWatcherRx := make(chan utils.MessageOrderWatcher, bufferSize)

	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(e.ID), PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, PeerUpdateCh)

	go bcast.Transmitter(utils.ListeningPort, NewOrderTx, OrderCompleteTx, ElevStatusTx, MasterOrderWatcherTx, LightsTx)
	go bcast.Receiver(utils.ListeningPort, NewOrderRx, OrderCompleteRx, ElevStatusRx, MasterOrderWatcherRx, LightsRx)

	go net.BroadcastElevatorStatus(e, ch)
	go net.BroadcastMasterOrderWatcher(e, ch)

	// Start polling functions in separate goroutines
	go elevio.PollButtons(ButtonPressCh)
	go elevio.PollFloorSensor(FloorSensorCh)
	go elevio.PollObstructionSwitch(ObstrCh)
	go elevio.PollStopButton(StopCh)

	go net.MessagePasser(ch, GlobalOrderArrayTx, OrderCompleteTx, NewOrderTx, ElevStatusTx, MasterOrderWatcherTx, OrderConfirmed, LightsTx, LightsConfirmedTx)

	go fsm.FSM(e, DoOrderCh, FloorSensorCh, ObstrCh, LocalStateUpdateCh, PeerTxEnable, IsOnlineCh, LocalLightsCh, LightsRx, ch)

	go orders.OrderHandler(e, ButtonPressCh, GlobalUpdateCh, NewOrderRx, OrderCompleteRx, PeerUpdateCh,
		DoOrderCh, LocalStateUpdateCh, ElevStatusRx, MasterUpdateCh, ch, IsOnlineCh, ActiveElevatorUpdateCh)

	go updater.Updater(e, GlobalUpdateCh, OrderWatcher, LocalStateUpdateCh, MasterOrderWatcherRx, ch, LocalLightsCh,
		ButtonPressCh, IsOnlineCh, ElevStatusRx, ActiveElevatorUpdateCh)

	select {}

}
