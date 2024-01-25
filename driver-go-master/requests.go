package main


import (
    "fmt"
    "Heislab-Sanntid/elevio"
)


func initializeQueue() {
    for i := range OrderArray {
        OrderArray[i][0] = NotDefined // etasje
        OrderArray[i][1] = NotDefined // Opp/Ned (0/1) Opp = 0, Ned = 1
        OrderArray[i][2] = NotDefined // cab (True/False)
    }
}


func printOrderArray() {
    for i := 0; i < MaxOrders; i++ {
        for j := 0; j < 3; j++ {
            if (OrderArray[i][0] != NotDefined) {
                fmt.Println(OrderArray[i][j])
                }
        }
    }
}

// fromFloor: etasje input kommer fra, button: type (opp, ned, cab)

func addToQueueFromFloorPanel(fromFloor int, Direction int) {
    for i := 0; i < MaxOrders; i++ {
        if (OrderArray[i][0] == fromFloor) &&
           (OrderArray[i][1] == Direction) {
            return
           }

        if (OrderArray[i][0] == NotDefined) && 
           (OrderArray[i][1] == NotDefined) && 
           (OrderArray[i][2] == NotDefined) {
            OrderArray[i][0] = fromFloor
            OrderArray[i][1] = Direction
            OrderArray[i][2] = False
            return
        }
    }
}

func addToQueueCab(toFloor int) {
    for i := 0; i < MaxOrders; i++ {

        if (OrderArray[i][0] == toFloor) &&
           (OrderArray[i][2] == True) {
         return
        }


        if (OrderArray[i][0] == NotDefined) && 
           (OrderArray[i][1] == NotDefined) && 
           (OrderArray[i][2] == NotDefined) {
            OrderArray[i][0] = toFloor

            if (currentFloor < toFloor) || 
               ((currentFloor == toFloor) && 
                (CurrentDirectionAlt == Down)) {
                OrderArray[i][1] = CabUp
                OrderArray[i][2] = True
                fmt.Println("Order added successfully, up")
                return
            }

            if (currentFloor > toFloor) || 
               (currentFloor == toFloor && 
                CurrentDirectionAlt == Up) {
                OrderArray[i][1] = CabDown
                OrderArray[i][2] = True
                fmt.Println("Order added successfully, down")
                return
            }
        }
    }
}

/* func checkNewOrders() {
	for f := 0; f < numFloors; f++ {
		for b := 0; b < numButtons; b++ {
			btnPressed := elevio.GetButton(elevio.ButtonType(b),f)
			if btnPressed && f != elevio.GetFloor() {
				fmt.Println("Button was pressed");
				elevio.SetButtonLamp(elevio.ButtonType(b), f, On)
				if b == elevio.BT_Cab {
					addToQueueCab(f)
					fmt.Println("Order added from inside cab")
				} else {
					addToQueueFromFloorPanel(f, elevio.ButtonType(b))
					fmt.Println("Order added from outside elevator")
				}
			}
		}
	}
} */


