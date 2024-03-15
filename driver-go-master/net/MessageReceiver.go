package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessageReceiver(NewOrderRx chan utils.MessageNewOrder, OrderCompleteRx chan utils.MessageOrderComplete, ElevatorStatusRx chan utils.MessageElevatorStatus,
	AckTx chan utils.MessageConfirmed, distribute chan interface{}, LightsRx chan utils.MessageLights) {

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-OrderCompleteRx:
			if newMsg.FromElevatorID != utils.ID {
				fmt.Println("Received order complete from elevator", newMsg.FromElevatorID)
				SendAck(newMsg.Type, AckTx, newMsg.Order)
				distribute <- newMsg
			}

		case newMsg := <-ElevatorStatusRx:
			if newMsg.FromElevator.ID != utils.ID {
				fmt.Println("Received elevator status from elevator", newMsg.FromElevator.ID)
				distribute <- newMsg
			}
		case newMsg := <-NewOrderRx:
			if newMsg.FromElevatorID != utils.ID {
				fmt.Println("Received new order from elevator", newMsg.FromElevatorID)
				fmt.Println("Order received is: ", newMsg.NewOrder)
				SendAck(newMsg.Type, AckTx, newMsg.NewOrder)
				distribute <- newMsg
			}
		}
	}
}

func MessageDistributor(distribute chan interface{}, OrderComplete chan utils.MessageOrderComplete,
	ElevatorStatus chan utils.MessageElevatorStatus, NewOrder chan utils.MessageNewOrder, SendLights chan [2][utils.NumFloors]bool) {

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-distribute:
			switch m := newMsg.(type) {
			case utils.MessageOrderComplete:
				OrderComplete <- m
			case utils.MessageElevatorStatus:
				ElevatorStatus <- m
			case utils.MessageNewOrder:
				NewOrder <- m
			case utils.MessageLights:
				SendLights <- m.Lights
			}
		}
	}
}

func SendAck(typeOfMessage string, AckTx chan utils.MessageConfirmed, order utils.Order) {

	AckTx <- utils.MessageConfirmed{
		Type:           "MessageConfirmed",
		Msg:            typeOfMessage,
		Order:          order,
		Confirmed:      true,
		FromElevatorID: utils.ID}

	fmt.Println("Sent ack for", typeOfMessage)
}

func LightsReceiver(LightsRx chan utils.MessageLights, SendLights chan [2][utils.NumFloors]bool) {
	for {
		time.Sleep(50 * time.Millisecond)
		newLights := <-LightsRx
		SendLights <- newLights.Lights
	}
}

func OrderWatcherReceiver(MasterOrderWatcherRx chan utils.MessageOrderWatcher, OrderWatcher chan utils.MessageOrderWatcher) {

	for {
		time.Sleep(50 * time.Millisecond)
		if !utils.Master {
			OrderWatcher <- <-MasterOrderWatcherRx
		}
	}
}
