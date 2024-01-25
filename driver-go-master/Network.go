package main

import (
    "encoding/json"
    "net"
    "fmt"
	"time"
)


// Master functions


// Probably an idea to periodically ping elevators on network to check if they are online. 
// Have a dynamic global list that maps elevators online to address. 
// This secures that elevators may rejoin the network, if they go offline. 

func ConnectToSlave(ElevatorAdress string) (*net.TCPConn, error) {
    tcpAddr, err := net.ResolveTCPAddr("tcp", ElevatorAdress)
    if err != nil {
        return nil, err
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        return nil, err
    }

    return conn, nil
}

func ReadAckFromConn(ConnectionToElevator net.Conn) (bool, error) {

	/*
	ReadAckFromConn reads an acknowledgment message from the given TCP connection and attempts to deserialize it from JSON to a boolean value.

	PARAM: ConnectionToElevator: net.Conn {The TCP connection from which the acknowledgment message is read. This connection should already be established with the elevator.}
	RETURN: bool {The deserialized acknowledgment message as a boolean value. True indicates a successful operation or acknowledgment, while false may indicate a failure or negative acknowledgment.}
	RETURN: error {An error object that encapsulates any issues encountered during the reading or deserialization process. Returns nil if the operation completes successfully without errors.}
	*/


    // Buffer to read the JSON acknowledgment
    ackBuffer := make([]byte, 1024)
    n, err := ConnectionToElevator.Read(ackBuffer)
    if err != nil {
        return false, fmt.Errorf("error reading acknowledgment: %w", err)
    }

    // Deserialize the JSON acknowledgment back to a boolean
    var ackMessage bool
    err = json.Unmarshal(ackBuffer[:n], &ackMessage)
    if err != nil {
        return false, fmt.Errorf("error deserializing acknowledgment: %w", err)
    }

    return ackMessage, nil
}


func SendOrder(ConnectionToElevator *net.TCPConn, order activeOrder, maxRetries int) bool {

	/*
	SendOrder attempts to send an order to the specified elevator connection. It retries up to maxRetries
	times in case of failure.

	PARAM: ConnectionToElevator: *net.TCPConn {The TCP connection to the elevator in question.}
	PARAM: order: activeOrder {The order which is being processed.}
	PARAM: maxRetries: int {The maximum number of attempts to send the order before giving up.}
	RETURN: bool {Returns true if the message is sent successfully, false if it fails after maxRetries attempts.
				  A False value would likely imply that the elevator is offline.}
	*/



    for attempt := 0; attempt < maxRetries; attempt++ {
        // Try to send the order
        if err := attemptToSendOrder(ConnectionToElevator, order); err != nil {
            fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
            time.Sleep(time.Second * time.Duration(attempt+1)) // Exponential back-off could be applied here
            continue
        }

        // If send is successful, break from the loop
        fmt.Println("Order successfully received by the elevator.")
        return true
    }

    fmt.Println("failed to send order after %d retries", maxRetries)
	return false
}


func attemptToSendOrder(ConnectionToElevator *net.TCPConn, order activeOrder) error {

	/*
	PARAM: ConnectionToElevator: *net.TCPConn {The TCP connection to the elevator in question.}
	PARAM: order: activeOrder {The order which is being processed.}
	RETURN: error {Returns err if something went wrong, nil if the message is sent without issues.}
	*/


    orderInJson, err := json.Marshal(order)
    if err != nil {
        return err
    }

    _, err = ConnectionToElevator.Write(orderInJson)
    if err != nil {
        return err
    }

    ackMessage, err := ReadAckFromConn(ConnectionToElevator)
    if err != nil {
        return err
    }

    if !ackMessage {
        return fmt.Errorf("negative acknowledgment received")
    }

    return nil
}


func PingAddress(ElevatorAdress string, timeout time.Duration) bool {

	/*
	PARAM: ElevatorAddress: string {The address of the elevator on the form ip_address:port}
	PARAM: timeout: time.Duration {The amount of time you are attempting to ping the address}
	RETURN: bool {True implies the elevator is online, false implies it is offline.}
	*/



    ConnectionToElevator, err := net.DialTimeout("tcp", ElevatorAdress, timeout)
    if err != nil {
        fmt.Printf("Ping failed: %v\n", err)
        return false
    }
    defer ConnectionToElevator.Close() 
    fmt.Println("Ping successful:", ElevatorAdress)
    return true
}




// Slave functions 

func StartTCPServer(ElevatorAdress string) error {
    listener, err := net.Listen("tcp", ElevatorAdress)
    if err != nil {
        return err
    }
    defer listener.Close()

    for {
		ConnectionToElevator, err := listener.Accept()
        if err != nil {
            fmt.Println("Accept error:", err)
            continue
        }
        go handleConnection(ConnectionToElevator)
    }
}

func handleConnection(ConnectionToElevator net.Conn) {
    defer ConnectionToElevator.Close()

    var buffer [1024]byte
    n, err := ConnectionToElevator.Read(buffer[:])
    if err != nil {
        fmt.Println("Read error:", err)
        return
    }

    var order activeOrder
    err = json.Unmarshal(buffer[:n], &order)
    if err != nil {
        fmt.Println("Unmarshal error:", err)
        return
    }

	ackMessage, err := json.Marshal(true)
    if err != nil {
        fmt.Println("Error serializing acknowledgment:", err)
        return
    }

    // Send the serialized JSON acknowledgment
    _, err = ConnectionToElevator.Write(ackMessage)
    if err != nil {
        fmt.Println("Error sending acknowledgment:", err)
        return
    }

	//processOrder(order) //TODO
}