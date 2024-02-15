package main

const (
    numFloors = 4
    numOfElevators = 3
    Up = 0
    Down = 1
)

var (
    HallOrderArray [numFloors][2]int //Represents the hall orders for each floor, 
                                     //and whether the order is up or down. 1 means it is active. 0 means not active
    CabOrderArray [numFloors][numOfElevators]int //Represents the cab orders for each floor. 1 means it is active. 0 means not active
    Elevators []Elevator

	//gOrderArray er bestående etasje orderen kommer fra/skal til, type knapp, og om ordren kom innvendig fra (true false)
)


