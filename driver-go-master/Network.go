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

    conn, err := PingElevator(address)
    if err == nil {
        return conn, err
    }

    // No existing connection found, so initialize a new one
    return InitializeConnection(address)
}


// SendOrder sends an order to the specified address over UDP.
func SendOrder(address string, order Order) error {
    conn, err := GetConnection(address)
    if err != nil {
        return err
    }

    orderBytes, err := json.Marshal(order)
    if err != nil {
        return fmt.Errorf("error serializing order: %v", err)
    }

    _, err = conn.Write(orderBytes)
    if err != nil {
        return fmt.Errorf("sending message failed: %v", err)
    }

    fmt.Printf("Message sent to %v: %s\n", address, string(orderBytes))
    return nil

    err = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		fmt.Printf("Error setting read deadline: %v\n", err)
		return err
	}

	// Buffer for reading incoming messages
	buffer := make([]byte, 1024)

	// Listen for responses
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Error reading from UDP: %v\n", err)
			break // Exit the loop on error (including timeout)
		}
		fmt.Printf("Received message from %v: %s\n", addr, string(buffer[:n]))
        if string(buffer[:n]) == "ack" {
            return nil
        }
	}
    return nil
}


// readOrder listens on a given port for incoming UDP datagrams and processes them as activeOrder structs.
// Will be running continiously, and will be listening for orders from other elevators.
// Note, must implement another function for handling the global order arrays, or local order arrays. 
// Or perhaps implement a case switch for handling the different types of orders.

func ReadOrder(port string, receiver chan<- Order)  {
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

    fmt.Printf("Listening for orders on UDP port %s\n", port)

    // Buffer for reading incoming messages
    buffer := make([]byte, 1024)

    for {
        // Read a UDP datagram
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            log.Printf("Error reading from UDP: %v", err)
            continue // Skip to the next iteration on error
        }

        // Deserialize the datagram into an activeOrder struct
        var order Order
        if err := json.Unmarshal(buffer[:n], &order); err != nil {
            log.Printf("Error deserializing order from %v: %v", addr, err)
            continue // Skip to the next iteration on error
        }

        // Process the order here (this is just an example print statement)
        fmt.Printf("Received order %+v from %v\n", order, addr)
        ackMessage := []byte("ack")
        if _, err := conn.WriteToUDP(ackMessage, addr); err != nil {
            log.Printf("Error sending ack to %v: %v", addr, err)
        } else {
            fmt.Printf("Ack sent to %v\n", addr)
        }

        receiver <- order


    }
}

// Pings an elevator at the specified address and returns the connection if successful.
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


