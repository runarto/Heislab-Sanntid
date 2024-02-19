package main

import (
    "encoding/json"
    "fmt"
    "net"
    "time"
    "sync"
    "log"
)

var (
    connections map[string]*net.UDPConn // Map of UDP connections (i.e other elevators)
    mutex       sync.Mutex // To safely access the connections map
)


type Message interface {
    Serialize() ([]byte, error)
}


// InitializeConnection sets up a UDP connection to the specified address.
func InitializeConnection(address string) (*net.UDPConn, error) {
    udpAddr, err := net.ResolveUDPAddr("udp", address)

    if err != nil {
        return nil, fmt.Errorf("resolving UDP address failed: %v", err)
    }

    conn, err := net.DialUDP("udp", nil, udpAddr)

    if err != nil {
        return nil, fmt.Errorf("dialing UDP connection failed: %v", err)
    }

    mutex.Lock()
    connections[address] = conn
    mutex.Unlock()

    return conn, nil
}

// GetConnection retrieves an existing UDP connection or creates a new one.
func GetConnection(address string) (*net.UDPConn, error) {
    mutex.Lock()
    defer mutex.Unlock()

    conn, err := PingElevator(address) // Check if connection is already established
    if err == nil {
        return conn, err // Connection already exists, return connection
    }

    // No existing connection found, so initialize a new one
    return InitializeConnection(address) 
}


func (m *MessageGlobalOrder) Serialize() ([]byte, error) {
    bytes, err := json.Marshal(m)
    if err != nil {
        return nil, err
    }
    return append([]byte{0x01}, bytes...), nil
}

func (m *MessageNewOrder) Serialize() ([]byte, error) {
    bytes, err := json.Marshal(m)
    if err != nil {
        return nil, err
    }
    return append([]byte{0x02}, bytes...), nil
}

func (m *MessageOrderComplete) Serialize() ([]byte, error) {
    bytes, err := json.Marshal(m)
    if err != nil {
        return nil, err
    }
    return append([]byte{0x03}, bytes...), nil
}


func BroadcastGlobalOrderSystem(msg Message) {
    for {
        serializedMsg, err := msg.Serialize()
        if err != nil {
            log.Printf("Error serializing message: %v", err)
            continue // Skip this iteration and try again after the sleep period
        }

        err = sendToAddress(broadcastAddr, serializedMsg)
        if err != nil {
            log.Printf("Error sending broadcast message: %v", err)
        }

        // Wait for 15 seconds before the next broadcast
        time.Sleep(15 * time.Second)
    }
}




func SendMessage(conn *net.UDPConn, msg Message, timeout time.Duration) error {
    serializedMsg, err := msg.Serialize()
    if err != nil {
        return err
    }

    // Send the serialized message
    _, err = conn.Write(serializedMsg)
    if err != nil {
        return err
    }

    // Wait for an acknowledgment
    err = WaitForAck(conn, timeout)
    return err
}


func WaitForAck(conn *net.UDPConn, timeout time.Duration) error {
    // Set a deadline for reading; this is how long to wait for the ack
    err := conn.SetReadDeadline(time.Now().Add(timeout))
    if err != nil {
        return fmt.Errorf("error setting read deadline: %v", err)
    }

    // Buffer to store the incoming message
    buffer := make([]byte, 1024) // Adjust the size according to your needs

    // Attempt to read the acknowledgment message
    _, _, err = conn.ReadFromUDP(buffer)
    if err != nil {
        return fmt.Errorf("error reading ack message: %v", err)
    }

    if buffer[0] != 1 {
        return fmt.Errorf("received message is not an ack")
    }

    // Here you should add logic to validate that the received message is indeed an ack
    // For simplicity, this example assumes any received message is an ack

    return nil
}




func HandleMessage(conn *net.UDPConn) {

    buffer := make([]byte, 1024)
    n, addr, err := conn.ReadFromUDP(buffer)
    if err != nil {
        log.Fatal(err)
    }

    localAddr := conn.LocalAddr().(*net.UDPAddr)
    messageType := buffer[0]
    messageBytes := buffer[1:n]

    if addr.IP.Equal(localAddr.IP) && addr.Port == localAddr.Port {
        fmt.Println("Received message from self, ignoring")
        return
    }

    switch messageType {
    case 0x01:
        var msg MessageGlobalOrder
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }
        // Handle global order update
        globalOrders = msg.globalOrders // Update global order system

    case 0x02:
        var msg MessageNewOrder
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }
        newOrder = msg.newOrder
        fromElevator = msg.e

        // Logic for adding fromElevator to ActiveElevators, overwriting if already exists
        // Handle new order

    case 0x03:
        var msg MessageOrderComplete
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }

        OrderComplete = msg.order
        fromElevator = msg.e

        // Logic for adding fromElevator to ActiveElevators, overwriting if already exists
        // Handle updating global order system


        // Add more as needed, only need to create Serialize() methods for new message types
        // Handle order completion
    }

    if err := SendAck(conn, remoteAddr); err != nil {
        log.Printf("Failed to send ack: %v", err)
    }

}

func SendAck(conn *net.UDPConn, addr *net.UDPAddr) error {
    ackMessage := []byte{1} // Ack message is a single byte set to 1
    _, err := conn.WriteToUDP(ackMessage, addr)
    return err
}

