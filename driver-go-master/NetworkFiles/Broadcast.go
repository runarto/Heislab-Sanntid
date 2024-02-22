package NetworkFiles

import (
    "github.com/runarto/Heislab-Sanntid"
    "github.com/runarto/Heislab-Sanntid/elevio"
    "encoding/json"
    "fmt"
    "net"
    "time"
    "sync"
    "log"
)

func (e * Elevator) ElevatorIsAlive() {

    msg := MessageAlive{"I am alive", e.ElevatorID}

    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        serializedMsg, err := msg.Serialize()
        if err != nil {
            log.Printf("Error serializing message: %v", err)
            continue
        }

        err = SendToAddress(_broadcastAddr, serializedMsg)
        if err != nil {
            log.Printf("Error sending broadcast message: %v", err)
        }
    }
}


// Master function
func BroadcastGlobalOrderSystem(msg Message) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        serializedMsg, err := msg.Serialize()
        if err != nil {
            log.Printf("Error serializing message: %v", err)
            continue
        }

        err = SendToAddress(_broadcastAddr, serializedMsg)
        if err != nil {
            log.Printf("Error sending broadcast message: %v", err)
        }
    }
}

func (e *Elevator) BroadcastElevatorInstance(msg Message) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        serializedMsg, err := json.Marshal(msg) // Assuming msg.Serialize() effectively performs JSON serialization
        if err != nil {
            log.Printf("Error serializing message: %v", err)
            continue
        }

        err = SendToAddress(_broadcastAddr, serializedMsg)
        if err != nil {
            log.Printf("Error sending broadcast message: %v", err)
        }
    }
}


func SetUpBroadcastListener() (*net.UDPConn, error) {
    // Resolve the address for listening to UDP broadcasts on the given port
    localAddr, err := net.ResolveUDPAddr("udp", ":"+_ListeningPort)
    if err != nil {
        return nil, fmt.Errorf("error resolving UDP address: %v", err)
    }

    // Listen for incoming UDP broadcasts
    conn, err := net.ListenUDP("udp", localAddr)
    if err != nil {
        return nil, fmt.Errorf("error setting up UDP listener: %v", err)
    }

    return conn, nil
}

