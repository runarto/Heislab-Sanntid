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

	OrderArraysTx chan MessageOrderArrays
	OrderArraysRx chan MessageOrderArrays

	ElevatorStatusTx chan ElevatorStatus
	ElevatorStatusRx chan ElevatorStatus

	MasterOrderWatcherTx chan MessageOrderWatcher
	MasterOrderWatcherRx chan MessageOrderWatcher

	AckTx chan OrderConfirmed
	AckRx chan OrderConfirmed

	GlobalUpdateCh chan GlobalOrderUpdate
	ButtonCh       chan elevio.ButtonEvent
	FloorCh        chan int
	ObstrCh        chan bool
	StopCh         chan bool
	BestOrderCh    chan Order
	PeersOnlineCh  chan NewPeersMessage
	ElevatorsCh    chan []Elevator
	MasterBarkCh   chan Order
	SlaveBarkCh    chan Order
	OrderWatcher   chan OrderWatcher
}
