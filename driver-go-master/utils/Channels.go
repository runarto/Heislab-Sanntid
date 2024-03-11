package utils

import (
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Channels struct {
	PeerUpdateCh chan peers.PeerUpdate
	PeerTxEnable chan bool

	NewOrderTx chan MessageNewOrder
	NewOrderRx chan MessageNewOrder

	OrderCompleteTx chan MessageOrderComplete
	OrderCompleteRx chan MessageOrderComplete

	ElevatorStatusTx chan MessageElevatorStatus
	ElevatorStatusRx chan MessageElevatorStatus

	MasterOrderWatcherTx chan MessageOrderWatcher
	MasterOrderWatcherRx chan MessageOrderWatcher

	AckTx chan MessageOrderConfirmed
	AckRx chan MessageOrderConfirmed

	LightsTx chan MessageLights
	LightsRx chan MessageLights

	GlobalUpdateCh     chan GlobalOrderUpdate
	ButtonCh           chan elevio.ButtonEvent
	FloorCh            chan int
	ObstrCh            chan bool
	StopCh             chan bool
	ElevatorsCh        chan []Elevator
	MasterBarkCh       chan Order
	SlaveBarkCh        chan Order
	OrderWatcher       chan OrderWatcher
	DoOrderCh          chan Order
	LocalStateUpdateCh chan Elevator
	IsOnlineCh         chan bool
	LocalLightsCh      chan [2][NumFloors]bool
	MasterUpdateCh     chan int
}
