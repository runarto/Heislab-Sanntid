package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func InitAckMatrix() {

	for i := 0; i < utils.NumOfElevators; i++ {
		utils.AckMatrix[i] = [utils.NumButtons][utils.NumFloors]utils.Ack{}
	}

}

func UpdateAckMatrix(msgType string, id int, order utils.Order) {

	utils.AckMutex.Lock()
	switch msgType {
	case "MessageNewOrder":
		entry := utils.AckMatrix[id]
		entry[order.Button][order.Floor] = utils.Ack{
			Active:    true,
			Completed: false,
			Confirmed: false,
			Time:      time.Now(),
		}

		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()
	case "MessageOrderComplete":
		entry := utils.AckMatrix[id]
		entry[order.Button][order.Floor] = utils.Ack{
			Active:    false,
			Completed: true,
			Confirmed: false,
			Time:      time.Now(),
		}
		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()

	case "MessageConfirmed":
		entry := utils.AckMatrix[id]
		if utils.Master {
			entry[order.Button][order.Floor] = utils.Ack{
				Active:    true,
				Completed: false,
				Confirmed: true,
				Time:      time.Now()}
		} else {
			entry[order.Button][order.Floor] = utils.Ack{
				Active:    false,
				Completed: true,
				Confirmed: true,
				Time:      time.Now()}
		}
		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()
	}
}

func AckReceiver(AckRx chan utils.MessageConfirmed) {

	for {
		time.Sleep(5 * time.Millisecond)
		select {
		case newMsg := <-AckRx:
			UpdateAckMatrix("MessageConfirmed", newMsg.FromElevatorID, newMsg.Order)
		}
	}
}

func SendMessage(msg interface{}, NewOrderTx chan utils.MessageNewOrder,
	OrderCompleteTx chan utils.MessageOrderComplete, DoOrderCh chan utils.Order) {

	var ToElevatorID int
	var Order utils.Order
	var channel chan interface{}
	var msgType string

	switch m := msg.(type) {
	case utils.MessageOrderComplete:
		ToElevatorID = m.ToElevatorID
		Order = m.Order
		msgType = "MessageOrderComplete"
		UpdateAckMatrix(m.Type, ToElevatorID, Order)
		OrderCompleteTx <- msg.(utils.MessageOrderComplete)

	case utils.MessageNewOrder:
		ToElevatorID = m.ToElevatorID
		Order = m.NewOrder
		msgType = "MessageNewOrder"
		UpdateAckMatrix(m.Type, ToElevatorID, Order)
		NewOrderTx <- msg.(utils.MessageNewOrder)
	}

	resendTimeout := 150 * time.Millisecond
	resendTimer := time.NewTimer(resendTimeout)
	timeout := time.NewTicker(500 * time.Millisecond)

	channel <- msg

	for {

		select {

		case <-resendTimer.C:

			if CheckIfConfirmed(Order, ToElevatorID) { // If the order is confirmed, we stop resending the message.
				return
			}

			channel <- msg
			fmt.Println("Resent message")

		case <-timeout.C:
			if msgType == "MessageNewOrder" {
				UpdateAckMatrix(msgType, ToElevatorID, Order)
				DoOrderCh <- Order // If no confirmation is received, the elevator will do the order itself.
			}
			return

		}
	}
}

func CheckIfConfirmed(order utils.Order, toElevatorID int) bool {

	utils.AckMutex.Lock()
	defer utils.AckMutex.Unlock()
	return utils.AckMatrix[toElevatorID][order.Button][order.Floor].Confirmed
}

func MessagePasser(messageSender chan interface{}, OrderCompleteTx chan utils.MessageOrderComplete,
	NewOrderTx chan utils.MessageNewOrder, ElevatorStatusTx chan utils.MessageElevatorStatus,
	MasterOrderWatcherTx chan utils.MessageOrderWatcher, LightsTx chan utils.MessageLights, OrderWatcher chan utils.OrderWatcher,
	AckRx chan utils.MessageConfirmed, DoOrderCh chan utils.Order) {

	var activeElevators []int

	go AckReceiver(AckRx)

	for {
		time.Sleep(5 * time.Millisecond)

		newMsg := <-messageSender

		switch m := newMsg.(type) {

		case utils.MessageOrderComplete:
			activeElevators = updater.GetActiveElevators()
			if len(activeElevators) == 1 {
				continue
			}
			go SendMessage(m, NewOrderTx, OrderCompleteTx, DoOrderCh)

		case utils.MessageNewOrder:
			activeElevators = updater.GetActiveElevators()
			fmt.Println("Active elevators: ", activeElevators)
			if len(activeElevators) == 1 {
				continue
			}
			m.Type = "MessageNewOrder"
			NewOrderTx <- newMsg.(utils.MessageNewOrder)
			fmt.Println("Send a new order message", newMsg.(utils.MessageNewOrder).NewOrder)
			order := newMsg.(utils.MessageNewOrder)
			go SendMessage(order, NewOrderTx, OrderCompleteTx, DoOrderCh)

		case utils.MessageElevatorStatus:
			m.Type = "MessageElevatorStatus"
			ElevatorStatusTx <- newMsg.(utils.MessageElevatorStatus)

		case utils.MessageOrderWatcher:
			activeElevators = updater.GetActiveElevators()
			if len(activeElevators) == 1 {
				continue
			}
			m.Type = "MessageOrderWatcher"
			MasterOrderWatcherTx <- newMsg.(utils.MessageOrderWatcher)
		}
	}
}

func BroadcastLights(SendLights chan [2][utils.NumFloors]bool, LightsTx chan utils.MessageLights) {

	lightsForSending := [2][utils.NumFloors]bool{}
	timeout := 200 * time.Millisecond
	resendTimer := time.NewTimer(timeout)

	for {
		time.Sleep(15 * time.Millisecond)
		select {
		case lights := <-SendLights:
			lightsForSending = lights
			if utils.Master {
				LightsTx <- utils.MessageLights{
					Type:           "MessageLights",
					Lights:         lightsForSending,
					FromElevatorID: utils.ID}
				resendTimer.Reset(timeout)
			}
		case <-resendTimer.C:
			if utils.Master {
				LightsTx <- utils.MessageLights{
					Type:           "MessageLights",
					Lights:         lightsForSending,
					FromElevatorID: utils.ID}
			}
			resendTimer.Reset(timeout)
		}
	}
}
