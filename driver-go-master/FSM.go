package main


import (
    "fmt"
    "github.com/runarto/Heislab-Sanntid/elevio"
    "time"
)


func NullButtons() { // Turns off all buttons
    elevio.SetStopLamp(false)
    for f := 0; f < numFloors; f++ {
        for b := 0; b < numButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
        }
    }
}


func (e *Elevator) InitElevator() {
    NullButtons()
    e.SetDoorState(Close) // Close the door
    
    for floor := elevio.GetFloor(); floor != 0; floor = elevio.GetFloor() {
        if floor > 0 || floor == -1 {
            e.GoDown()
        }
        time.Sleep(100 * time.Millisecond)
    }
    e.StopElevator()
    e.CurrentFloor = elevio.GetFloor()
    fmt.Println("Elevator is ready for use")


    


}



func (e *Elevator) floorLights(floor int) {
    if (floor >= 0 && floor <= 3) {
        elevio.SetFloorIndicator(floor);
        e.CurrentFloor = floor
    }
}

func (e *Elevator) HandleOrdersAtFloor(floor int, OrderCompleteTx chan MessageOrderComplete) bool {
     // Update the current floor
    var ordersDone []Order // Number of orders done

    
    for button := 0; button < numButtons; button++ {
        if e.LocalOrderArray[button][floor] == True { // If there is an active order at the floor

            if floor == bestOrder.Floor {

                if e.CurrentDirection == Up && button == HallUp {
                    Order := Order{floor, HallUp}
                    ordersDone = append(ordersDone, Order)
                    // HallUp order, and the elevator is going up (take order)
                    continue 
                } 

                if (e.CurrentDirection == Up && button == HallDown) && (e.LocalOrderArray[HallUp][floor] == False) {
                    check := e.CheckHallOrdersAbove(floor)
                    fmt.Println("Check above ( ElevatorAtFloor() ): ", check)
                    if check.Button == elevio.ButtonType(button) && check.Floor == floor { // There are no orders above the current floor
                        Order := Order{floor, HallDown}
                        ordersDone = append(ordersDone, Order) // Update the local order array
                        // HallDown order, and the elevator is going up (take order)
                        continue
                    }
                }

                if e.CurrentDirection == Down && button == HallDown {
                    Order := Order{floor, HallDown}
                    ordersDone = append(ordersDone, Order) // Update the local order array
                    // HallDown order, and the elevator is going down (take order)
                    continue
                }

                if (e.CurrentDirection == Down && button == HallUp) && (e.LocalOrderArray[HallDown][floor] == False) {
                    check := e.CheckHallOrdersBelow(floor)
                    if check.Button == elevio.ButtonType(button) && check.Floor == floor { // There are no orders below the current floor
                        Order := Order{floor, HallUp}
                        ordersDone = append(ordersDone, Order) // Update the local order array
                        // HallUp order, and the elevator is going down (take order)
                        continue
                    }
                }

            }

            if button == Cab {
                fmt.Println("Cab order at floor: ", floor)
                Order := Order{floor, Cab}
                ordersDone = append(ordersDone, Order) // Update the local order array
                // Cab order (take order)
                continue
            }


        }  
    }
    if len(ordersDone) > 0 {
        for i := 0; i < len(ordersDone); i++ {
            fmt.Println("Order done: ", ordersDone[i])
            e.UpdateOrderSystem(ordersDone[i]) // Update the local order array
            floor := ordersDone[i].Floor
            button := ordersDone[i].Button
            LocallyCompletedOrders[floor][button] = True
        }

        OrderCompleteTx <- MessageOrderComplete{Type: "OrderComplete", 
                                                Orders: ordersDone, 
                                                E: *e,
                                                FromElevatorID: e.ID}


        

        return true // There are active orders at the floor

    } else {

        return false // There are no active orders at the floor
    }

}


