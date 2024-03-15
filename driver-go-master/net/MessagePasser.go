package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
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

func SendMessage(msg interface{}, AckRx chan utils.MessageConfirmed, NewOrderTx chan utils.MessageNewOrder,
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

		case ack := <-AckRx:

			if ack.Confirmed && ack.Order == Order && ack.FromElevatorID == ToElevatorID {
				UpdateAckMatrix(ack.Msg, ack.FromElevatorID, Order)
				fmt.Println("Ack received")
				return
			}

		case <-resendTimer.C:
			channel <- msg
			fmt.Println("Resent message")

		case <-timeout.C:
			if msgType == "MessageNewOrder" {
				DoOrderCh <- Order
			}
			return

		}
	}
}

func MessagePasser(messageSender chan interface{}, OrderCompleteTx chan utils.MessageOrderComplete,
	NewOrderTx chan utils.MessageNewOrder, ElevatorStatusTx chan utils.MessageElevatorStatus,
	MasterOrderWatcherTx chan utils.MessageOrderWatcher, LightsTx chan utils.MessageLights, OrderWatcher chan utils.OrderWatcher,
	AckRx chan utils.MessageConfirmed, DoOrderCh chan utils.Order) {

	var activeElevators []int

	for {
		time.Sleep(5 * time.Millisecond)

		newMsg := <-messageSender

		switch m := newMsg.(type) {

		case utils.MessageOrderComplete:
			activeElevators = updater.GetActiveElevators()
			if len(activeElevators) == 1 {
				continue
			}
			go SendMessage(m, AckRx, NewOrderTx, OrderCompleteTx, DoOrderCh)

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
			go SendMessage(order, AckRx, NewOrderTx, OrderCompleteTx, DoOrderCh)

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

func WaitForAck(msg interface{}, msgType string, activeElevators []int, Ack chan utils.MessageConfirmed, toElevatorID int, ch chan interface{}) bool {
	var quit bool

	timeout := 150 * time.Millisecond
	responseTimer := time.NewTimer(timeout)
	responses := make(map[int]bool)

	for {
		time.Sleep(1 * time.Millisecond)
		select {
		case newMsg := <-Ack:
			switch msgType {
			case "MessageOrderComplete":
				quit, _ = HandleConfirmation(newMsg, msgType, msg, toElevatorID, responses)
				return quit
			case "MessageNewOrder":
				quit, _ = HandleConfirmation(newMsg, msgType, msg, toElevatorID, responses)
				return quit
			case "MessageLights":
				quit, responses = HandleConfirmation(newMsg, msgType, msg, toElevatorID, responses)
				if quit && len(responses) == len(activeElevators) {
					return quit
				} else {
					responseTimer.Reset(timeout)
				}
			}
		case <-responseTimer.C:
			return false
		}
	}
}

func HandleConfirmation(c utils.MessageConfirmed, msgType string,
	msg interface{}, toElevatorID int, responses map[int]bool) (bool, map[int]bool) {

	switch msgType {
	case "MessageOrderComplete":
		if c.FromElevatorID == utils.MasterID && !utils.Master {
			fmt.Println("Order complete confirmed by master")
			return true, nil
		}
	case "MessageNewOrder":
		if c.FromElevatorID == utils.MasterID && !utils.Master {
			fmt.Println("Order confirmed by master")
			return true, nil
		} else if c.FromElevatorID == toElevatorID && utils.Master {
			fmt.Println("Order confirmed by", c.FromElevatorID)
			return true, nil
		}
	case "MessageLights":
		fmt.Println("Lights confirmed by", c.FromElevatorID)
		responses[c.FromElevatorID] = true
		return true, responses
	}
	return false, nil
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

func BroadcastOrders(update chan map[int][3][utils.NumFloors]bool, NewOrderTx chan utils.MessageNewOrder) {

	orders := map[int][3][utils.NumFloors]bool{}
	timeout := 250 * time.Millisecond
	resendTimer := time.NewTimer(timeout)

	for {
		time.Sleep(50 * time.Millisecond)
		select {
		case ordersUpdate := <-update:
			orders = ordersUpdate
		case <-resendTimer.C:
			if utils.Master {
				ResendOrders(orders, NewOrderTx) // Master resends orders every 3 seconds, in case of lost messages.
			}
			resendTimer.Reset(timeout)
		}
	}
}

func ResendOrders(orders map[int][3][utils.NumFloors]bool, NewOrderTx chan utils.MessageNewOrder) {
	for id := range orders {
		if id == utils.ID {
			continue
		}
		for b := 0; b < 2; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if orders[id][b][f] {
					order := utils.Order{
						Floor:  f,
						Button: elevio.ButtonType(b),
					}
					NewOrderTx <- utils.MessageNewOrder{
						Type:           "MessageNewOrder",
						NewOrder:       order,
						ToElevatorID:   id,
						FromElevatorID: utils.ID,
					}
					fmt.Println("Resent order")
					time.Sleep(250 * time.Millisecond)
				}
			}
		}
	}
}
