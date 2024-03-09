package orders

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/Network/peers"
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
	DoOrderCh chan utils.Order) {

	fmt.Println("Function: ProcessNewOrder")

	switch utils.Master {
	case true:
		BestElevator := ChooseElevator(order)
		fmt.Println("Best elevator for order", order, ": ", BestElevator.ID)
		//time.Sleep(5 * time.Second)
		if BestElevator.ID == e.ID {
			DoOrderCh <- order
			fmt.Println("Sending order to FSM...")
		} else {
			SendOrder(order, e, ch, BestElevator.ID)
		}

		fmt.Println("Sending message for global order update...")
		//time.Sleep(5 * time.Second)
		go func() {
			GlobalUpdateCh <- utils.GlobalOrderUpdate{
				Order:          order,
				FromElevatorID: e.ID,
				IsComplete:     false,
				IsNew:          true}
		}()

	case false:
		SendOrder(order, e, ch, utils.MasterID)

	}

}

func ProcessOrderComplete(orderComplete utils.MessageOrderComplete, e utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate) {

	fmt.Println("ProcessOrderComplete")

	GlobalOrderUpdate := utils.GlobalOrderUpdate{
		Order:          orderComplete.Order,
		FromElevatorID: orderComplete.FromElevatorID,
		IsComplete:     true,
		IsNew:          false}

	GlobalUpdateCh <- GlobalOrderUpdate

}

func HandlePeersUpdate(p peers.PeerUpdate, IsOnlineCh chan bool, MasterUpdateCh chan int, ActiveElevatorUpdate chan utils.Status) {

	fmt.Println("Function: HandlePeersUpdate")

	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)

	var ActiveElevators peerUpdate

	if len(p.Peers) == 0 {

		fmt.Println("No peers available, elevator is disconnected")

		IsOnlineCh <- false

		MasterUpdateCh <- utils.ID

		return

	} else {

		IsOnlineCh <- true

		DetermineMaster(p, MasterUpdateCh)

		ActiveElevators = HandleNewPeers(p, ActiveElevators)
		ActiveElevators = HandleLostPeers(p, ActiveElevators)

		if ActiveElevators.New != "" || ActiveElevators.Lost != nil {

			HandleActiveElevators(ActiveElevators, ActiveElevatorUpdate)

		}
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

		for i, _ := range p.Lost {
			lostElevatorID := p.Lost[i]
			peerUpdate.Lost = append(peerUpdate.Lost, lostElevatorID)
		}

	}

	return peerUpdate

}

func DetermineMaster(p peers.PeerUpdate, MasterUpdateCh chan int) {

	fmt.Println("Function: DetermineMaster")
	fmt.Println("Master ID: ", utils.MasterID)
	newMasterID, _ := strconv.Atoi(p.Peers[0])

	if newMasterID != utils.MasterID {
		fmt.Println("The new master is ", newMasterID)
		MasterUpdateCh <- newMasterID
	}

}

func HandleActiveElevators(ActiveElevators peerUpdate, ActiveElevatorUpdate chan utils.Status) {

	if ActiveElevators.New != "" {
		newElevatorID, _ := strconv.Atoi(ActiveElevators.New)
		UpdateElevatorsOnNetwork(newElevatorID, true, ActiveElevatorUpdate)
	}

	if ActiveElevators.Lost != nil {
		for i, _ := range ActiveElevators.Lost {
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

func WaitForAck(msgCh chan interface{}, e utils.Elevator, msgType string, watcher chan utils.OrderWatcher, IsOnlineCh chan bool) {
	var id_received []int

	timeout := 1 * time.Second
	timer := time.NewTimer(timeout)
	for {
		select {
		case msg := <-msgCh:
			switch m := msg.(type) {
			case utils.MessageLightsConfirmed:
				if msgType == m.Type {
					fmt.Println("Received MessageLights")
					timer.Reset(timeout)
					id_received = append(id_received, m.FromElevatorID)
					if IsEqual(id_received, ActiveElevatorIDs) {
						return
					}
				}

			case utils.MessageOrderConfirmed:
				if msgType == m.Type && m.Confirmed && m.FromElevatorID == utils.MasterID {
					fmt.Println("Received MessageOrderConfirmed")
					watcher <- utils.OrderWatcher{
						Order:         m.ForOrder,
						ForElevatorID: e.ID,
						IsComplete:    false,
						IsNew:         false,
						IsConfirmed:   true}

					timer.Reset(timeout)
					return
				}

			default:
				fmt.Printf("Unsupported message type: %T\n", m)
			}
		case <-timer.C:

			UpdatePeers(ActiveElevatorIDs, id_received, IsOnlineCh)

			fmt.Println("Timeout")

			return
		}
	}
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
