package net

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func MessagePasser(messageSender <-chan interface{}, OrderCompleteTx chan utils.MessageOrderComplete,
	NewOrderTx chan utils.MessageNewOrder, ElevatorStatusTx chan utils.MessageElevatorStatus,
	MasterOrderWatcherTx chan utils.MessageOrderWatcher, LightsTx chan utils.MessageLights, OrderWatcher chan utils.OrderWatcher,
	AckTx chan utils.MessageConfirmed, AckRx chan utils.MessageConfirmed) {

	activeElevators := updater.GetActiveElevators()

	for {
		time.Sleep(5 * time.Millisecond)
		select {

		case newMsg := <-messageSender:

			switch m := newMsg.(type) {

			case utils.MessageOrderComplete:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageOrderComplete"
				OrderCompleteTx <- newMsg.(utils.MessageOrderComplete)
				go WaitForAck(newMsg, m.Type, activeElevators, AckRx, utils.NotDefined, OrderCompleteTx)

			case utils.MessageNewOrder:
				if len(activeElevators) == 1 {
					continue
				}
				m.Type = "MessageNewOrder"
				NewOrderTx <- newMsg.(utils.MessageNewOrder)
				fmt.Println("Send a new order message", newMsg.(utils.MessageNewOrder).NewOrder)
				orderMessage := newMsg.(utils.MessageNewOrder)
				go WaitForAck(newMsg, m.Type, activeElevators, AckRx, orderMessage.ToElevatorID, NewOrderTx)

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
				m.Type = "MessageLights"
				LightsTx <- newMsg.(utils.MessageLights)

			case utils.MessageConfirmed:
				AckTx <- newMsg.(utils.MessageConfirmed)

			}

			activeElevators = updater.GetActiveElevators()

		}

	}
}

func WaitForAck(msg interface{}, msgType string, activeElevators []int, Ack chan utils.MessageConfirmed, toElevatorID int, channel ...interface{}) {
	var quit bool

	timeout := 100 * time.Millisecond
	responseTimer := time.NewTimer(timeout)

	fmt.Println("Master is ", utils.MasterID)
	fmt.Println("My id is: ", utils.ID)
	if toElevatorID != utils.ID {
		fmt.Println("message not meant for me")
		return
	}

	for {
		time.Sleep(5 * time.Millisecond)
		select {
		case newMsg := <-Ack:
			quit = HandleConfirmation(newMsg, msgType, msg, toElevatorID)
			if quit {
				return
			}
		case <-responseTimer.C:
			ResendMessage(msg, msgType, channel[0])
		}
	}
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
	case utils.MessageNewOrder:
		m.Type = "MessageNewOrder"
		ch, ok := channel[0].(chan utils.MessageNewOrder)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
	case utils.MessageElevatorStatus:
		m.Type = "MessageElevatorStatus"
		ch, ok := channel[0].(chan utils.MessageElevatorStatus)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
	case utils.MessageOrderWatcher:
		m.Type = "MessageOrderWatcher"
		ch, ok := channel[0].(chan utils.MessageOrderWatcher)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
	case utils.MessageLights:
		m.Type = "MessageLights"
		ch, ok := channel[0].(chan utils.MessageLights)
		if !ok {
			fmt.Println("Invalid channel type")
			return
		}
		ch <- m
	}
}

func HandleConfirmation(c utils.MessageConfirmed, msgType string, msg interface{}, toElevatorID int) bool {

	switch msgType {
	case "MessageOrderComplete":
		if c.FromElevatorID == utils.MasterID && !utils.Master {
			fmt.Println("Order complete confirmed by master")
			return true
		}
	case "MessageNewOrder":
		if c.FromElevatorID == utils.MasterID && !utils.Master {
			fmt.Println("Order confirmed by master")
			return true
		} else if c.FromElevatorID == toElevatorID && utils.Master {
			return true
		}
	}
	return false
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
	timeout := 100 * time.Millisecond
	resendTimer := time.NewTimer(timeout)

	for {
		time.Sleep(15 * time.Millisecond)
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
			time.Sleep(5 * time.Millisecond)
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
				}
			}
		}
	}
}
