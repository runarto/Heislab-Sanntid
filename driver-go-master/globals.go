package main

const (
    MaxOrders       = 12
    On              = true
    Off             = false
    Open            = 1
    Close           = 0
    NotDefined      = -1
    ElevUp          = 1
    ElevDown        = -1
    FloorUp         = 0
    FloorDown       = 1
    FromCab         = 2
    numFloors       = 4
    numButtons      = 3
    True            = 1
    False           = 0
    Cab             = 1
    Up              = 0
    Down            = 1
    CabUp           = 0
    CabDown         = 1
)

type State int

const (
    Stop State = iota
    Moving
    Still
)

var (
    CurrentState      State
    CurrentDirection  int
    CurrentDirectionAlt int
    LastDefinedFloor  int
    currentFloor      int
    IsDoorOpen        int
    OrderComplete     bool
    OrderArray        [MaxOrders][3]int
	//gOrderArray er bestående etasje orderen kommer fra/skal til, type knapp, og om ordren kom innvendig fra (true false)
)


