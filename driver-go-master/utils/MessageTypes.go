package utils

import (
	"fmt"
)

type MessageGlobalOrderArrays struct { // Send periodically to update the global order system
	Type           string           `json:"type"` // Explicitly indicate the message type
	GlobalOrders   GlobalOrderArray `json:"globalOrders"`
	FromElevatorID int              `json:"fromElevatorID"` // The elevator that sent the order
}

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
	Type           string                            `json:"type"`           // A type identifier for decoding on the receiving end
	HallOrders     [2][NumFloors]HallAck             `json:"hallOrders"`     // The hall orders of the elevator
	CabOrders      [NumOfElevators][NumFloors]CabAck `json:"cabOrders"`      // The cab orders of the elevator
	FromElevatorID int                               `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageOrderConfirmed struct {
	Type           string `json:"type"`           // A type identifier for decoding on the receiving end
	Confirmed      bool   `json:"confirmed"`      // Whether or not the order was confirmed by the master
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
	ForOrder       Order  `json:"forOrder"`
}

type MessageLights struct {
	Type           string                          `json:"type"`           // A type identifier for decoding on the receiving end
	Lights         [NumButtons - 1][NumFloors]bool `json:"lights"`         // The lights of the elevator
	FromElevatorID int                             `json:"fromElevatorID"` // The elevator to send the order to
}

type MessageLightsConfirmed struct {
	Type           string `json:"type"`           // A type identifier for decoding on the receiving end
	Confirmed      bool   `json:"confirmed"`      // Whether or not the order was confirmed by the master
	FromElevatorID int    `json:"fromElevatorID"` // The elevator to send the order to
}

func PackMessage(msgType string, params ...interface{}) interface{} {
	switch msgType {
	case "MessageGlobalOrderArrays":
		msg := MessageGlobalOrderArrays{
			Type:           msgType,
			GlobalOrders:   params[0].(GlobalOrderArray),
			FromElevatorID: params[1].(int)}

		return msg
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
			HallOrders:     params[0].([2][NumFloors]HallAck),
			CabOrders:      params[1].([NumOfElevators][NumFloors]CabAck),
			FromElevatorID: params[2].(int)}

		return msg
	case "MessageOrderConfirmed":
		msg := MessageOrderConfirmed{
			Type:           msgType,
			Confirmed:      params[0].(bool),
			FromElevatorID: params[1].(int),
			ForOrder:       params[2].(Order)}

		return msg
	case "MessageLights":
		msg := MessageLights{
			Type:           msgType,
			Lights:         params[0].([NumButtons - 1][NumFloors]bool),
			FromElevatorID: params[1].(int)}

		return msg
	case "MessageLightsConfirmed":
		msg := MessageLightsConfirmed{
			Type:           msgType,
			Confirmed:      params[0].(bool),
			FromElevatorID: params[1].(int)}
		return msg
	}

	return nil
}

func HandleMessage(msg interface{}, params ...interface{}) {
	switch m := msg.(type) {
	case MessageGlobalOrderArrays:
		m.Type = "MessageGlobalOrderArrays"
		if ch, ok := params[0].(chan MessageGlobalOrderArrays); ok {
			ch <- m
			fmt.Println("Sent a", m.Type, "message")
		}
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
