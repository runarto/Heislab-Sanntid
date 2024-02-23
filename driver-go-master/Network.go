package main

import (
    "encoding/json"
    "fmt"
    "net"
    "time"
    "sync"
    "log"
)



type Data struct {
    Message []byte
    Address *net.UDPAddr
}


type Message interface {
    Serialize() ([]byte, error)
}




func (m *MessageGlobalOrderArray) Serialize() ([]byte, error) {
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

func (m *MessageAlive) Serialize() ([]byte, error) {
    bytes, err := json.Marshal(m)
    if err != nil {
        return nil, err
    }
    return append([]byte{0x05}, bytes...), nil
}


func SendMessage(address string, msg Message) error {
    serializedMsg, err := msg.Serialize()
    conn, err := net.Dial("udp", address)
    if err != nil {
        return err
    }
    _, err = conn.Write(serializedMsg)
    if err != nil {
        return err
    }

    err = WaitForAck(conn, 1*time.Second)
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




func HandleMessage(conn *net.UDPConn, receiver <- chan Data) {

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

    receiver <- Data{messageBytes, addr}

}



func SendAck(conn *net.UDPConn, addr *net.UDPAddr) error {
    ackMessage := []byte{1} // Ack message is a single byte set to 1
    _, err := conn.WriteToUDP(ackMessage, addr)
    return err
}


func PingElevator() (*net.UDPConn, error) {
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