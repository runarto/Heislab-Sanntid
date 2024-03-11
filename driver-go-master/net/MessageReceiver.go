package net

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessageReceiver(OrderCompleteRx chan utils.MessageOrderComplete, ElevatorStatusRx chan utils.MessageElevatorStatus, NewOrderRx chan utils.MessageNewOrder,
	MasterOrderWatcherRx chan utils.MessageOrderWatcher, LightsRx chan utils.MessageLights, sender chan interface{}, distribute chan interface{}) {

	for {
		select {

		case newMsg := <-OrderCompleteRx:
			if newMsg.FromElevatorID != utils.ID {
				fmt.Println("Received a MessageOrderComplete message")
				SendAck(newMsg.Type, sender)
				distribute <- newMsg
			}

		case newMsg := <-ElevatorStatusRx:
			if newMsg.FromElevator.ID != utils.ID {
				distribute <- newMsg
			}

		case newMsg := <-NewOrderRx:
			if newMsg.FromElevatorID != utils.ID {
				fmt.Println("Received a MessageNewOrder message")
				SendAck(newMsg.Type, sender)
				distribute <- newMsg
			}

		case newMsg := <-MasterOrderWatcherRx:
			if newMsg.FromElevatorID != utils.ID {
				distribute <- newMsg
			}

		case newMsg := <-LightsRx:
			if newMsg.FromElevatorID != utils.ID {
				fmt.Println("Received a MessageLights message")
				SendAck(newMsg.Type, sender)
				distribute <- newMsg
			}
		}
	}

}

func MessageDistributor(distribute chan interface{}, OrderComplete chan utils.MessageOrderComplete,
	ElevatorStatus chan utils.MessageElevatorStatus, NewOrder chan utils.MessageNewOrder, OrderWatcher chan utils.MessageOrderWatcher,
	Lights chan utils.MessageLights) {

	for {

		select {

		case newMsg := <-distribute:
			switch m := newMsg.(type) {
			case utils.MessageOrderComplete:
				OrderComplete <- m
			case utils.MessageElevatorStatus:
				ElevatorStatus <- m
			case utils.MessageNewOrder:
				NewOrder <- m
			case utils.MessageOrderWatcher:
				OrderWatcher <- m
			case utils.MessageLights:
				Lights <- m
			}
		}
	}
}

func SendAck(typeOfMessage string, ch chan interface{}) {
	msg := utils.PackMessage("MessageConfirmed", typeOfMessage, true, utils.ID)
	ch <- msg
}
