package utils

import (
	"encoding/json"
	"fmt"
)

type MessageOrderComplete struct { // Send when an order is completed
	Type  string `json:"type"` // Explicitly indicate the message type
	Order Order  `json:"orders"`
}

type MessageNewOrder struct { // Send when a new order is received
	Type     string `json:"type"` // Explicitly indicate the message type
	NewOrder Order  `json:"newOrder"`
}

type MessageElevatorStatus struct {
	Type     string   `json:"type"`     // A type identifier for decoding on the receiving end
	Elevator Elevator `json:"elevator"` // The Elevator instance
}

type MessageOrderWatcher struct {
	Type   string                                 `json:"type"`   // A type identifier for decoding on the receiving end
	Orders map[string][NumButtons][NumFloors]bool `json:"orders"` // The orders the master has received
}

type MessageLights struct {
	Type   string             `json:"type"`   // A type identifier for decoding on the receiving end
	Lights [2][NumFloors]bool `json:"lights"` // The lights that the master is sending
}

type MessageConfirmed struct {
	Type      string `json:"type"`      // A type identifier for decoding on the receiving end
	Hashvalue int    `json:"hashvalue"` // The elevator to send the order to
}

type Message struct {
	Type           string      `json:"type"`           // A type identifier for decoding on the receiving end
	ToElevatorID   int         `json:"toElevatorID"`   // The elevator to send the order to
	FromElevatorID int         `json:"fromElevatorID"` // The elevator to send the order to
	Msg            interface{} `json:"msg"`            // The message to send
}

// PackMessage packs the given parameters into a Message struct based on the provided message type.
// It returns the packed Message.
func PackMessage(msgType string, params ...interface{}) Message {

	switch msgType {
	case "MessageOrderComplete":
		msg := Message{
			Type:           msgType,
			ToElevatorID:   params[0].(int),
			FromElevatorID: params[1].(int),
			Msg:            MessageOrderComplete{Type: "MessageOrderComplete", Order: params[2].(Order)}}

		return msg

	case "MessageNewOrder":
		msg := Message{
			Type:           msgType,
			ToElevatorID:   params[0].(int),
			FromElevatorID: params[1].(int),
			Msg:            MessageNewOrder{Type: "MessageNewOrder", NewOrder: params[2].(Order)}}

		return msg

	case "MessageConfirmed":
		msg := Message{
			Type:           msgType,
			ToElevatorID:   params[0].(int),
			FromElevatorID: params[1].(int),
			Msg:            MessageConfirmed{Type: "MessageConfirmed", Hashvalue: params[2].(int)}}

		return msg

	case "ElevatorStatus":
		msg := Message{
			Type:           msgType,
			ToElevatorID:   params[0].(int),
			FromElevatorID: params[1].(int),
			Msg:            MessageElevatorStatus{Type: "ElevatorStatus", Elevator: params[2].(Elevator)},
		}
		return msg

	case "MessageLights":
		msg := Message{
			Type: msgType,
			Msg:  MessageLights{Type: "MessageLights", Lights: params[0].([2][NumFloors]bool)}}

		return msg

	}

	return Message{}
}

// DecodeMessage decodes a message based on its type.
// It takes a Message struct and a msgType string as input.
// The function serializes the interface{} to JSON and then deserializes it into the specific struct based on the msgType.
// It returns the decoded Message struct.
func DecodeMessage(msg Message, msgType string) Message {
	// First, serialize the interface{} to JSON
	jsonData, err := json.Marshal(msg.Msg)
	if err != nil {
		fmt.Println("Error marshalling interface{}:", err)
		return Message{}
	}

	// Now, deserialize JSON into the specific struct based on msgType
	switch msgType {
	case "MessageNewOrder":
		var newOrder MessageNewOrder
		if err := json.Unmarshal(jsonData, &newOrder); err != nil {
			fmt.Println("Error unmarshalling to MessageNewOrder:", err)
			return Message{}
		}
		return Message{
			Type:           "MessageNewOrder",
			ToElevatorID:   msg.ToElevatorID,
			FromElevatorID: msg.FromElevatorID,
			Msg:            newOrder,
		}

	case "MessageOrderComplete":
		var orderComplete MessageOrderComplete
		if err := json.Unmarshal(jsonData, &orderComplete); err != nil {
			fmt.Println("Error unmarshalling to MessageOrderComplete:", err)
			return Message{}
		}
		return Message{
			Type:           "MessageOrderComplete",
			ToElevatorID:   msg.ToElevatorID,
			FromElevatorID: msg.FromElevatorID,
			Msg:            orderComplete,
		}

	default:
		fmt.Println("Unsupported message type:", msgType)
		return Message{}
	}
}
