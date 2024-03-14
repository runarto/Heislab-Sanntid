package orders

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

type peerUpdate struct {
	New  string
	Lost []string
}

var ActiveElevatorIDs []int
var Elevators []utils.Elevator

func SendOrder(order utils.Order, e utils.Elevator, ch chan interface{}, toElevatorID int) {

	msg := utils.PackMessage("MessageNewOrder", order, toElevatorID, e.ID)
	ch <- msg

}

func ProcessNewOrder(order utils.Order, e utils.Elevator, ch chan interface{}, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	DoOrderCh chan utils.Order, watcher chan utils.OrderWatcher, IsOnlineCh chan bool, LocalOrders [3][utils.NumFloors]bool, isOnline bool) [3][utils.NumFloors]bool {

	fmt.Println("Function: ProcessNewOrder")

	switch utils.Master {
	case true:
		BestElevator := ChooseElevator(order)
		fmt.Println("Best elevator for order", order, ": ", BestElevator.ID)

		fmt.Println("Sending message for global order update...")

		go func() {
			GlobalUpdateCh <- utils.GlobalOrderUpdate{
				Order:          order,
				ForElevatorID:  BestElevator.ID,
				FromElevatorID: e.ID,
				IsComplete:     false,
				IsNew:          true}
		}()

		if BestElevator.ID == e.ID || !isOnline {

			DoOrderCh <- order
			fmt.Println("Sending order to FSM...")
		} else {
			SendOrder(order, e, ch, BestElevator.ID)
		}

	case false:

		if isOnline {
			SendOrder(order, e, ch, utils.MasterID)
		} else {
			DoOrderCh <- order
		}

	}

	return LocalOrders

}

func ProcessOrderComplete(orderComplete utils.MessageOrderComplete, e utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate) {

	GlobalOrderUpdate := utils.GlobalOrderUpdate{
		Order:          orderComplete.Order,
		ForElevatorID:  orderComplete.FromElevatorID,
		FromElevatorID: orderComplete.FromElevatorID,
		IsComplete:     true,
		IsNew:          false}

	GlobalUpdateCh <- GlobalOrderUpdate

}

func HandlePeersUpdate(p peers.PeerUpdate, IsOnlineCh chan bool, MasterUpdateCh chan int, ActiveElevatorUpdate chan utils.Status, Online *bool) {

	fmt.Println("Function: HandlePeersUpdate")

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)

	var ActiveElevators peerUpdate

	if len(p.Peers) == 0 {

		fmt.Println("No peers available, elevator is disconnected")

		IsOnlineCh <- false
		*Online = false

		MasterUpdateCh <- utils.ID

		return

	} else {

		val, _ := strconv.Atoi(p.Peers[0])
		utils.NextMasterID = val

		if DidMasterGoOffline(val) {
			MasterUpdateCh <- val
		}

		IsOnlineCh <- true
		*Online = true

		ActiveElevators = HandleNewPeers(p, ActiveElevators)
		ActiveElevators = HandleLostPeers(p, ActiveElevators)

		if ActiveElevators.New != "" || ActiveElevators.Lost != nil {

			HandleActiveElevators(ActiveElevators, ActiveElevatorUpdate)

		}

	}

}

func DidMasterGoOffline(val int) bool {
	if val > utils.MasterID {
		return true
	} else {
		return false
	}
}

func HandleNewPeers(p peers.PeerUpdate, peerUpdate peerUpdate) peerUpdate {

	if p.New != "" {
		newElevatorID := p.New
		peerUpdate.New = newElevatorID
	}

	return peerUpdate
}

func HandleLostPeers(p peers.PeerUpdate, peerUpdate peerUpdate) peerUpdate {

	if p.Lost != nil {
		peerUpdate.Lost = append(peerUpdate.Lost, p.Lost...)
	}

	return peerUpdate

}

func HandleActiveElevators(ActiveElevators peerUpdate, ActiveElevatorUpdate chan utils.Status) {

	if ActiveElevators.New != "" {
		newElevatorID, _ := strconv.Atoi(ActiveElevators.New)
		UpdateElevatorsOnNetwork(newElevatorID, true, ActiveElevatorUpdate)
	}

	if ActiveElevators.Lost != nil {
		for i := range ActiveElevators.Lost {
			lostElevatorID, _ := strconv.Atoi(ActiveElevators.Lost[i])
			UpdateElevatorsOnNetwork(lostElevatorID, false, ActiveElevatorUpdate)
		}
	}

}

func UpdateElevatorsOnNetwork(elevatorID int, isActive bool, ActiveElevatorUpdate chan utils.Status) {

	if isActive {
		ActiveElevatorIDs = appendElevatorID(ActiveElevatorIDs, elevatorID)
		ActiveElevatorUpdate <- utils.Status{ID: elevatorID, IsOnline: true}
	} else {
		ActiveElevatorIDs = removeElevatorID(ActiveElevatorIDs, elevatorID)
		ActiveElevatorUpdate <- utils.Status{ID: elevatorID, IsOnline: false}
	}
}

func appendElevatorID(slice []int, value int) []int {
	for _, item := range slice {
		if item == value {
			return slice // Return the original slice if value already exists
		}
	}
	return append(slice, value) // Append value to slice if it doesn't exist
}

