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

	messageSender := make(chan interface{}, bufferSize)
	messageDistributor := make(chan interface{}, bufferSize)
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
	OrderComplete := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrder := make(chan utils.MessageNewOrder, bufferSize)
	ElevStatus := make(chan utils.MessageElevatorStatus, bufferSize)
	OrderWatcherMsg := make(chan utils.MessageOrderWatcher, bufferSize)

	DoOrderCh := make(chan utils.Order, bufferSize)
	LocalStateUpdateCh := make(chan utils.Elevator, bufferSize)
	PeerUpdateCh := make(chan peers.PeerUpdate, bufferSize)
	PeerTxEnable := make(chan bool, bufferSize)
	LightsRx := make(chan utils.MessageLights, bufferSize)
	Lights := make(chan utils.MessageLights, bufferSize)
	LightsTx := make(chan utils.MessageLights, bufferSize)
	LightsConfirmedTx := make(chan utils.MessageLightsConfirmed, bufferSize)
	OrderConfirmed := make(chan utils.MessageOrderConfirmed, bufferSize)
	OrderCompleteTx := make(chan utils.MessageOrderComplete, bufferSize)
	OrderCompleteRx := make(chan utils.MessageOrderComplete, bufferSize)
	NewOrderTx := make(chan utils.MessageNewOrder, bufferSize)
	NewOrderRx := make(chan utils.MessageNewOrder, bufferSize)
	ElevStatusTx := make(chan utils.MessageElevatorStatus, bufferSize)
	ElevStatusRx := make(chan utils.MessageElevatorStatus, bufferSize)
	MasterOrderWatcherTx := make(chan utils.MessageOrderWatcher, bufferSize)
	MasterOrderWatcherRx := make(chan utils.MessageOrderWatcher, bufferSize)
	AckRx := make(chan utils.MessageConfirmed, bufferSize)

	go peers.Transmitter(utils.ListeningPort+1, strconv.Itoa(e.ID), PeerTxEnable)
	go peers.Receiver(utils.ListeningPort+1, PeerUpdateCh)

	go bcast.Transmitter(utils.ListeningPort, NewOrderTx, OrderCompleteTx, ElevStatusTx, MasterOrderWatcherTx, LightsTx,
		LightsConfirmedTx, OrderConfirmed)
	go bcast.Receiver(utils.ListeningPort, NewOrderRx, OrderCompleteRx, ElevStatusRx, MasterOrderWatcherRx, LightsRx)

	go updater.BroadcastMasterOrderWatcher(e, messageSender)

	// Start polling functions in separate goroutines
	go elevio.PollButtons(ButtonPressCh)
	go elevio.PollFloorSensor(FloorSensorCh)
	go elevio.PollObstructionSwitch(ObstrCh)
	go elevio.PollStopButton(StopCh)

	go net.MessagePasser(messageSender, OrderCompleteTx, NewOrderTx, ElevStatusTx, MasterOrderWatcherTx, LightsTx, AckRx, OrderWatcher)
	go net.MessageReceiver(OrderCompleteRx, ElevStatusRx, NewOrderRx, MasterOrderWatcherRx, LightsRx, messageSender, messageDistributor)
	go net.MessageDistributor(messageDistributor, OrderComplete, ElevStatus, NewOrder, OrderWatcherMsg, Lights)

	go fsm.FSM(e, DoOrderCh, FloorSensorCh, ObstrCh, LocalStateUpdateCh, PeerTxEnable, IsOnlineCh, LocalLightsCh, Lights, messageSender)

	go orders.OrderHandler(e, ButtonPressCh, GlobalUpdateCh, NewOrder, OrderComplete, PeerUpdateCh,
		DoOrderCh, LocalStateUpdateCh, MasterUpdateCh, messageSender, IsOnlineCh, ActiveElevatorUpdateCh, OrderWatcher)

	go updater.LocalUpdater(e, GlobalUpdateCh, OrderWatcher, LocalStateUpdateCh, messageSender,
		LocalLightsCh, ButtonPressCh, IsOnlineCh, ActiveElevatorUpdateCh, DoOrderCh, MasterUpdateCh)

	go updater.GlobalUpdater(ElevStatus, OrderWatcherMsg)

	select {}

}
