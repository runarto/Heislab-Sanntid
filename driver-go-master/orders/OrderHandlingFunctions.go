package orders

import (
	"fmt"
	"strconv"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/updater"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Processes a new order based on the given parameters}
//*
//* @param      order          The order
//* @param      e              The elevator
//* @param      messageHandler Channel for sending orders to the network
//* @param      toElevatorID   Identifier for the elevator to send the order to
// */

func SendOrder(order utils.Order, e utils.Elevator, messageHandler chan utils.Message, toElevatorID int) {

	msg := utils.PackMessage("MessageNewOrder", toElevatorID, utils.ID, order)
	messageHandler <- msg

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Processes a new order based on the given parameters}
//*
//* @param      order          The order
//* @param      e              The elevator
//* @param      messageHandler Channel for sending orders to the network
//* @param      AllOrdersCh    The all orders channel
//* @param      DoOrderCh      The do order channel
//* @param      isOnline       Indicates if online
//* @param      fromElevatorID Identifier for the elevator that sent the order
// */

func ProcessNewOrder(order utils.Order, e utils.Elevator, messageHandler chan utils.Message, AllOrdersCh chan utils.GlobalOrderUpdate,
	DoOrderCh chan utils.Order, isOnline bool, fromElevatorID int) {

	fmt.Println("Function: ProcessNewOrder")

	switch utils.Master {
	case true:

		fmt.Println(utils.Orders)
		if !CheckIfOrderIsAlreadyActive(order) {
			return
		}

		// CAB ORDER----------------------------------------------

		if order.Button == utils.Cab {

			AllOrdersCh <- utils.GlobalOrderUpdate{
				Order:         order,
				ForElevatorID: fromElevatorID,
				IsComplete:    false,
				IsNew:         true}

			SendOrder(order, e, messageHandler, fromElevatorID)
			return
		}

		// HALL ORDER---------------------------------------------

		BestElevator := utils.ChooseElevator(order)

		AllOrdersCh <- utils.GlobalOrderUpdate{
			Order:         order,
			ForElevatorID: BestElevator.ID,
			IsComplete:    false,
			IsNew:         true}

		fmt.Println("Best elevator for order", order, ": ", BestElevator.ID)

		fmt.Println("Sending message for global order update...")

		SendOrder(order, e, messageHandler, BestElevator.ID)

	case false:

		if isOnline {
			SendOrder(order, e, messageHandler, utils.MasterID) // Send to master if online
			return
		} else {
			DoOrderCh <- order // If offline, send to FSM
			return
		}
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Processes the order complete based on the given parameters}
//*
//* @param      orderComplete  The order that is complete
//* @param      e              The elevator
//* @param      AllOrdersCh    Channel for updating the global order array
//* @param      FromElevatorID Identifier for the elevator that sent the order
// */

func ProcessOrderComplete(orderComplete utils.Order, e utils.Elevator, AllOrdersCh chan utils.GlobalOrderUpdate, FromElevatorID int) {

	GlobalOrderUpdate := utils.GlobalOrderUpdate{
		Order:         orderComplete,
		ForElevatorID: FromElevatorID,
		IsComplete:    true,
		IsNew:         false}

	AllOrdersCh <- GlobalOrderUpdate

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Handles the peers update based on the given parameters}
//*
//* @param      p                   Contains active peers, new peers and lost peers
//* @param      IsOnlineCh          Channel for updating the online status
//* @param      ActiveElevatorUpdate Channel for updating the active elevators
//* @param      Online              The online status
// */

func HandlePeersUpdate(p peers.PeerUpdate, IsOnlineCh chan bool, ActiveElevatorUpdateCh chan utils.Status, Online *bool) {

	fmt.Println("Function: HandlePeersUpdate")

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)

	if len(p.Peers) == 0 || len(p.Peers) == 1 {

		fmt.Println("No other peers available, elevator is disconnected")

		IsOnlineCh <- false
		*Online = false

		utils.NextMasterID = utils.NotDefined

	} else {

		FindNewMaster(p.Peers)

		IsOnlineCh <- true
		*Online = true

	}

	if p.New != "" || p.Lost != nil {

		HandleActiveElevators(p.New, p.Lost, ActiveElevatorUpdateCh)

	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Finds the new master based on the given parameters}
// *
// * @param      peers The active elevators
// */

func FindNewMaster(peers []string) {

	var nextMaster = utils.ID

	for i := range peers {
		peerID, _ := strconv.Atoi(peers[i])
		if peerID < nextMaster {
			nextMaster = peerID
		}
	}
	utils.NextMasterID = nextMaster
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Handles the active elevators based on the given parameters}
// *
// * @param      new                    The new elevator
// * @param      lost                   The lost elevators
// * @param      ActiveElevatorUpdateCh Channel for updating the active elevators
// */

func HandleActiveElevators(new string, lost []string, ActiveElevatorUpdateCh chan utils.Status) {

	if new != "" {
		newElevatorID, _ := strconv.Atoi(new)
		ActiveElevatorUpdateCh <- utils.Status{ID: newElevatorID, IsOnline: true}
	}

	for i := range lost {
		lostElevatorID, _ := strconv.Atoi(lost[i])
		ActiveElevatorUpdateCh <- utils.Status{ID: lostElevatorID, IsOnline: false}
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Handles the new order based on the given parameters}
//*
//* @param      newOrder        The new order
//* @param      e               The elevator
//* @param      messageHandler  Channel for sending orders to the network
//* @param      AllOrdersCh     Channel for updating the global order array
//* @param      DoOrderCh       Channel for sending orders to the FSM
//* @param      isOnline        Indicates if online
//* @param      FromElevatorID  Identifier for the elevator that sent the order
//* @param      ToElevatorID    Identifier for the elevator to send the order to
// */

func HandleNewOrder(newOrder utils.MessageNewOrder, e utils.Elevator, messageHandler chan utils.Message,
	AllOrdersCh chan utils.GlobalOrderUpdate, DoOrderCh chan utils.Order, isOnline bool, FromElevatorID int, ToElevatorID int) {

	order := newOrder.NewOrder

	if utils.Master && FromElevatorID != utils.ID {

		fmt.Println("---NEW ORDER TO DELEGATE---")

		ProcessNewOrder(order, e, messageHandler, AllOrdersCh, DoOrderCh, isOnline, FromElevatorID)

	} else if (!utils.Master && ToElevatorID == utils.ID) ||
		(ToElevatorID == utils.ID && FromElevatorID == utils.ID && utils.Master) {

		fmt.Println("---NEW ORDER RECEIVED---")

		AllOrdersCh <- utils.GlobalOrderUpdate{
			Order:         order,
			ForElevatorID: utils.ID,
			IsComplete:    false,
			IsNew:         true,
		}

		DoOrderCh <- order
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Handles the button event based on the given parameters}
//*
//* @param      newOrder        The new order
//* @param      e               The elevator
//* @param      messageHandler  Channel for sending orders to the network
//* @param      AllOrdersCh     Channel for updating the global order array
//* @param      DoOrderCh       Channel for sending orders to the FSM
//* @param      isOnline        Indicates if online
// */

func HandleButtonEvent(newOrder elevio.ButtonEvent, e utils.Elevator, messageHandler chan utils.Message,
	AllOrdersCh chan utils.GlobalOrderUpdate, DoOrderCh chan utils.Order, isOnline bool) {

	fmt.Println("---LOCAL BUTTON PRESSED---")

	order := utils.Order{
		Floor:  newOrder.Floor,
		Button: newOrder.Button,
	}

	fmt.Println("Button pressed: ", order)
	fmt.Println("Local order array:")

	if !isOnline {
		DoOrderCh <- order
	} else {

		SendOrder(order, e, messageHandler, utils.MasterID)

	}

	if !utils.Master {
		updater.SlaveOrderWatcherUpdate(true, false, order, e, &updater.SlaveOrderWatcher)
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Checks if the order is already active based on the given parameters}
//*
//* @param      Order  the order in question
// */

func CheckIfOrderIsAlreadyActive(Order utils.Order) bool {
	utils.OrdersMutex.Lock()
	defer utils.OrdersMutex.Unlock()

	b := Order.Button
	f := Order.Floor

	if b == utils.Cab {
		return true
	}

	for id := range utils.Orders {
		fmt.Println("value: ", utils.Orders[id][b][f])
		if utils.Orders[id][b][f] {
			return false
		}
	}
	return true

}
