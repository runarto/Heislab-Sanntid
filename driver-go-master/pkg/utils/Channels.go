package utils

import (
	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
)

type Channels struct {
	NewOrderTx chan MessageNewOrder
	NewOrderRx chan MessageNewOrder

	OrderCompleteTx chan MessageOrderComplete
	OrderCompleteRx chan MessageOrderComplete

	OrderArraysTx chan MessageOrderArrays
	OrderArraysRx chan MessageOrderArrays

	ElevatorStatusTx chan ElevatorStatus
	ElevatorStatusRx chan ElevatorStatus

	AckStructTx chan AckMatrix
	AckStructRx chan AckMatrix

	GlobalUpdateCh chan GlobalOrderUpdate
	PeerUpdateCh   chan peers.PeerUpdate
	peerTxEnable   chan bool
	ButtonCh       chan elevio.ButtonEvent
	FloorCh        chan int
	ObstrCh        chan bool
	StopCh         chan bool
	BestOrderCh    chan Order
}