func checkOrderCompletion() int {
    completedOrders := 0
    numOfOrders := amountOfOrders()

    if numOfOrders == 0 {
        return 0
    }

    for i := 0; i < MaxOrders; i++ {
        orderFloor := OrderArray[i][0]
        direction := OrderArray[i][1]   // 0 for up, 1 for down
        fromCab := OrderArray[i][2]     // 1 if from cab, 0 otherwise

        if (orderFloor == NotDefined) {
            continue
        }

        nextDirection := getNextMotorDirection(i) //Down
        fmt.Println("The next direction is", nextDirection)


        if numOfOrders == 1 {
            if (fromCab == False && currentFloor == orderFloor) {
                print("here")
                processOrder(i, orderFloor, direction)
                completedOrders++
                return completedOrders
            } else if (currentFloor == orderFloor) {
                processOrder(i, orderFloor, 2)
                completedOrders++
                return completedOrders
            }
        }

        if ((fromCab == True && currentFloor == orderFloor)) {
            processOrder(i, orderFloor, 2) // Process the cab order
            completedOrders++
            fmt.Println("Order completed from cab.")
        }

        if (fromCab == False && currentFloor == orderFloor) {

            if ( (direction == Up && CurrentDirection == ElevUp) || (direction == Down && nextDirection == Down)) {
                processOrder(i, orderFloor, direction) // Process the external order
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
            }

            if (direction == Down && CurrentDirection == ElevDown || (direction == Up && nextDirection == Up)) {
                processOrder(i, orderFloor, direction) // Process the external order
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
            }

            if ( (currentFloor == 3 && direction == Up) || (currentFloor == 0 && direction == Down)) {
                processOrder(i, orderFloor, direction) // Process the external order
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
            }




        }



/*         fmt.Println("Next direction is", nextDirection)
        

        // Process orders from the cab
        if fromCab == True && currentFloor == orderFloor && direction == nextDirection {
            processOrder(i, orderFloor, 2) // Process the cab order
            completedOrders++
            fmt.Println("Order completed from cab.")
        }

        // Process external orders (from outside the elevator)
        if fromCab == False && currentFloor == orderFloor {
            fmt.Println("here")
        
            if direction == nextDirection || CurrentState == Still {
                processOrder(i, orderFloor, direction) // Process the external order
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
                
            }
            if (currentFloor == 3 && nextDirection == Down) {
                processOrder(i, orderFloor, 1) //Hvis du er i tredje etasje, skal du ned
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
                
            }
    
            if (currentFloor == 0 && nextDirection == Up) {
                processOrder(i, orderFloor, 0)
                completedOrders++
                fmt.Println("Order completed from floor")
                continue
                
            }
        

    } */
    
    }
    fmt.Println("Order complete")
    return completedOrders
}


    
func processOrder(index int, floor int, direction int) {
        elevio.SetButtonLamp(elevio.ButtonType(direction), floor, Off)
        removeOrder(index)
    }



func removeOrder(index int) {
    // Set all elements of the order to an 'empty' or 'inactive' state
    OrderArray[index][0] = NotDefined // Assuming -1 indicates an inactive order
    OrderArray[index][1] = NotDefined
    OrderArray[index][2] = NotDefined
}

func amountOfOrders() int {
    orderAmount := 0
    for i := 0; i < MaxOrders; i++ {
        if OrderArray[i][0] == NotDefined &&
           OrderArray[i][1] == NotDefined &&
           OrderArray[i][2] == NotDefined {
            continue
        } else {
            orderAmount++
        }
    }
    return orderAmount
}

func elevatorDirection() elevio.MotorDirection {
    if IsDoorOpen == Close && amountOfOrders() > 0 {
        fmt.Println("Door is open, and amount of orders is greater than one.")
        if CurrentDirection == ElevUp {
            fmt.Println("Last direction was up. ")
            for i := 0; i < MaxOrders; i++ {
                if OrderArray[i][0] > currentFloor && OrderArray[i][0] != NotDefined {
                    fmt.Println("Elevator go up")
                    return elevio.MD_Up
                } else if OrderArray[i][0] < currentFloor && OrderArray[i][0] != NotDefined {
                    fmt.Println("Elevator go down")
                    return elevio.MD_Down
                }
            }
        } else if CurrentDirection == ElevDown {
            fmt.Println("Last direction was down.")
            for i := 0; i < MaxOrders; i++ {
                if OrderArray[i][0] < currentFloor && OrderArray[i][0] != NotDefined {
                    fmt.Println("Elevator go down")
                    return elevio.MD_Down
                } else if OrderArray[i][0] > currentFloor && OrderArray[i][0] != NotDefined {
                    fmt.Println("Elevator go up")
                    return elevio.MD_Up
                }
            }
        } 
    }
    fmt.Println("oops, elevator has no orders")
    return elevio.MD_Stop
}



func Obstruction() {
    for elevio.GetObstruction() {
        elevio.SetDoorOpenLamp(On)
    }

    elevio.SetDoorOpenLamp(Off)
}


func getNextMotorDirection(i int) int {
    if CurrentDirection == ElevUp && OrderArray[i+1][0] != -1 {
        if (OrderArray[i+1][0] > currentFloor) {
            return Up
        } else if (OrderArray[i+1][0] < currentFloor) {
            return Down
        } else {
            if (OrderArray[i+1][2] == True) {
                return Down
            } else {
                return Up
            }
        }
        // blablabla
    } else if (CurrentDirection == ElevDown && OrderArray[i+1][0] != -1) {
        if (OrderArray[i+1][0] < currentFloor) {
            return Down
        } else if (OrderArray[i+1][0] > currentFloor) {
            return Up    
        } else {
            if (OrderArray[i+1][2] == True) {
                return Up 
            } else {
                return Down
            }

        }
    }
    return -1
}
