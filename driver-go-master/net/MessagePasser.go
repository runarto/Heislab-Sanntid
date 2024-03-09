package net

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessagePasser(msg <-chan interface{}, arrays chan utils.MessageGlobalOrderArrays, orderComplete chan utils.MessageOrderComplete,
	newOrder chan utils.MessageNewOrder, elevatorStatus chan utils.MessageElevatorStatus, orderWatcher chan utils.MessageOrderWatcher,
	orderConfirmed chan utils.MessageOrderConfirmed, lights chan utils.MessageLights, lightsConfirmed chan utils.MessageLightsConfirmed) {

	for {
		select {

		case newMsg := <-msg:

			switch m := newMsg.(type) {
			case utils.MessageGlobalOrderArrays:
				m.Type = "MessageGlobalOrderArrays"
				arrays <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageOrderComplete:
				m.Type = "MessageOrderComplete"
				orderComplete <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageNewOrder:
				m.Type = "MessageNewOrder"
				newOrder <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageElevatorStatus:
				m.Type = "MessageElevatorStatus"
				elevatorStatus <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageOrderWatcher:
				m.Type = "MessageOrderWatcher"
				orderWatcher <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageOrderConfirmed:
				m.Type = "MessageOrderConfirmed"
				orderConfirmed <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageLights:
				m.Type = "MessageLights"
				lights <- m
				fmt.Println("Sent a", m.Type, "message")

			case utils.MessageLightsConfirmed:
				m.Type = "MessageLightsConfirmed"
				lightsConfirmed <- m
				fmt.Println("Sent a", m.Type, "message")

			}
		}
	}
}
