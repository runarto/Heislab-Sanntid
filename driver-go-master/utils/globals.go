package utils

import (
	"sync"
	"time"
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
	MasterTimeout  = 10 * time.Second
	DoorOpenTime   = 3
)

const (
	HallUp   = 0
	HallDown = 1
	Cab      = 2

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

var NextMasterID int

var ID int

var (
	AckMatrix = make(map[int][NumButtons][NumFloors]Ack)
	AckMutex  sync.Mutex
)

type State int

const (
	Stop     State = iota // 0
	Moving                // 1
	Still                 // 2
	DoorOpen              // 3
)

type GlobalOrderUpdate struct {
	Order          Order
	ForElevatorID  int
	FromElevatorID int
	IsComplete     bool
	IsNew          bool
}

type NewPeersMessage struct {
	LostPeers []int
	NewPeers  []int
}

type Ack struct {
	Active    bool
	Completed bool
	Confirmed bool
	Time      time.Time
}

type OrderWatcherArray struct {
	WatcherMutex   sync.Mutex
	HallOrderArray [2][NumFloors]Ack // Represents the hall orders
}

type OrderWatcher struct {
	Order         Order
	ForElevatorID int
	IsComplete    bool
	IsNew         bool
	IsConfirmed   bool
}

type Status struct {
	ID       int
	IsOnline bool
}
