package net

import (
	"fmt"
	"reflect"
	"time"

	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Packet struct}
//*
//* @param      Msg          The message
//* @param      HashNumber   The hash number
//* @param      ToElevatorID The to elevator identifier
//*

type Packet struct {
	Msg          utils.Message `json:"msg"`
	HashNumber   int           `json:"hashNumber"`
	ToElevatorID int           `json:"toElevatorID"`
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

var maxValue = 123456
var sendLights = make(chan [2][utils.NumFloors]bool)
var updateLights = make(chan [2][utils.NumFloors]bool)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

func NetworkHandler(messageHandler chan utils.Message, messageReceiver chan Packet, messageSender chan Packet, messageDistributor chan utils.Message,
	OrderHandlerNetworkUpdateCh chan utils.Message, SlaveUpdateCh chan utils.Order) {

	hashValue := 0
	confirmationMap := make(map[int]Packet)
	retriesMap := make(map[int]int)
	resendTimer := time.NewTicker(50 * time.Millisecond)
	var activeOrders [2][utils.NumFloors]bool
	const retries = 16

	for {

		select {

		case setLights := <-updateLights:
			activeOrders = setLights
			sendLights <- activeOrders

		case msg := <-messageHandler: // Messages for sending

			updateActiveOrders(msg, &activeOrders, sendLights)
			messageSender <- toPacket(msg, getHashValue(hashValue))
			hashValue = getHashValue(hashValue)
			confirmationMap[hashValue] = toPacket(msg, hashValue)

		case packet := <-messageReceiver: // Messages for receiving

			if forMe(packet) {

				switch packet.Msg.Type {
				case "MessageConfirmed":

					SendToSlaveWatcher(confirmationMap, packet.HashNumber, SlaveUpdateCh)
					delete(confirmationMap, packet.HashNumber)
					delete(retriesMap, packet.HashNumber)
					fmt.Println("Message confirmed: ", packet.HashNumber)
				case "MessageOrderComplete", "MessageNewOrder":
					confirm := confirmPacket(packet)
					messageSender <- toPacket(confirm.(utils.Message), packet.HashNumber)

					updateActiveOrders(packet.Msg, &activeOrders, sendLights)
					newMsg := utils.DecodeMessage(packet.Msg, packet.Msg.Type)
					OrderHandlerNetworkUpdateCh <- newMsg
				}
			}

		case <-resendTimer.C:

			ResendPacks(&confirmationMap, &retriesMap, messageSender, retries)

		}

	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Sends the packet to the slave watcher}
//*
//* @param      confirmationMap  Map that indicates if a packet has been confirmed
//* @param      hashValue        The hash value
//* @param      SlaveUpdateCh    Channel for updating the slave watcher
//*

func SendToSlaveWatcher(confirmationMap map[int]Packet, hashValue int, SlaveUpdateCh chan utils.Order) {

	packet := confirmationMap[hashValue]

	switch packet.Msg.Type {
	case "MessageOrderComplete":
		return
	case "MessageNewOrder":
		newMsg := utils.DecodeMessage(packet.Msg, packet.Msg.Type)
		SlaveUpdateCh <- newMsg.Msg.(utils.MessageNewOrder).NewOrder
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Updates the active orders}
//*
//* @param      msg         The message to update from
//* @param      activeOrders The active orders
//* @param      sendLights  Channel for sending the lights
//*

func updateActiveOrders(msg utils.Message, activeOrders *[2][utils.NumFloors]bool, sendLights chan [2][utils.NumFloors]bool) {

	if !utils.Master {
		return
	}

	fmt.Println("here")

	newMsg := utils.DecodeMessage(msg, msg.Type)

	fmt.Println("newmsg", newMsg.Type)

	switch newMsg.Type {
	case "MessageOrderComplete":
		f := newMsg.Msg.(utils.MessageOrderComplete).Order.Floor
		b := newMsg.Msg.(utils.MessageOrderComplete).Order.Button
		if b != utils.Cab {
			activeOrders[b][f] = false
		}
		fmt.Println("Order complete: ", activeOrders)
		sendLights <- *activeOrders
	case "MessageNewOrder":
		f := newMsg.Msg.(utils.MessageNewOrder).NewOrder.Floor
		fmt.Println("new order floor: ", f)
		b := newMsg.Msg.(utils.MessageNewOrder).NewOrder.Button
		fmt.Println("new order button: ", b)
		if b != utils.Cab {
			activeOrders[b][f] = true
		}
		fmt.Println("New order: ", activeOrders)
		sendLights <- *activeOrders
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Converts a message to a packet}
//*
//* @param      msg       The message
//* @param      hashValue The hash value
//*
//* @return     {The packet}

// Converts a message to a packet.
func toPacket(msg utils.Message, hashValue int) Packet {
	return Packet{Msg: msg, HashNumber: hashValue, ToElevatorID: msg.ToElevatorID}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Converts a packet to a message}
//*
//* @param      packet  The packet
//*
//* @return     {The message}

// Converts a packet to a message.
func confirmPacket(packet Packet) interface{} {
	hash := packet.HashNumber
	msg := utils.PackMessage("MessageConfirmed", packet.Msg.FromElevatorID, utils.ID, hash)
	return msg
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {gets the hash value for the packet}
//*
//* @param      hashValue  the previous hash value
//*
//* @return     {the new hash value}

func getHashValue(hashValue int) int {
	if hashValue >= maxValue {
		return 0
	} else {
		return hashValue + 1
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Resends the packets}
//*
//* @param      confirmationMap  Map that indicates if a packet has been confirmed
//* @param      retriesMap       Map that indicates the number of retries for a packet
//* @param      packetSender     Channel for sending packets
//* @param      retries          The amount of retries
//*

func ResendPacks(confirmationMap *map[int]Packet, retriesMap *map[int]int, packetSender chan Packet, retries int) {
	for _, packet := range *confirmationMap {
		packetSender <- packet
		(*retriesMap)[packet.HashNumber]++
		if (*retriesMap)[packet.HashNumber] > retries {
			delete(*confirmationMap, packet.HashNumber)
			delete(*retriesMap, packet.HashNumber)
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Checks if the packet is for the current elevator}
//*
//* @param      packet  The packet
//*
//* @return     {True if the packet is for the current elevator, False otherwise}
//*

func forMe(packet Packet) bool {
	return packet.ToElevatorID == utils.ID
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {BroadcastMaster broadcasts the master ID to the other elevators}
//*
//* @param      MasterID_Tx  Channel for sending the master ID
//*

// BroadcastMaster broadcasts the master ID to the other elevators.
func BroadcastMaster(MasterID_Tx chan int) {

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	var masterID int

	for {
		masterID = utils.NextMasterID
		select {
		case <-ticker.C:
			MasterID_Tx <- masterID
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {BroadcastLights broadcasts the lights to the other elevators (master only)}
//*
//* @param      LightsTx  Channel for sending the lights
//*

// BroadcastLights broadcasts the lights to the other elevators.
func BroadcastLights(LightsTx chan utils.MessageLights) {
	ticker := time.NewTicker(75 * time.Millisecond)
	defer ticker.Stop()

	for {
		lights := <-sendLights
		select {
		case <-ticker.C:
			if utils.Master {
				fmt.Print("Broadcasting lights: ", lights, "\n")
				LightsTx <- utils.MessageLights{Type: "Lights", Lights: lights}
			}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {BroadcastElevatorState broadcasts the elevator state to the other elevators}
//*
//* @param      LocalElevatorStateUpdateCh  Channel for updating the local elevator state
//* @param      ElevatorStateTx             Channel for sending the elevator state
//* @param      e                           The elevator
//*

// BroadcastElevatorState broadcasts the elevator state to the other elevators.
func BroadcastElevatorState(LocalElevatorStateUpdateCh chan utils.Elevator, ElevatorStateTx chan utils.MessageElevatorStatus, e *utils.Elevator) {
	ticker := time.NewTicker(75 * time.Millisecond)
	defer ticker.Stop()

	for {
		state := <-LocalElevatorStateUpdateCh
		*e = state
		select {
		case <-ticker.C:
			ElevatorStateTx <- utils.MessageElevatorStatus{Type: "ElevatorStatus", Elevator: state}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {BroadcastOrderWatcher broadcasts the active orders to the other elevators}
//*
//* @param      MasterOrderWatcherTx  Channel for sending the master order watcher
//* @param      activeOrders          Channel for receiving the active orders
//*

// BroadcastOrderWatcher broadcasts the active orders to the other elevators.
func BroadcastOrderWatcher(MasterOrderWatcherTx chan utils.MessageOrderWatcher, activeOrders chan map[int][utils.NumButtons][utils.NumFloors]bool) {
	ticker := time.NewTicker(75 * time.Millisecond)
	defer ticker.Stop()

	for {
		Orders := <-activeOrders
		select {
		case <-ticker.C:
			if utils.Master {
				OrdersStrings := utils.Map_IntToString(Orders)
				MasterOrderWatcherTx <- utils.MessageOrderWatcher{Type: "OrderWatcher", Orders: OrdersStrings}
			}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

var prevLights = [2][utils.NumFloors]bool{}
var elevators = make(map[int]utils.Elevator)
var prevOrders = make(map[int][utils.NumButtons][utils.NumFloors]bool)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Receives broadcasts from the other elevators}
//*
//* @param      MasterOrderWatcherRx  Channel for receiving the master order watcher
//* @param      ElevatorStateRx       Channel for receiving the elevator state
//* @param      LightsRx              Channel for receiving the lights
//* @param      ArrayUpdater          Channel for updating elevator states and master order watcher
//* @param      SetLights             Channel for setting the lights
//* @param      MasterID_Rx           Channel for receiving the master ID
//* @param      MasterUpdateCh        Channel for updating the master
//*

func ReceiveBroadcasts(MasterOrderWatcherRx chan utils.MessageOrderWatcher, ElevatorStateRx chan utils.MessageElevatorStatus, LightsRx chan utils.MessageLights,
	ArrayUpdater chan utils.Message, SetLights chan [2][utils.NumFloors]bool, MasterID_Rx chan int, MasterUpdateCh chan int) {

	for {
		select {
		case msg := <-MasterOrderWatcherRx:
			Orders := utils.Map_StringToInt(msg.Orders)
			if !reflect.DeepEqual(Orders, prevOrders) && !utils.Master {
				ArrayUpdater <- utils.Message{Type: "OrderWatcher", ToElevatorID: utils.NotDefined, FromElevatorID: utils.NotDefined, Msg: msg}
				prevOrders = Orders
			}
			utils.Orders = Orders
		case state := <-ElevatorStateRx:
			if !reflect.DeepEqual(state.Elevator, elevators[state.Elevator.ID]) {
				ArrayUpdater <- utils.Message{Type: "ElevatorStatus", ToElevatorID: utils.NotDefined, FromElevatorID: utils.NotDefined, Msg: state}
				elevators[state.Elevator.ID] = state.Elevator
			}
		case lights := <-LightsRx:
			if !reflect.DeepEqual(lights.Lights, prevLights) {
				SetLights <- lights.Lights
				prevLights = lights.Lights
			}
		case masterUpdate := <-MasterID_Rx:
			if masterUpdate != utils.MasterID {
				HandleMasterUpdate(masterUpdate)
			}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Handles the master update}
//*
//* @param      val   The ID of the new master
//*

// When the master changes,
func HandleMasterUpdate(val int) {
	utils.MasterMutex.Lock()
	utils.MasterIDmutex.Lock()
	fmt.Println("Master update: ", val)
	if val == utils.ID {
		utils.MasterID = val
		utils.Master = true
		fmt.Println("I am master")
	} else {
		utils.MasterID = val
		utils.Master = false
		fmt.Println("The master is elevator ", val)
	}
	utils.MasterIDmutex.Unlock()
	utils.MasterMutex.Unlock()

	newLights := GetLights(prevOrders) // Get the lights from the orders
	updateLights <- newLights          // Set the lights

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Gets the lights to send to the other elevators}
//*
//* @param      orders  The orders
//*
//* @return     {The lights}

func GetLights(orders map[int][utils.NumButtons][utils.NumFloors]bool) [2][utils.NumFloors]bool {
	var lights [2][utils.NumFloors]bool
	for id := range orders {
		for b := 0; b < utils.NumButtons; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if orders[id][b][f] {
					lights[b][f] = true
				}
			}
		}
	}
	return lights
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
