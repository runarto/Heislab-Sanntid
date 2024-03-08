package utils

import (
	"sync"
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
	DoorOpenTime   = 3
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

var (
	MasterID      int
	MasterIDmutex sync.Mutex
)

type State int

const (
	Stop     State = iota // 0
	Moving                // 1
	Still                 // 2
	DoorOpen              // 3
)

type GlobalOrderArray struct {
	HallOrderArray [2][NumFloors]bool              // Represents the hall orders
	CabOrderArray  [NumOfElevators][NumFloors]bool // Represents the cab orders
}

var GlobalOrders = GlobalOrderArray{
	HallOrderArray: [2][NumFloors]bool{},
	CabOrderArray:  [NumOfElevators][NumFloors]bool{},
}

var MasterOrderWatcher = OrderWatcherArray{
	WatcherMutex:   sync.Mutex{},
	HallOrderArray: [2][NumFloors]HallAck{},
	CabOrderArray:  [NumOfElevators][NumFloors]CabAck{},
}

var SlaveOrderWatcher = OrderWatcherArray{
	WatcherMutex:   sync.Mutex{},
	HallOrderArray: [2][NumFloors]HallAck{},
	CabOrderArray:  [NumOfElevators][NumFloors]CabAck{},
}

var BestOrder = Order{
	Floor:  NotDefined,
	Button: elevio.BT_HallUp}

type GlobalOrderUpdate struct {
	Order          Order
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
	WatcherMutex   sync.Mutex
	HallOrderArray [2][NumFloors]HallAck             // Represents the hall orders
	CabOrderArray  [NumOfElevators][NumFloors]CabAck // Represents the cab orders
}

type OrderWatcherArrayToSend struct {
	HallOrderArray [2][NumFloors]HallAck
	CabOrderArray  [NumOfElevators][NumFloors]CabAck
}

type OrderWatcher struct {
	Order         Order
	ForElevatorID int
	IsComplete    bool
	IsNew         bool
	IsConfirmed   bool
}