func (e *Elevator) HandleButtonEvent(newOrderTx chan MessageNewOrder, orderCompleteTx chan MessageOrderComplete, newOrder Order) {
    if button == Cab {

        newOrderTx <- MessageNewOrder{Type: "MessageNewOrder", NewOrder: newOrder, E: myElevator, ToElevatorID: myElevator.ID}

        if e.CheckIfOrderIsActive(newOrder) { // Check if the order is active
            if bestOrder.Floor == e.CurrentFloor {
                e.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
            } else {
                e.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
            }

        } else {

            e.ProcessElevatorOrders(newOrder, orderCompleteTx)

        }
    } else {

        if e.isMaster {
            // Handle order locally
            e.ProcessElevatorOrders(newOrder, orderCompleteTx)

        } else {

            newOrderTx <- MessageNewOrder{Type: "MessageNewOrder", NewOrder: newOrder, E: myElevator, ToElevatorID: myElevator.ID}
            e.ProcessElevatorOrders(newOrder, orderCompleteTx)

        }
    }
}

func (e *Elevator) ProcessElevatorOrders(newOrder Order, orderCompleteTx chan MessageOrderComplete) {

    e.UpdateOrderSystem(newOrder)

    amountOfOrders := e.CheckAmountOfActiveOrders()

    if amountOfOrders > 0 {

        bestOrder = e.ChooseBestOrder() // Choose the best order
        fmt.Println("Best order: ", bestOrder)

        if bestOrder.Floor == e.CurrentFloor {
            e.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Handle the elevator at the floor
        } else {
            e.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
        }
    } else {
        e.StopElevator()
    }
}


func (e* Elevator) HandleNewOrder(newOrder Order, fromElevator Elevator, toElevatorID int, orderCompleteTx chan MessageOrderComplete) {

    if e.isMaster && toElevatorID != e.ID {
        // Update global order system locally
        // Find the best elevator for the order
        // Send the order to the best elevator ( if hall order )
        //  newOrder := MessageNewOrder{
        //    Type:     "MessageNewOrder",
        //    NewOrder: newOrder,
        //    E: myElevator, // Use the correct field name as defined in your ElevatorStatus struct
        //    ToElevatorID: bestElevatorID,}
    } else if toElevatorID == e.ID && !e.isMaster {

        fmt.Println("New order received: ", newOrder)
        UpdateElevatorsOnNetwork(fromElevator)

        e.UpdateOrderSystem(newOrder) // Update the local order array

        e.PrintLocalOrderSystem()

        bestOrder = e.ChooseBestOrder() // Choose the best order
        fmt.Println("Best order: ", bestOrder)

        if bestOrder.Floor == e.CurrentFloor { // If the best order is at the current floor
            e.HandleElevatorAtFloor(bestOrder.Floor, orderCompleteTx) // Complete orders at the floor
        } else {
            e.DoOrder(bestOrder, orderCompleteTx) // Move the elevator to the best order
        }
    } else {
        // Order not meant for you. Just update the global order system
        UpdateElevatorsOnNetwork(fromElevator)

    }
}


func (e* Elevator) HandlePeersUpdate(p PeerUpdate, elevatorStatusTx chan ElevatorStatus) {

    fmt.Printf("Peer update:\n")
    fmt.Printf("  Peers:    %q\n", p.Peers)
    fmt.Printf("  New:      %q\n", p.New)
    fmt.Printf("  Lost:     %q\n", p.Lost)

    for _, peer := range p.Peers {
        found := false
        peerID, _ := strconv.Atoi(peer)
        for i, _ := range Elevators {
            if Elevators[i].ID == peerID {
                found = true
                Elevators[i].isActive = true
            }
        }

        if !found {
            elevatorStatusTx <- ElevatorStatus{
                Type: "ElevatorStatus",
                E:    myElevator,
            }
        }

    }

    for i, _ := range Elevators {
        for _, peer := range p.Lost {
            peerID, _ := strconv.Atoi(peer)
            if Elevators[i].ID == peerID {
                Elevators[i].isActive = false
                UpdateElevatorsOnNetwork(Elevators[i])
            }
        }
    }

    e.DetermineMaster() // Determine the master elevator
}