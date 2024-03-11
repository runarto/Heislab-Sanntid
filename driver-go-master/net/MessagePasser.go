package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessagePasser(messageSender <-chan interface{}, OrderCompleteTx chan utils.MessageOrderComplete,
	NewOrderTx chan utils.MessageNewOrder, ElevatorStatusTx chan utils.MessageElevatorStatus,
	MasterOrderWatcherTx chan utils.MessageOrderWatcher, LightsTx chan utils.MessageLights, ack chan utils.MessageConfirmed,
	OrderWatcher chan utils.OrderWatcher) {

	activeElevators := updater.GetActiveElevators()

	for {
		select {

		case newMsg := <-messageSender:

			switch m := newMsg.(type) {

			case utils.MessageOrderComplete:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageOrderComplete"
				OrderCompleteTx <- newMsg.(utils.MessageOrderComplete)
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageNewOrder:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageNewOrder"
				NewOrderTx <- newMsg.(utils.MessageNewOrder)
				fmt.Println("Sent a", m.Type, "message")
				go WaitForAck(newMsg, m.Type, activeElevators, ack, NewOrderTx, OrderWatcher)

			case utils.MessageElevatorStatus:
				m.Type = "MessageElevatorStatus"
				ElevatorStatusTx <- newMsg.(utils.MessageElevatorStatus)

			case utils.MessageOrderWatcher:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageOrderWatcher"
				MasterOrderWatcherTx <- newMsg.(utils.MessageOrderWatcher)

			case utils.MessageLights:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageLights"
				LightsTx <- newMsg.(utils.MessageLights)
				fmt.Println("Sent a", m.Type, "message")
				go WaitForAck(newMsg, m.Type, activeElevators, ack, LightsTx)

			}

			activeElevators = updater.GetActiveElevators()

		}

	}
}

func WaitForAck(msg interface{}, msgType string, activeElevators []int, ack chan utils.MessageConfirmed, channel ...interface{}) {
	var quit bool
	var responses map[int]bool

	timeout := 1 * time.Second
	responseTimer := time.NewTimer(timeout)

	for {
		select {
		case newMsg := <-ack:
			quit, responses = HandleConfirmation(newMsg, msgType, msg, responses, channel[0])
			if quit && responses == nil {
				return
			} else if quit && len(responses) == len(activeElevators)-1 {
				fmt.Println("All elevators have confirmed the message")
				return
			} else {
				responseTimer.Reset(timeout)
			}
		case <-responseTimer.C:
			fmt.Println("Response timeout, resending message")
			ResendMessage(msg, msgType, channel[0])
		}
	}
}

func response(responses map[int]bool, newMsg utils.MessageConfirmed) map[int]bool {
	responses[newMsg.FromElevatorID] = newMsg.Confirmed
	return responses
}

func ResendMessage(msg interface{}, msgType string, channel ...interface{}) {
	switch m := msg.(type) {
	case utils.MessageOrderComplete:
		m.Type = "MessageOrderComplete"
		ch, ok := channel[0].(chan utils.MessageOrderComplete)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
		fmt.Println("Resent a", m.Type, "message")
	case utils.MessageNewOrder:
		m.Type = "MessageNewOrder"
		ch, ok := channel[0].(chan utils.MessageNewOrder)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
		fmt.Println("Resent a", m.Type, "message")
	case utils.MessageElevatorStatus:
		m.Type = "MessageElevatorStatus"
		ch, ok := channel[0].(chan utils.MessageElevatorStatus)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
		fmt.Println("Resent a", m.Type, "message")
	case utils.MessageOrderWatcher:
		m.Type = "MessageOrderWatcher"
		ch, ok := channel[0].(chan utils.MessageOrderWatcher)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
		fmt.Println("Resent a", m.Type, "message")
	case utils.MessageLights:
		m.Type = "MessageLights"
		ch, ok := channel[0].(chan utils.MessageLights)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
		fmt.Println("Resent a", m.Type, "message")
	}
}

func HandleConfirmation(c utils.MessageConfirmed, msgType string, msg interface{}, responses map[int]bool, channel ...interface{}) (bool, map[int]bool) {
	fmt.Println("Received a MessageConfirmed message for", msgType, "from elevator", c.FromElevatorID)

	switch msgType {
	case "MessageNewOrder":
		if c.FromElevatorID == utils.MasterID && !utils.Master {
			fmt.Println("Order confirmed by master")
			OrderWatcher, ok := channel[0].(chan utils.OrderWatcher)
			if !ok {
				fmt.Println("Invalid channel type")
				return false, nil
			}
			OrderWatcher <- utils.OrderWatcher{
				Order:         msg.(utils.MessageNewOrder).NewOrder,
				ForElevatorID: c.FromElevatorID,
				IsComplete:    false,
				IsNew:         true,
				IsConfirmed:   true}

			return true, nil
		}
	case "MessageLights":
		if c.FromElevatorID != utils.ID {
			fmt.Println("Lights confirmed by elevator", c.FromElevatorID)
			responses = response(responses, c)
			return true, responses
		}
	}
	return false, nil
}
