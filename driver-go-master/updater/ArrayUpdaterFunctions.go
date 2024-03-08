package updater

import (
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

var CabOrders [utils.NumOfElevators][utils.NumButtons]bool
var HallOrders [2][utils.NumFloors]bool

var MasterOrderWatcher utils.OrderWatcherArray
var SlaveOrderWatcher utils.OrderWatcherArray

func UpdateGlobalOrderArray(isNew bool, isComplete bool, o utils.Order, e utils.Elevator, fromElevatorID int, c *utils.Channels) {

	change := false

	switch isNew {
	case true:
		if o.Button == utils.Cab && !isOrderActive(o, fromElevatorID) {
			CabOrders[fromElevatorID][o.Floor] = true
			change = true
		} else if o.Button != utils.Cab && !isOrderActive(o, fromElevatorID) {
			HallOrders[o.Button][o.Floor] = true
			change = true
		}

	case false:

		if o.Button == utils.Cab && isOrderActive(o, fromElevatorID) {
			CabOrders[fromElevatorID][o.Floor] = false
			change = true
		} else if o.Button != utils.Cab && isOrderActive(o, fromElevatorID) {
			HallOrders[o.Button][o.Floor] = false
			change = true
		}
	}

	if change {

		watcherUpdate := utils.OrderWatcher{
			Order:         o,
			ForElevatorID: fromElevatorID,
			IsComplete:    isComplete,
			IsNew:         isNew}

		c.OrderWatcher <- watcherUpdate

	}

	if utils.Master && change {

		c.LocalLightsCh <- HallOrders

		utils.CreateAndSendMessage(c, "MessageLights", HallOrders, e.ID)

	}
}

func isOrderActive(o utils.Order, id int) bool {

	if o.Button == utils.Cab {
		return CabOrders[id][o.Floor]

	} else {
		return HallOrders[o.Button][o.Floor]
	}
}

func UpdateWatcher(isNew bool, isComplete bool, o utils.Order, e utils.Elevator, c *utils.Channels) {

	if isNew && utils.Master {

		MasterOrderWatcher.WatcherMutex.Lock()
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = true
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = false
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		MasterOrderWatcher.WatcherMutex.Unlock()
	} else if isNew && !utils.Master {
		SlaveOrderWatcher.WatcherMutex.Lock()
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = true
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = false
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		SlaveOrderWatcher.WatcherMutex.Unlock()
	}

	if isComplete && utils.Master {
		MasterOrderWatcher.WatcherMutex.Lock()
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = false
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = true
		MasterOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		MasterOrderWatcher.WatcherMutex.Unlock()
	} else if isComplete && !utils.Master {
		SlaveOrderWatcher.WatcherMutex.Lock()
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Active = false
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Completed = true
		SlaveOrderWatcher.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		SlaveOrderWatcher.WatcherMutex.Unlock()
	}

}

func UpdateAndSendNewState(e *utils.Elevator, s utils.Elevator, c *utils.Channels) {

	ReadAndSendOrdersDone(e, s, c)

	*e = s

	utils.CreateAndSendMessage(c, "MessageElevatorStatus", e)

}

func ReadAndSendOrdersDone(e *utils.Elevator, s utils.Elevator, c *utils.Channels) {

	for b := 0; b < utils.NumButtons; b++ {
		for f := 0; f < utils.NumFloors; f++ {

			if e.LocalOrderArray[b][f] != !s.LocalOrderArray[b][f] { // e is previous state, s is new state. if s[b][f] is false and e[f][b] is true, then the order is done.

				utils.CreateAndSendMessage(c, "MessageOrderComplete", utils.Order{Floor: f, Button: elevio.ButtonType(b)}, e.ID, s.ID)

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
