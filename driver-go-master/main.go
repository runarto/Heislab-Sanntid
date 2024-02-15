package main

import (
    "fmt"
    "Exercise3/elevio"
    "sync"
)



func main() {

    numFloors := 4

    // Initialize the elevator
    elevio.Init("localhost:15657", numFloors)
  
    var myElevator Elevator = Elevator {
        CurrentDirection: elevio.MD_Stop, // Assuming MD_Stop is a constant from the elevio package representing a stopped elevator
        doorOpen:         false,
        Obstruction:      false,
        stopButton:       false,
        ActiveOrders:     []activeOrder{}, // Initialize with an empty slice of activeOrder structs
        NetworkAdress:    "192.168.1.100", // Example IP address
        IsMaster:         true,
    }

    // Create channels for handling events
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors := make(chan int)
    drv_obstr := make(chan bool)
    drv_stop := make(chan bool)

    // Start polling functions in separate goroutines
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)


    if (!e.IsMaster) {
        drv_orders := make(chan activeOrder)
        go ReadOrder("15657", myElevator)

    }

    // Main event loop
    for {
        select {
        case order := <-drv_orders:
            // Add the order to the local order array


        case btn := <-drv_buttons:
            // Check if it's from the cab or the hall. 
            // If hall -> FindBestElevator(), else -> AddOrder(



        case floor := <-drv_floors:
            // Iterate over the local order array, and check if there are any orders for the current floor.


        case obstr := <-drv_obstr:


        case stop := <-drv_stop:
          
        }
        
    }
}
