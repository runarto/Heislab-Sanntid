package updater

import (
	"fmt"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
	"github.com/runarto/Heislab-Sanntid/watchdog"
)

var MasterOrderWatcher utils.OrderWatcherArray
var SlaveOrderWatcher utils.OrderWatcherArray
var bufferSize = 100

func Updater(e utils.Elevator, AllOrdersCh chan utils.GlobalOrderUpdate, OrderWatcher chan utils.OrderWatcher,
	LocalStateUpdateCh chan utils.Elevator, messageHandler chan utils.Message, ActiveElevatorUpdateCh <-chan utils.Status,
	ArrayUpdater chan utils.Message, SlaveUpdateCh chan utils.Order, ActiveOrdersCh chan map[int][3][utils.NumFloors]bool,
	ButtonPressCh chan elevio.ButtonEvent) {

	utils.Orders = InitOrders()

	fmt.Println(utils.Orders)
	MasterBarkCh := make(chan utils.Order, bufferSize)
	SlaveBarkCh := make(chan utils.Order, bufferSize)
	go watchdog.Watchdog(e, &MasterOrderWatcher, &SlaveOrderWatcher, MasterBarkCh, SlaveBarkCh, messageHandler)

	fmt.Println("Updater started")

	for {
		select {
		// Master update ----------------
		case ordersUpdate := <-AllOrdersCh:
			UpdateGlobalOrderArray(ordersUpdate, e, OrderWatcher, &utils.Orders, ActiveOrdersCh)
		// Slave and master update ----------------
		case WatcherUpdate := <-OrderWatcher:
			UpdateWatcher(WatcherUpdate, WatcherUpdate.Order, e, &MasterOrderWatcher, &SlaveOrderWatcher)
		case elevatorID := <-ActiveElevatorUpdateCh:
			fmt.Println("---ACTIVE ELEVATOR UPDATE RECEIVED---")
			UpdateElevatorStatusAndHandleOrders(elevatorID, utils.Orders, messageHandler, AllOrdersCh, ButtonPressCh)

		// From network
		case msg := <-ArrayUpdater:
			switch msg.Type {
			case "ElevatorStatus":
				UpdateOrAddActiveElevator(msg.Msg.(utils.MessageElevatorStatus))
			case "OrderWatcher":
				UpdateOrderWatcherArray(msg.Msg.(utils.MessageOrderWatcher), &MasterOrderWatcher)

			}
		case confirmedOrder := <-SlaveUpdateCh:
			SlaveOrderWatcherUpdate(false, true, confirmedOrder, e, &SlaveOrderWatcher)
		}
	}
}
