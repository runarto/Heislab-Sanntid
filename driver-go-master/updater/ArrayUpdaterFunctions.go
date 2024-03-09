package updater

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func UpdateGlobalOrderArray(isNew bool, isComplete bool, o utils.Order, e utils.Elevator,
	fromElevatorID int, orderWatcher chan utils.OrderWatcher, LocalLightsCh chan [2][utils.NumFloors]bool, ch chan interface{},
	IsOnlineCh chan bool, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool, HallOrders *[2][utils.NumFloors]bool) {

	change := false

	switch isNew {
	case true:
		if o.Button == utils.Cab && !isOrderActive(o, fromElevatorID, CabOrders, HallOrders) {
			CabOrders[fromElevatorID][o.Floor] = true
			change = true
		} else if o.Button != utils.Cab && !isOrderActive(o, fromElevatorID, CabOrders, HallOrders) {
			HallOrders[o.Button][o.Floor] = true
			change = true
		}

	case false:

		if o.Button == utils.Cab && isOrderActive(o, fromElevatorID, CabOrders, HallOrders) {
			CabOrders[fromElevatorID][o.Floor] = false
			change = true
		} else if o.Button != utils.Cab && isOrderActive(o, fromElevatorID, CabOrders, HallOrders) {
			HallOrders[o.Button][o.Floor] = false
			change = true
		}
	}

	if change {

		fmt.Println("Change was true.")

		watcherUpdate := utils.OrderWatcher{
			Order:         o,
			ForElevatorID: fromElevatorID,
			IsComplete:    isComplete,
			IsNew:         isNew,
			IsConfirmed:   false}

		orderWatcher <- watcherUpdate

	}

	if utils.Master && change {

		fmt.Println("Sending lights from master")

		LocalLightsCh <- *HallOrders

		msg := utils.PackMessage("MessageLights", *HallOrders, e.ID)
		ch <- msg

		go orders.WaitForAck(ch, e, "MessageLightsAck", orderWatcher, IsOnlineCh)

	}
}

func isOrderActive(o utils.Order, id int, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool, HallOrders *[2][utils.NumFloors]bool) bool {

	if o.Button == utils.Cab {
		return CabOrders[id][o.Floor]

	} else {
		return HallOrders[o.Button][o.Floor]
	}
}

func UpdateWatcher(WatcherUpdate utils.OrderWatcher, o utils.Order, e utils.Elevator, MasterOrderWatcher *utils.OrderWatcherArray,
	SlaveOrderWatcher *utils.OrderWatcherArray) {

	isNew := WatcherUpdate.IsNew
	isComplete := WatcherUpdate.IsComplete
	isConfirmed := WatcherUpdate.IsConfirmed

	if isNew && utils.Master && o.Button != utils.Cab {

		MasterOrderWatcher.WatcherMutex.Lock()
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = true
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = false
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		MasterOrderWatcher.WatcherMutex.Unlock()
	} else if isNew && !utils.Master && o.Button != utils.Cab {
		SlaveOrderWatcher.WatcherMutex.Lock()
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = true
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Confirmed = false
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		SlaveOrderWatcher.WatcherMutex.Unlock()
	}

	if isComplete && utils.Master && o.Button != utils.Cab {
		MasterOrderWatcher.WatcherMutex.Lock()
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = false
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = true
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		MasterOrderWatcher.WatcherMutex.Unlock()
	} else if isConfirmed && !utils.Master && o.Button != utils.Cab {
		SlaveOrderWatcher.WatcherMutex.Lock()
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = false
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Confirmed = true
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		SlaveOrderWatcher.WatcherMutex.Unlock()
	}

}

func UpdateAndSendNewState(e *utils.Elevator, s utils.Elevator, ch chan interface{}) {

	ReadAndSendOrdersDone(e, s, ch)
	*e = s
	msg := utils.PackMessage("MessageElevatorStatus", *e)
	ch <- msg

}

func ReadAndSendOrdersDone(e *utils.Elevator, s utils.Elevator, ch chan interface{}) {

	for b := 0; b < utils.NumButtons; b++ {
		for f := 0; f < utils.NumFloors; f++ {

			if e.LocalOrderArray[b][f] && !s.LocalOrderArray[b][f] { // e is previous state, s is new state. if s[b][f] is false and e[f][b] is true, then the order is done.

				msg := utils.PackMessage("MessageOrderComplete", utils.Order{Floor: f, Button: elevio.ButtonType(b)}, e.ID, s.ID)
				ch <- msg

			}
		}
	}

}

func CopyMasterOrderWatcher(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray) {

	if !utils.Master {
		m.WatcherMutex.Lock()
		m.CabOrderArray = copy.CabOrders
		m.HallOrderArray = copy.HallOrders
		m.WatcherMutex.Unlock()
	}

}

func HandleActiveElevators(new utils.MessageElevatorStatus) {
	utils.ElevatorsMutex.Lock()
	found := false
	for i, e := range utils.Elevators {
		if e.ID == new.FromElevator.ID {
			utils.Elevators[i] = new.FromElevator
			found = true
		}
	}
	if !found {
		utils.Elevators = append(utils.Elevators, new.FromElevator)
	}
	utils.ElevatorsMutex.Unlock()
}

func UpdateActiveElevators(status utils.Status) {
	if !status.IsOnline {
		utils.ElevatorsMutex.Lock()
		utils.Elevators = SearchForElevatorAndRemove(status.ID)
		utils.ElevatorsMutex.Unlock()
	}
}

func SearchForElevatorAndRemove(id int) []utils.Elevator {
	var activeElevators []utils.Elevator
	for i, e := range utils.Elevators {
		if e.ID == id {
			activeElevators = append(utils.Elevators[:i], utils.Elevators[i+1:]...)
		}
	}
	return activeElevators
}
