package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessageReceiver(NewOrderRx chan utils.MessageNewOrder, OrderCompleteRx chan utils.MessageOrderComplete, ElevatorStatusRx chan utils.MessageElevatorStatus,
	AckTx chan utils.MessageConfirmed, distribute chan interface{}, LightsRx chan utils.MessageLights, OrdersRx chan utils.MessageOrders, continueChannel chan bool) {

	OrdersActive := [2][utils.NumFloors]bool{}
	var proceed bool

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-OrderCompleteRx:
			if newMsg.FromElevatorID != utils.ID {
				SendAck(newMsg.Type, AckTx, newMsg.Order)
				f := newMsg.Order.Floor
				b := newMsg.Order.Button
				OrdersActive, proceed = UpdateActiveOrders(OrdersActive, int(b), f, false, newMsg.ToElevatorID)
				if !proceed {
					continue
				}
				fmt.Println("Received order complete from elevator", newMsg.FromElevatorID)
				distribute <- newMsg
			}

		case newMsg := <-ElevatorStatusRx:
			if newMsg.FromElevator.ID != utils.ID {
				distribute <- newMsg
			}
		case newMsg := <-NewOrderRx:
			if newMsg.FromElevatorID != utils.ID {
				SendAck(newMsg.Type, AckTx, newMsg.NewOrder)
				f := newMsg.NewOrder.Floor
				b := newMsg.NewOrder.Button
				OrdersActive, proceed = UpdateActiveOrders(OrdersActive, int(b), f, true, newMsg.ToElevatorID)
				if !proceed {
					continue
				}
				fmt.Println("Received new order from elevator", newMsg.FromElevatorID)
				fmt.Println("Order received is: ", newMsg.NewOrder)
				distribute <- newMsg

			}
		}
	}
}

func UpdateActiveOrders(Orders [2][utils.NumFloors]bool, b int, f int, isNew bool, toElevatorID int) ([2][utils.NumFloors]bool, bool) {

	if toElevatorID != utils.ID {
		return Orders, false
	}

	if b == utils.Cab {
		return Orders, true
	}

	if Orders[b][f] && !isNew {
		Orders[b][f] = isNew
		return Orders, true
	}

	if Orders[b][f] && isNew {
		return Orders, false
	}

	if !Orders[b][f] && isNew {
		Orders[b][f] = true
		return Orders, true
	}

	if !Orders[b][f] && !isNew {
		return Orders, true
	}

	return Orders, false

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

func MasterBroadcastReceiver(MasterRx chan int, MasterUpdateCh chan int) {

	for {
		time.Sleep(10 * time.Millisecond)
		masterID := <-MasterRx
		if masterID != utils.MasterID {
			MasterUpdateCh <- masterID
		}

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
