package updater

import (
	"fmt"
	"sync"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
	"github.com/runarto/Heislab-Sanntid/watchdog"
)

var (
	ordersMutex sync.Mutex
	Orders      = InitOrders()
)

var MasterOrderWatcher utils.OrderWatcherArray
var SlaveOrderWatcher utils.OrderWatcherArray
var bufferSize = 100

func LocalUpdater(e utils.Elevator, GlobalUpdateCh chan utils.GlobalOrderUpdate, OrderWatcher chan utils.OrderWatcher,
	LocalStateUpdateCh chan utils.Elevator, ch chan interface{}, ButtonCh chan elevio.ButtonEvent, IsOnlineCh chan bool, ActiveElevatorUpdate <-chan utils.Status, DoOrderCh chan utils.Order,
	MasterUpdateCh chan int, OrdersUpdate chan map[int][utils.NumButtons][utils.NumFloors]bool,
	LocalOrders chan [utils.NumButtons][utils.NumFloors]bool, SendLights chan [2][utils.NumFloors]bool) {

	fmt.Println(Orders)
	MasterBarkCh := make(chan utils.Order, bufferSize)
	SlaveBarkCh := make(chan utils.Order, bufferSize)
	go watchdog.Watchdog(e, &MasterOrderWatcher, &SlaveOrderWatcher, MasterBarkCh, SlaveBarkCh, ButtonCh, ch)

	fmt.Println("Updater started")

	for {
		select {
		case GlobalUpdate := <-GlobalUpdateCh:
			fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")
			UpdateGlobalOrderArray(GlobalUpdate, e, OrderWatcher, ch, IsOnlineCh, &Orders, OrdersUpdate, SendLights)
		case WatcherUpdate := <-OrderWatcher:
			fmt.Println("---ORDER WATCHER UPDATE RECEIVED---")
			UpdateWatcher(WatcherUpdate, WatcherUpdate.Order, e, &MasterOrderWatcher, &SlaveOrderWatcher)
		case s := <-LocalStateUpdateCh: // Update the local elevator instance
			UpdateAndSendNewState(&e, s, ch, GlobalUpdateCh, Orders, LocalOrders)
		case elevatorID := <-ActiveElevatorUpdate:
			fmt.Println("---ACTIVE ELEVATOR UPDATE RECEIVED---")
			UpdateActiveElevators(elevatorID, Orders, ch, DoOrderCh, MasterUpdateCh, GlobalUpdateCh)
		}
	}
}

func GlobalUpdater(ElevStatus <-chan utils.MessageElevatorStatus, MasterOrderWatcherRx <-chan utils.MessageOrderWatcher) {

	for {
		select {
		case activeElevator := <-ElevStatus:
			HandleActiveElevators(activeElevator)
		case copy := <-MasterOrderWatcherRx:
			CopyMasterOrderWatcher(copy, &MasterOrderWatcher)
			ordersMutex.Lock()
			Orders = UpdateOrders(copy, Orders)
			ordersMutex.Unlock()
		}
	}

}

func BroadcastMasterOrderWatcher(ch chan interface{}) {

	// BroadcastAckMatrix broadcasts the acknowledgement matrix to other elevators.
	// It waits for 5 seconds before starting the broadcast and then sends the acknowledgement matrix every 5 seconds.
	// The acknowledgement matrix is sent only if there are more than one elevators and the current elevator is the master.
	// The acknowledgement matrix includes the order watcher and the ID of the current elevator.
	var OrdersForSending map[string][utils.NumButtons][utils.NumFloors]bool

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for range ticker.C {
		time.Sleep(5 * time.Millisecond)
		if utils.Master {
			OrdersForSending = Map_IntToString(Orders)
			msg := utils.PackMessage("MessageOrderWatcher", OrdersForSending, utils.ID)
			ch <- msg
		}
	}
}
