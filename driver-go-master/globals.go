package main

const (
    numFloors = 4
    numOfElevators = 3
    Up = 0
    Down = 1
)

var (

    type GlobalOrderArray struct {
        HallOrderArray [numFloors][2]int // Represents the hall orders
        CabOrderArray [numOfElevators][numFloors]int // Represents the cab orders
    }



    Elevators []Elevator // Vet ikke om man trenger det, men fint å ha akkurat nå. 


)


