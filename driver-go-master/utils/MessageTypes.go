package utils

import (
	"fmt"
)

type MessageOrderComplete struct { // Send when an order is completed
	Type           string `json:"type"` // Explicitly indicate the message type
	Order          Order  `json:"orders"`
	ToElevatorID   int    `json:"toElevatorID"`
	FromElevatorID int    `json:"fromElevatorID"` // The elevator that completed the order
}

type MessageNewOrder struct { // Send when a new order is received
	Type           string `json:"type"` // Explicitly indicate the message type
	NewOrder       Order  `json:"newOrder"`
	ToElevatorID   int    `json:"toElevatorID"`
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageElevatorStatus struct {
	Type         string   `json:"type"`     // A type identifier for decoding on the receiving end
	FromElevator Elevator `json:"elevator"` // The Elevator instance
}

type MessageOrderWatcher struct {
	Type           string                          `json:"type"`           // A type identifier for decoding on the receiving end
	HallOrders     [2][NumFloors]bool              `json:"hallOrders"`     // The hall orders of the elevator
	CabOrders      [NumOfElevators][NumFloors]bool `json:"cabOrders"`      // The cab orders of the elevator
	FromElevatorID int                             `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageOrderConfirmed struct {
	Type           string `json:"type"`           // A type identifier for decoding on the receiving end
	Confirmed      bool   `json:"confirmed"`      // Whether or not the order was confirmed by the master
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
	ForOrder       Order  `json:"forOrder"`
}

type MessageLights struct {
	Type           string             `json:"type"`           // A type identifier for decoding on the receiving end
	Lights         [2][NumFloors]bool `json:"lights"`         // The lights of the elevator
	FromElevatorID int                `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageLightsConfirmed struct {
	Type           string `json:"type"`           // A type identifier for decoding on the receiving end
	Confirmed      bool   `json:"confirmed"`      // Whether or not the order was confirmed by the master
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageConfirmed struct {
	Type           string `json:"type"`           // A type identifier for decoding on the receiving end
	Msg            string `json:"msg"`            // The message to be confirmed
	Confirmed      bool   `json:"confirmed"`      // Whether or not the order was confirmed by the master
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
}

func PackMessage(msgType string, params ...interface{}) interface{} {
	switch msgType {
	case "MessageOrderComplete":
		msg := MessageOrderComplete{
			Type:           msgType,
			Order:          params[0].(Order),
			ToElevatorID:   params[1].(int),
			FromElevatorID: params[2].(int)}

		return msg
	case "MessageNewOrder":
		msg := MessageNewOrder{
			Type:           msgType,
			NewOrder:       params[0].(Order),
			ToElevatorID:   params[1].(int),
			FromElevatorID: params[2].(int)}

		return msg
	case "MessageElevatorStatus":
		msg := MessageElevatorStatus{
			Type:         msgType,
			FromElevator: params[0].(Elevator)}

		return msg
	case "MessageOrderWatcher":
		msg := MessageOrderWatcher{
			Type:           msgType,
			HallOrders:     params[0].([2][NumFloors]bool),
			CabOrders:      params[1].([NumOfElevators][NumFloors]bool),
			FromElevatorID: params[2].(int)}

		return msg

	case "MessageLights":
		msg := MessageLights{
			Type:           msgType,
			Lights:         params[0].([2][NumFloors]bool),
			FromElevatorID: params[1].(int)}

		return msg

	case "MessageConfirmed":
		msg := MessageConfirmed{
			Type:           msgType,
			Msg:            params[0].(string),
			Confirmed:      params[1].(bool),
			FromElevatorID: params[2].(int)}
		return msg
	}

	return nil
}

func HandleMessage(msg interface{}, params ...interface{}) {
	switch m := msg.(type) {
	case MessageOrderComplete:
		m.Type = "MessageOrderComplete"
		if ch, ok := params[0].(chan MessageOrderComplete); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	case MessageNewOrder:
		m.Type = "MessageNewOrder"
		if ch, ok := params[0].(chan MessageNewOrder); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	case MessageElevatorStatus:
		m.Type = "MessageElevatorStatus"
		if ch, ok := params[0].(chan MessageElevatorStatus); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	case MessageOrderWatcher:
		m.Type = "MessageOrderWatcher"
		if ch, ok := params[0].(chan MessageOrderWatcher); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	case MessageOrderConfirmed:
		m.Type = "MessageOrderConfirmed"
		if ch, ok := params[0].(chan MessageOrderConfirmed); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	case MessageLights:
		m.Type = "MessageLights"
		if ch, ok := params[0].(chan MessageLights); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
	}
}
