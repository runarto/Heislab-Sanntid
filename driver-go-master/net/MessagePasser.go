package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

var ackReceived chan bool

func InitAckMatrix() {

	for i := 0; i < utils.NumOfElevators; i++ {
		utils.AckMatrix[i] = [utils.NumButtons][utils.NumFloors]utils.Ack{}
	}

}

func UpdateAckMatrix(msgType string, id int, order utils.Order, isActive bool, isConfirmed bool, isComplete bool) {

	utils.AckMutex.Lock()
	switch msgType {
	case "MessageNewOrder":
		entry := utils.AckMatrix[id]
		entry[order.Button][order.Floor] = utils.Ack{
			Active:    isActive,
			Completed: isComplete,
			Confirmed: isConfirmed,
			Time:      time.Now(),
		}

		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()
	case "MessageOrderComplete":
		entry := utils.AckMatrix[id]
		entry[order.Button][order.Floor] = utils.Ack{
			Active:    isActive,
			Completed: isComplete,
			Confirmed: isConfirmed,
			Time:      time.Now(),
		}
		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()

	case "MessageConfirmed":
		entry := utils.AckMatrix[id]
		if utils.Master {
			entry[order.Button][order.Floor] = utils.Ack{
				Active:    isActive,
				Completed: isComplete,
				Confirmed: isConfirmed,
				Time:      time.Now()}
			fmt.Println("Master: message confirmed")
		} else {
			entry[order.Button][order.Floor] = utils.Ack{
				Active:    isActive,
				Completed: isComplete,
				Confirmed: isConfirmed,
				Time:      time.Now()}
			fmt.Println("Slave message confirmed")
		}
		utils.AckMatrix[id] = entry
		utils.AckMutex.Unlock()
	}
}

func AckReceiver(AckRx chan utils.MessageConfirmed) {

	for {
		time.Sleep(5 * time.Millisecond)
		newMsg := <-AckRx
		switch newMsg.Msg {
		case "MessageNewOrder":
			UpdateAckMatrix("MessageConfirmed", newMsg.FromElevatorID, newMsg.Order, true, true, false)
		case "MessageOrderComplete":
			UpdateAckMatrix("MessageConfirmed", newMsg.FromElevatorID, newMsg.Order, false, true, true)
		case "MessageOrders":
			if newMsg.FromElevatorID == utils.NextMasterID {
				ackReceived <- true
			}

		}
	}
}

func SendMessage(msg interface{}, NewOrderTx chan utils.MessageNewOrder,
	OrderCompleteTx chan utils.MessageOrderComplete, DoOrderCh chan utils.Order) {

	var ToElevatorID int
	var Order utils.Order
	var msgType string

	switch m := msg.(type) {
	case utils.MessageOrderComplete:
		ToElevatorID = m.ToElevatorID
		Order = m.Order
		msgType = "MessageOrderComplete"
		UpdateAckMatrix(m.Type, ToElevatorID, Order, false, false, true)
		OrderCompleteTx <- msg.(utils.MessageOrderComplete)
		fmt.Println("sendmessage: ordercomplete")
	case utils.MessageNewOrder:
		ToElevatorID = m.ToElevatorID
		Order = m.NewOrder
		msgType = "MessageNewOrder"
		UpdateAckMatrix(m.Type, ToElevatorID, Order, true, false, false)
		NewOrderTx <- msg.(utils.MessageNewOrder)
	}

	resendTimeout := 150 * time.Millisecond
	resendTimer := time.NewTimer(resendTimeout)
	timeout := time.NewTicker(1500 * time.Millisecond)

	for {

		select {

		case <-resendTimer.C:

			if CheckIfConfirmed(Order, ToElevatorID) { // If the order is confirmed, we stop resending the message.'
				fmt.Println("Order was complete")
				UpdateAckMatrix("MessageConfirmed", ToElevatorID, Order, false, false, false)
				return
			}
			switch msgType {
			case "MessageNewOrder":
				NewOrderTx <- msg.(utils.MessageNewOrder)
			case "MessageOrderComplete":
				OrderCompleteTx <- msg.(utils.MessageOrderComplete)
			}
			resendTimer.Reset(resendTimeout)
			fmt.Println("Resent message")

		case <-timeout.C:
			if CheckIfConfirmed(Order, ToElevatorID) { // If the order is confirmed, we stop resending the message.'
				fmt.Println("Order was complete")
				UpdateAckMatrix("MessageConfirmed", ToElevatorID, Order, false, false, false)
				return
			}
			if msgType == "MessageNewOrder" {
				UpdateAckMatrix(msgType, ToElevatorID, Order, false, false, false)
				DoOrderCh <- Order // If no confirmation is received, the elevator will do the order itself.
			}
			return

		}
	}
}

func CheckIfConfirmed(order utils.Order, toElevatorID int) bool {

	utils.AckMutex.Lock()
	value := utils.AckMatrix[toElevatorID][order.Button][order.Floor].Confirmed
	defer utils.AckMutex.Unlock()
	return value
}

func MessagePasser(messageSender chan interface{}, OrderCompleteTx chan utils.MessageOrderComplete,
	NewOrderTx chan utils.MessageNewOrder, ElevatorStatusTx chan utils.MessageElevatorStatus,
	MasterOrderWatcherTx chan utils.MessageOrderWatcher, LightsTx chan utils.MessageLights, OrderWatcher chan utils.OrderWatcher,
	AckRx chan utils.MessageConfirmed, DoOrderCh chan utils.Order, OrdersTx chan utils.MessageOrders) {

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
			if len(activeElevators) == 1 {
				continue
			}
			m.Type = "MessageNewOrder"
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
	timeout := 150 * time.Millisecond
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

func BroadcastMaster(MasterTx chan int) {

	var masterID int

	for {

		time.Sleep(150 * time.Millisecond)

		masterID = utils.NextMasterID

		MasterTx <- masterID

	}
}

func WaitForAck(OrdersTx chan utils.MessageOrders, Orders utils.MessageOrders, ackReceived chan bool) {

	resendTimeout := 150 * time.Millisecond
	resendTimer := time.NewTimer(resendTimeout)

	for {

		select {

		case status := <-ackReceived:
			if status {
				return
			}

		case <-resendTimer.C:
			OrdersTx <- Orders

		}

	}

}
