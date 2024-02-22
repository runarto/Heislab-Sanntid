package main

import (
    "encoding/json"
    "fmt"
    "net"
    "time"
    "sync"
    "log"
)


// Master function
func BroadcastGlobalOrderSystem(msg Message) {
    for {
        serializedMsg, err := msg.Serialize()
        if err != nil {
            log.Printf("Error serializing message: %v", err)
            continue // Skip this iteration and try again after the sleep period
        }

        err = SendToAddress(_broadcastAddr, serializedMsg)
        if err != nil {
            log.Printf("Error sending broadcast message: %v", err)
        }

        // Wait for 15 seconds before the next broadcast
        time.Sleep(15 * time.Second)
    }
}

func (e* Elevator) BroadcastElevatorInstance(msg Message) {
	serializedMsg, err := msg.Serialize()
	if err != nil {
		log.Printf("Error serializing message: %v", err)
		return
	}

    err = SendToAddress(_broadcastAddr, serializedMsg)
    if err != nil {
        log.Printf("Error sending broadcast message: %v", err)
    }

}