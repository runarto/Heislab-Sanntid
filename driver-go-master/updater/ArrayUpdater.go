package updater

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
	"github.com/runarto/Heislab-Sanntid/watchdog"
)

func Updater(e utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate, OrderWatcher chan utils.OrderWatcher,
	LocalStateUpdateCh chan utils.Elevator, MasterOrderWatcherRx <-chan utils.MessageOrderWatcher, ch chan interface{},
	LocalLightsCh chan [2][utils.NumFloors]bool, ButtonCh chan elevio.ButtonEvent, IsOnlineCh chan bool,
	ElevStatusRx <-chan utils.MessageElevatorStatus, ActiveElevatorUpdate <-chan utils.Status) {

	var CabOrders [utils.NumOfElevators][utils.NumFloors]bool
	var HallOrders [2][utils.NumFloors]bool

	var MasterOrderWatcher utils.OrderWatcherArray
	var SlaveOrderWatcher utils.OrderWatcherArray

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
				UpdateGlobalOrderArray(true, false, GlobalUpdate.Order, e, GlobalUpdate.FromElevatorID, OrderWatcher,
					LocalLightsCh, ch, IsOnlineCh, &CabOrders, &HallOrders)

			case false: // Coplete order
				UpdateGlobalOrderArray(false, true, GlobalUpdate.Order, e, GlobalUpdate.FromElevatorID, OrderWatcher,
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

			fmt.Println("---LOCAL STATE UPDATE RECEIVED---")

			UpdateAndSendNewState(&e, s, ch)

		case copy := <-MasterOrderWatcherRx:

			fmt.Println("---MASTER ORDER WATCHER UPDATE RECEIVED---")

			CopyMasterOrderWatcher(copy, &MasterOrderWatcher)

		case activeElevator := <-ElevStatusRx:

			fmt.Println("---ELEVATOR STATUS RECEIVED---")
			HandleActiveElevators(activeElevator)

		case elevatorID := <-ActiveElevatorUpdate:

			fmt.Println("---ACTIVE ELEVATOR UPDATE RECEIVED---")

			UpdateActiveElevators(elevatorID)

		}

	}

}
