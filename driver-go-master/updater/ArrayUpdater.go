package updater

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
	"github.com/runarto/Heislab-Sanntid/watchdog"
)

var CabOrders [utils.NumOfElevators][utils.NumFloors]bool
var HallOrders map[int][2][utils.NumFloors]bool

var MasterOrderWatcher utils.OrderWatcherArray
var SlaveOrderWatcher utils.OrderWatcherArray

func LocalUpdater(e utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate, OrderWatcher chan utils.OrderWatcher,
	LocalStateUpdateCh chan utils.Elevator, ch chan interface{}, LocalLightsCh chan [2][utils.NumFloors]bool,
	ButtonCh chan elevio.ButtonEvent, IsOnlineCh chan bool, ActiveElevatorUpdate <-chan utils.Status, DoOrderCh chan utils.Order,
	MasterUpdateCh chan int) {

	MasterBarkCh := make(chan utils.Order)
	SlaveBarkCh := make(chan utils.Order)
	go watchdog.Watchdog(e, &MasterOrderWatcher, &SlaveOrderWatcher, MasterBarkCh, SlaveBarkCh, ButtonCh, ch)

	fmt.Println("Updater started")

	for {

		select {
		case GlobalUpdate := <-GlobalUpdateCh:

			fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")

			isNew := GlobalUpdate.IsNew

			switch isNew {
			case true: // New order
				UpdateGlobalOrderArray(true, false, GlobalUpdate.Order, e, GlobalUpdate.ForElevatorID, OrderWatcher,
					LocalLightsCh, ch, IsOnlineCh, &CabOrders, &HallOrders)

			case false: // Coplete order
				UpdateGlobalOrderArray(false, true, GlobalUpdate.Order, e, GlobalUpdate.ForElevatorID, OrderWatcher,
					LocalLightsCh, ch, IsOnlineCh, &CabOrders, &HallOrders)

			}

		case WatcherUpdate := <-OrderWatcher:

			fmt.Println("---ORDER WATCHER UPDATE RECEIVED---")

			IsNew := WatcherUpdate.IsNew

			switch IsNew {
			case true:

				UpdateWatcher(WatcherUpdate, WatcherUpdate.Order, e, &MasterOrderWatcher, &SlaveOrderWatcher)

			case false:

				UpdateWatcher(WatcherUpdate, WatcherUpdate.Order, e, &MasterOrderWatcher, &SlaveOrderWatcher)

			}

		case s := <-LocalStateUpdateCh: // Update the local elevator instance

			UpdateAndSendNewState(&e, s, ch, GlobalUpdateCh)

		case elevatorID := <-ActiveElevatorUpdate:

			fmt.Println("---ACTIVE ELEVATOR UPDATE RECEIVED---")

			UpdateActiveElevators(elevatorID, CabOrders, ch, DoOrderCh, MasterUpdateCh)

		}

	}

}

func GlobalUpdater(ElevStatus <-chan utils.MessageElevatorStatus, MasterOrderWatcherUpdate <-chan utils.MessageOrderWatcher) {

	for {

		select {

		case activeElevator := <-ElevStatus:

			HandleActiveElevators(activeElevator)

		case copy := <-MasterOrderWatcherUpdate:

			CopyMasterOrderWatcher(copy, &MasterOrderWatcher, &CabOrders)

		}
	}

}

func BroadcastMasterOrderWatcher(e utils.Elevator, ch chan interface{}) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for range ticker.C {

		if utils.Master {

			msg := utils.PackMessage("MessageOrderWatcher", HallOrders[e.ID], CabOrders, e.ID)
			ch <- msg
		}
	}
}