func removeElevatorID(slice []int, value int) []int {
	for i, item := range slice {
		if item == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // Return the original slice if value doesn't exist
}

func ProcessElevatorStatus(new utils.Elevator) {

	found := false
	for i, elevator := range Elevators {
		if elevator.ID == new.ID {
			found = true
			Elevators[i] = new
			break
		}
	}

	if !found {
		Elevators = append(Elevators, new)
	}

}

func IsEqual(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Ints(a)
	sort.Ints(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func UpdatePeers(prev []int, new []int, IsOnlineCh chan bool) {

	peers, _, _ := Compare(prev, new)

	if len(peers) == 0 && len(new) == 0 {
		IsOnlineCh <- false
		return
	}

}

func SmallestSlice(a []int, b []int) int {
	if len(a) < len(b) {
		return len(a)
	}
	return len(b)
}

func Compare(prev []int, new []int) (peers []string, newValues []string, lost []string) {
	mPrev := make(map[int]bool)
	mNew := make(map[int]bool)

	for _, value := range prev {
		mPrev[value] = true
	}

	for _, value := range new {
		mNew[value] = true
		if mPrev[value] {
			peers = append(peers, strconv.Itoa(value))
		} else {
			newValues = append(newValues, strconv.Itoa(value))
		}
	}

	for _, value := range prev {
		if !mNew[value] {
			lost = append(lost, strconv.Itoa(value))
		}
	}

	return peers, newValues, lost
}

func HandleMasterUpdate(val int, e utils.Elevator) {
	utils.MasterMutex.Lock()
	utils.MasterIDmutex.Lock()
	fmt.Println("Master update: ", val)
	if val == e.ID {
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
}

func HandleNewOrder(newOrder utils.MessageNewOrder, LocalOrders [3][utils.NumFloors]bool, e utils.Elevator, ch chan interface{},
	GlobalUpdateCh chan utils.GlobalOrderUpdate, DoOrderCh chan utils.Order,
	WatcherUpdate chan utils.OrderWatcher, IsOnlineCh chan bool, isOnline bool) [3][utils.NumFloors]bool {

	order := newOrder.NewOrder

	if utils.Master && e.ID != newOrder.FromElevatorID && order.Button != utils.Cab {

		fmt.Println("---NEW ORDER TO DELEGATE---")

		ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh, WatcherUpdate, IsOnlineCh, LocalOrders, isOnline)

	} else if !utils.Master && newOrder.ToElevatorID == e.ID && newOrder.FromElevatorID != e.ID {

		fmt.Println("---NEW ORDER RECEIVED---")

		GlobalUpdateCh <- utils.GlobalOrderUpdate{
			Order:          order,
			ForElevatorID:  e.ID,
			FromElevatorID: newOrder.FromElevatorID,
			IsComplete:     false,
			IsNew:          true,
		}

		time.Sleep(100 * time.Millisecond)

		DoOrderCh <- order

	} else if (newOrder.ToElevatorID != e.ID && !utils.Master) || (order.Button == utils.Cab) {

		fmt.Println("Sending to updater...")

		if order.Button == utils.Cab {

			go func() {
				GlobalUpdateCh <- utils.GlobalOrderUpdate{
					Order:          order,
					ForElevatorID:  newOrder.FromElevatorID,
					FromElevatorID: newOrder.FromElevatorID,
					IsComplete:     false,
					IsNew:          true}
			}()

		} else if newOrder.FromElevatorID == utils.MasterID && order.Button != utils.Cab {

			go func() {
				GlobalUpdateCh <- utils.GlobalOrderUpdate{
					Order:          order,
					ForElevatorID:  newOrder.ToElevatorID,
					FromElevatorID: newOrder.FromElevatorID,
					IsComplete:     false,
					IsNew:          true}
			}()
		}
	}

	return LocalOrders
}

func HandleButtonEvent(newOrder elevio.ButtonEvent, e utils.Elevator, ch chan interface{},
	GlobalUpdateCh chan utils.GlobalOrderUpdate, LocalOrders [3][utils.NumFloors]bool, DoOrderCh chan utils.Order, WatcherUpdate chan utils.OrderWatcher,
	IsOnlineCh chan bool, isOnline bool) {
	fmt.Println("---LOCAL BUTTON PRESSED---")

	order := utils.Order{
		Floor:  newOrder.Floor,
		Button: newOrder.Button,
	}

	fmt.Println("Button pressed: ", order)
	fmt.Println("Local order array:")
	fmt.Println(LocalOrders)

	if order.Button == utils.Cab {

		go func() {
			GlobalUpdateCh <- utils.GlobalOrderUpdate{
				Order:          order,
				ForElevatorID:  e.ID,
				FromElevatorID: e.ID,
				IsComplete:     false,
				IsNew:          true}
		}()

		go func() {
			fmt.Println("Master ID: ", utils.MasterID)
			SendOrder(order, e, ch, utils.MasterID)
		}()

		fmt.Println("Sending cab order to FSM...")
		DoOrderCh <- order // Send to FSM

	} else {

		ProcessNewOrder(order, e, ch, GlobalUpdateCh, DoOrderCh, WatcherUpdate, IsOnlineCh, LocalOrders, isOnline)

	}

}
