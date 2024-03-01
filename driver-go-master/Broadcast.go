package main

import (
    "time"
)


func BroadcastElevatorStatus(elevator Elevator, statusTx chan ElevatorStatus) {

    if len(Elevators) > 1 {
        elevatorStatusMessage := ElevatorStatus{
            Type: "ElevatorStatus",
            E:    elevator, // Use the correct field name as defined in your ElevatorStatus struct
            AckStruct: ackStruct, 
        }

        // Initial broadcast
        statusTx <- elevatorStatusMessage

        // Optional: Periodic updates
        ticker := time.NewTicker(time.Second * 5)
        defer ticker.Stop()
        for range ticker.C {
            statusTx <- elevatorStatusMessage // Broadcast the current status
        }
    }
}

