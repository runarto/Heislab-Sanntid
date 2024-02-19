package main

import (
    "github.com/runarto/Heislab-Sanntid/elevio"
)


const (
    numFloors = 4
    numOfElevators = 3
    NotDefined = -1
    numButtons = 3
)

const (
    HallUp = 0
    HallDown = 1
    Cab = 2

    True = 1
    False = 0

    On = 1
    Off = 0

    Up = 1
    Down = -1
)

const (
    Open = true
    Close = false
)

var _ListeningPort = "20000"
var _broadcastAddr = "255.255.255.255:20000"
// Can we assume that we know the IP of the elevators initially?

type State int

const (
    Stop State = iota// 0
    Moving // 1
    Still // 2
)

type Order struct {
	Floor  int
	Button elevio.ButtonType
	// An order contains the floor (from/to), and the type of button.
}



type GlobalOrderArray struct {
    HallOrderArray [numFloors][2]int // Represents the hall orders
    CabOrderArray [numOfElevators][numFloors]int // Represents the cab orders
}

type MessageGlobalOrder struct { // Send periodically to update the global order system
    // 0x01
    globalOrders GlobalOrderArray
}

type MessageNewOrder struct { // Send when a new order is received
    // 0x02 
    newOrder Order
    e Elevator
}

type MessageOrderComplete struct { // Send when an order is completed
    // 0x03
    order Order
    e Elevator
}

// Thought: This should work, because the last updated "e" instance from 
// elevator is from when an order was received.

var ActiveElevators []Elevator = make([]Elevator, numOfElevators)

var bestOrder Order = Order{NotDefined, elevio.BT_HallUp}





