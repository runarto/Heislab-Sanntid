package main


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

type State int

const (
    Stop State = iota// 0
    Moving // 1
    Still // 2
)



type GlobalOrderArray struct {
    HallOrderArray [numFloors][2]int // Represents the hall orders
    CabOrderArray [numOfElevators][numFloors]int // Represents the cab orders
}





