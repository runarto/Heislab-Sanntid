package net

import (
	"time"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessageReceiver(OrderCompleteRx chan utils.MessageOrderComplete, ElevatorStatusRx chan utils.MessageElevatorStatus,
	sender chan interface{}, distribute chan interface{}) {

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-OrderCompleteRx:
			SendAck(newMsg.Type, sender)
			distribute <- newMsg

		case newMsg := <-ElevatorStatusRx:
			if newMsg.FromElevator.ID != utils.ID {
				distribute <- newMsg
			}
		}
	}

}

func MessageDistributor(distribute chan interface{}, OrderComplete chan utils.MessageOrderComplete,
	ElevatorStatus chan utils.MessageElevatorStatus) {

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-distribute:
			switch m := newMsg.(type) {
			case utils.MessageOrderComplete:
				OrderComplete <- m
			case utils.MessageElevatorStatus:
				ElevatorStatus <- m
			}
		}
	}
}

func SendAck(typeOfMessage string, ch chan interface{}) {
	msg := utils.PackMessage("MessageConfirmed", typeOfMessage, true, utils.ID)
	ch <- msg
}

func LightsReceiver(LightsRx chan utils.MessageLights, SendLights chan [2][utils.NumFloors]bool) {
	for {
		time.Sleep(5 * time.Millisecond)
		newLights := <-LightsRx
		SendLights <- newLights.Lights
	}
}

func NewOrderReceiver(NewOrderRx chan utils.MessageNewOrder, NewOrder chan utils.MessageNewOrder, messageSender chan interface{}) {
	for {
		time.Sleep(5 * time.Millisecond)
		SendAck("MessageNewOrder", messageSender)
		NewOrder <- <-NewOrderRx
	}
}

func OrderWatcherReceiver(MasterOrderWatcherRx chan utils.MessageOrderWatcher, OrderWatcher chan utils.MessageOrderWatcher) {

	for {
		time.Sleep(5 * time.Millisecond)
		if !utils.Master {
			OrderWatcher <- <-MasterOrderWatcherRx
		}
	}
}
