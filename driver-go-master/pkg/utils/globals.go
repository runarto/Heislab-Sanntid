package utils

import (
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
)

const (
	NumFloors      = 4
	NumOfElevators = 3
	NotDefined     = -1
	NumButtons     = 3
	ListeningPort  = 29876
	Timeout        = 3 * time.Second
	MaxRetries     = 3
	SlaveTimeout   = 3 * time.Second
	MasterTimeout  = 15 * time.Second
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

var MasterOrderWatcher = OrderWatcherArray{
	HallOrderArray: [2][NumFloors]HallAck{},
	CabOrderArray:  [NumOfElevators][NumFloors]CabAck{},
}

var SlaveOrderWatcher = OrderWatcherArray{
	HallOrderArray: [2][NumFloors]HallAck{},
	CabOrderArray:  [NumOfElevators][NumFloors]CabAck{},
}

var BestOrder = Order{
	Floor:  NotDefined,
	Button: elevio.BT_HallUp}

type GlobalOrderUpdate struct {
	Orders         []Order
	FromElevatorID int
	IsComplete     bool
	IsNew          bool
}

type NewPeersMessage struct {
	LostPeers []int
	NewPeers  []int
}

type HallAck struct {
	Active    bool
	Completed bool
	Confirmed bool
	Time      time.Time
}

type CabAck struct {
	Active    bool
	Completed bool
	Confirmed bool
	Time      time.Time
}

type OrderWatcherArray struct {
	HallOrderArray [2][NumFloors]HallAck             // Represents the hall orders
	CabOrderArray  [NumOfElevators][NumFloors]CabAck // Represents the cab orders
}

type OrderWatcher struct {
	Orders        []Order
	ForElevatorID int
	New           bool
	Complete      bool
}
