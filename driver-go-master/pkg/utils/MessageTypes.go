package utils

import (
	"time"
)

type Ack struct {
	Active    bool
	Completed bool
	Time      time.Time
}

type GlobalOrderStruct struct {
	HallOrderArray [2][NumFloors]Ack              // Represents the hall orders
	CabOrderArray  [NumOfElevators][NumFloors]Ack // Represents the cab orders
}

type MessageOrderArrays struct { // Send periodically to update the global order system
	Type            string                     `json:"type"` // Explicitly indicate the message type
	GlobalOrders    GlobalOrderArray           `json:"globalOrders"`
	LocalOrderArray [NumButtons][NumFloors]int `json:"localOrderArray"` // The local order array of the elevator
	ToElevatorID    int                        `json:"toElevatorID"`    // The elevator to send the order to
	FromElevatorID  int                        `json:"fromElevatorID"`  // The elevator that sent the order
}

type MessageOrderComplete struct { // Send when an order is completed
	Type           string  `json:"type"` // Explicitly indicate the message type
	Orders         []Order `json:"orders"`
	ToElevatorID   int     `json:"toElevatorID"`
	FromElevatorID int     `json:"fromElevatorID"` // The elevator that completed the order
}

type MessageNewOrder struct { // Send when a new order is received
	Type           string `json:"type"` // Explicitly indicate the message type
	NewOrder       Order  `json:"newOrder"`
	ToElevatorID   int    `json:"toElevatorID"`
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
}

type ElevatorStatus struct {
	Type         string   `json:"type"`     // A type identifier for decoding on the receiving end
	FromElevator Elevator `json:"elevator"` // The Elevator instance
}

type AckMatrix struct {
	Type           string            `json:"type"`           // A type identifier for decoding on the receiving end
	OrderWatcher   GlobalOrderStruct `json:"ackStruct"`      // The Elevator instance
	FromElevatorID int               `json:"fromElevatorID"` // The elevator to send the order to
}
