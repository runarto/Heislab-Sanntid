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

    return conn, nil // Returns a connection to the specified address
}

//GetConnection retrieves an existing UDP connection or creates a new one.
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

func (m *MessageElevatorIDandIP) Serialize() ([]byte, error) {
    bytes, err := json.Marshal(m)
    if err != nil {
        return nil, err
    }
    return append([]byte{0x04}, bytes...), nil
}

func SendToAddress(address string, data []byte) error {
    conn, err := net.Dial("udp", address)
    if err != nil {
        return err
    }
    defer conn.Close()

    _, err = conn.Write(data)
    return err
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




func HandleMessage(conn *net.UDPConn, newOrderReceiver chan<- MessageNewOrder, OrderCompleteReceiver chan <- MessageOrderComplete) {

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
        globalOrderArray = msg.globalOrders // Update global order system
        resetCounter() // Reset the watchdog counter

    case 0x02:
        var msg MessageNewOrder
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }
        

        newOrderReceiver <- msg
		// Send message to channel

        

    case 0x03:
        var msg MessageOrderComplete
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }

    

        OrderCompleteReceiver <- msg
		// Send message to channel


    case 0x04:
        var msg MessageElevator
        if err := json.Unmarshal(messageBytes, &msg); err != nil {
            log.Fatal(err)
        }

        // Add elevator to list of active elevators

        
    }


    if err := SendAck(conn, addr); err != nil {
        log.Printf("Failed to send ack: %v", err)
    }

}

func SendAck(conn *net.UDPConn, addr *net.UDPAddr) error {
    ackMessage := []byte{1} // Ack message is a single byte set to 1
    _, err := conn.WriteToUDP(ackMessage, addr)
    return err
}


func PingElevator(address string) (*net.UDPConn, error) {
    if conn, exists := connections[address]; exists {

        pingMessage := []byte("ping")
        _, err := conn.Write(pingMessage)

        if err != nil {
            return nil, fmt.Errorf("error sending ping: %v", err)
        }

        // Set a read deadline for the pong response
        conn.SetReadDeadline(time.Now().Add(1 * time.Second))

        // Buffer for reading the pong response
        buffer := make([]byte, 1024)

        // Attempt to read the "pong" response
        _, _, err = conn.ReadFromUDP(buffer)
        if err != nil {
            // If there's an error (including timeout), consider the connection as not online
            return nil, fmt.Errorf("no pong response or error reading pong: %v", err)
        }

        // Reset the deadline
        conn.SetReadDeadline(time.Time{})

        // Assume a valid pong response means the connection is online
        return conn, nil
    }

    return nil, fmt.Errorf("no connection found for address %v", address)
}

// ReadPing listens on a given port for incoming UDP "ping" messages and responds with "ack".
func ReadPing(port string) {
    // Resolve the local UDP address
    localAddr, err := net.ResolveUDPAddr("udp", ":"+port)
    if err != nil {
        log.Fatalf("Error resolving UDP address: %v", err)
    }

    // Listen for incoming UDP datagrams
    conn, err := net.ListenUDP("udp", localAddr)
    if err != nil {
        log.Fatalf("Error setting up UDP listener: %v", err)
    }
    defer conn.Close()

    fmt.Printf("Listening for pings on UDP port %s\n", port)

    // Buffer for reading incoming messages
    buffer := make([]byte, 1024)

    for {
        // Read a UDP datagram
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            log.Printf("Error reading from UDP: %v", err)
            continue // Skip to the next iteration on error
        }

        // Log the received ping message
        fmt.Printf("Received ping from %v: %s\n", addr, string(buffer[:n]))

        // Send an "ack" message back to the sender
        ackMessage := []byte("pong")
        if _, err := conn.WriteToUDP(ackMessage, addr); err != nil {
            log.Printf("Error sending ack to %v: %v", addr, err)
        } else {
            fmt.Printf("Ack sent to %v\n", addr)
        }
    }
}