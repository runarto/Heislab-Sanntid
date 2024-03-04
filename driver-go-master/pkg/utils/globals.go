package utils

import (
	"github.com/runarto/Heislab-Sanntid/elevio"
)

const (
	NumFloors      = 4
	NumOfElevators = 3
	NotDefined     = -1
	NumButtons     = 3
	ListeningPort  = 29876
)

const (
	HallUp   = 0
	HallDown = 1
	Cab      = 2

	True  = 1
	False = 0

	On  = 1
	Off = 0

	Up      = 1
	Stopped = 0
	Down    = -1
)

const (
	Open  = true
	Close = false
)

var MasterElevatorID = NotDefined

type State int

const (
	Stop   State = iota // 0
	Moving              // 1
	Still               // 2
)

type GlobalOrderArray struct {
	HallOrderArray [2][NumFloors]int              // Represents the hall orders
	CabOrderArray  [NumOfElevators][NumFloors]int // Represents the cab orders
}

var GlobalOrders = GlobalOrderArray{
	HallOrderArray: [2][NumFloors]int{},
	CabOrderArray:  [NumOfElevators][NumFloors]int{},
}

var OrderWatcher = GlobalOrderStruct{
	HallOrderArray: [2][NumFloors]Ack{},
	CabOrderArray:  [NumOfElevators][NumFloors]Ack{},
}

var BestOrder = Order{
	Floor:  NotDefined,
	Button: elevio.BT_HallUp}