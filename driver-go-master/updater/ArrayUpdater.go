package updater

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/utils"
	"github.com/runarto/Heislab-Sanntid/watchdog"
)

func Updater(c *utils.Channels, e utils.Elevator) {

	for {

		select {
		case GlobalUpdate := <-c.GlobalUpdateCh:

			fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")

			isNewOrComplete := GlobalUpdate.IsNew

			switch isNewOrComplete {
			case true:
				UpdateGlobalOrderArray(true, false, GlobalUpdate.Order, e, GlobalUpdate.FromElevatorID, c)

			case false:
				UpdateGlobalOrderArray(false, true, GlobalUpdate.Order, e, GlobalUpdate.FromElevatorID, c)

			}

		case WatcherUpdate := <-c.OrderWatcher:

			isNewOrComplete := WatcherUpdate.IsNew

			switch isNewOrComplete {
			case true:

				UpdateWatcher(true, false, WatcherUpdate.Order, e, c)

			case false:

				UpdateWatcher(false, true, WatcherUpdate.Order, e, c)

			}

		case s := <-c.LocalStateUpdateCh: // Update the local elevator instance

			UpdateAndSendNewState(&e, s, c)

		case copy := <-c.MasterOrderWatcherRx:

			CopyMasterOrderWatcher(copy, &MasterOrderWatcher)

		default:

			watchdog.Watchdog(c, e, &MasterOrderWatcher, &SlaveOrderWatcher)

		}

	}

}
