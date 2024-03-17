package updater

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
// * @brief      {Updates the global order array based on the given parameters}
// *
// * @param      ordersUpdate     An order, either new or complete
// * @param      e                The elevator
// * @param      OrderWatcherCh   The order watcher channel
// * @param      Orders           The orders for all elevators
// * @param      ActiveOrdersCh   Channel for updating the orders that are being broadcasted
// */

func UpdateGlobalOrderArray(ordersUpdate utils.GlobalOrderUpdate, e utils.Elevator, OrderWatcherCh chan utils.OrderWatcher,
	Orders *map[int][3][utils.NumFloors]bool, ActiveOrdersCh chan map[int][3][utils.NumFloors]bool) {

	if !utils.Master && utils.MasterID != utils.NotDefined {
		return
	}

	fmt.Println("---GLOBAL ORDER UPDATE RECEIVED---")

	fmt.Println("Function: UpdateGlobalOrderArray")

	isNew := ordersUpdate.IsNew
	o := ordersUpdate.Order
	forElevatorID := ordersUpdate.ForElevatorID

	change := false
	utils.OrdersMutex.Lock()
	temp := *Orders

	switch isNew {
	case true:
		if !IsOrderActive(temp, o) {
			fmt.Println("Update: New order")
			temp = Update(forElevatorID, o, temp, true)
			change = true
		}

	case false:
		if IsOrderActive(temp, o) {
			fmt.Println("Update: Order complete")
			temp = Update(forElevatorID, o, temp, false)
			change = true
		}
	}

	*Orders = temp
	utils.OrdersMutex.Unlock()

	if change && ordersUpdate.Order.Button != utils.Cab {
		ActiveOrdersCh <- *Orders
		SendWatcherUpdateIfChanged(e, ordersUpdate, OrderWatcherCh)
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

/*
 * @brief      {Checks if the specified order is active}
 *
 * @param      Orders  The orders for all elevators
 * @param      Order   The order
 *
 * @return     {True if the order is active, false otherwise}
 */

func IsOrderActive(Orders map[int][3][utils.NumFloors]bool, Order utils.Order) bool {
	for e := range Orders {
		if Orders[e][Order.Button][Order.Floor] {
			return true
		}
	}
	return false
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the orders for the specified elevator}
// *
// * @param      id     The elevator ID
// * @param      o      The order
// * @param      Orders The orders for all elevators
// * @param      isNew  Indicates if the order is new or complete
// *
// * @return     {The updated orders}
// */
func Update(id int, o utils.Order, Orders map[int][3][utils.NumFloors]bool, isNew bool) map[int][3][utils.NumFloors]bool {

	switch isNew {
	case true:
		copy := Orders[id]
		copy[o.Button][o.Floor] = true
		Orders[id] = copy
	case false:
		for ids := range Orders {
			copy := Orders[ids]
			copy[o.Button][o.Floor] = false
			Orders[ids] = copy
		}
	}
	return Orders
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the order watcher based on the given parameters}
// *
// * @param      WatcherUpdate  The order watcher update
// * @param      o              The order
// * @param      e              The elevator
// * @param      m              The master order watcher
// * @param      s              The slave order watcher
// */
func UpdateWatcher(WatcherUpdate utils.OrderWatcher, o utils.Order, e utils.Elevator, m *utils.OrderWatcherArray,
	s *utils.OrderWatcherArray) {

	fmt.Println("---ORDER WATCHER UPDATE RECEIVED---")

	isNew := WatcherUpdate.IsNew
	isComplete := WatcherUpdate.IsComplete
	isConfirmed := WatcherUpdate.IsConfirmed

	if utils.Master {

		MasterOrderWatcherUpdate(isNew, isComplete, o, e, m)

	} else {

		SlaveOrderWatcherUpdate(isNew, isConfirmed, o, e, s)

	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the master order watcher based on the given parameters}
// *
// * @param      isNew       Indicates if the order is new
// * @param      isComplete  Indicates if the order is complete
// * @param      o           The order
// * @param      e           The elevator
// * @param      m           The master order watcher
// */
func MasterOrderWatcherUpdate(isNew bool, isComplete bool, o utils.Order, e utils.Elevator, m *utils.OrderWatcherArray) {

	if isNew && !isComplete && o.Button != utils.Cab {
		m.WatcherMutex.Lock()

		m.HallOrderArray[o.Button][o.Floor].Active = true
		m.HallOrderArray[o.Button][o.Floor].Completed = false
		m.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		m.WatcherMutex.Unlock()
	}

	if isComplete && o.Button != utils.Cab {
		m.WatcherMutex.Lock()

		m.HallOrderArray[o.Button][o.Floor].Active = false
		m.HallOrderArray[o.Button][o.Floor].Completed = true
		m.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		m.WatcherMutex.Unlock()
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the slave order watcher based on the given parameters}
// *
// * @param      isNew        Indicates if the order is new
// * @param      isConfirmed  Indicates if the order is confirmed
// * @param      o            The order
// * @param      e            The elevator
// * @param      s            The slave order watcher
// */
func SlaveOrderWatcherUpdate(isNew bool, isConfirmed bool, o utils.Order, e utils.Elevator, s *utils.OrderWatcherArray) {

	if isNew && !isConfirmed && o.Button != utils.Cab {
		s.WatcherMutex.Lock()

		s.HallOrderArray[o.Button][o.Floor].Active = true
		s.HallOrderArray[o.Button][o.Floor].Confirmed = false
		s.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		s.WatcherMutex.Unlock()
	}

	if isConfirmed && o.Button != utils.Cab {
		s.WatcherMutex.Lock()

		s.HallOrderArray[o.Button][o.Floor].Active = false
		s.HallOrderArray[o.Button][o.Floor].Confirmed = false
		s.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		s.WatcherMutex.Unlock()
	}

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Copies the master order watcher from the master to the slave}
// *
// * @param      copy       The master order watcher
// * @param      m          The slave order watcher
// * @param      DoOrderCh  Channel for doing orders
// */
func CopyMasterOrderWatcher(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray, DoOrderCh chan utils.Order) {

	if !utils.Master {
		m.WatcherMutex.Lock()
		UpdateOrderWatcherArray(copy, m)
		m.WatcherMutex.Unlock()
	}

}

// *--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the order watcher array based on the given parameters, and updates the orders}
// *
// * @param      copy  The map of orders from the master
// * @param      m     The master order watcher
// */
func UpdateOrderWatcherArray(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray) {

	if utils.Master {
		return
	}

	NewOrders := utils.InitOrders()

	receivedOrders := utils.Map_StringToInt(copy.Orders)
	size := len(receivedOrders)

	for e := 0; e < size; e++ {
		temp := NewOrders[e]
		for b := 0; b < 2; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if receivedOrders[e][b][f] || m.HallOrderArray[b][f].Active {
					m.HallOrderArray[b][f].Active = true
					m.HallOrderArray[b][f].Completed = false
					m.HallOrderArray[b][f].Time = time.Now()
					temp[b][f] = true
				} else {
					m.HallOrderArray[b][f].Active = false
					m.HallOrderArray[b][f].Completed = true
					temp[b][f] = false
				}
			}
		}
		NewOrders[e] = temp
	}
	utils.OrdersMutex.Lock()
	utils.Orders = NewOrders
	fmt.Println("Updated orders: ", utils.Orders)
	utils.OrdersMutex.Unlock()
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Handles the active elevators based on the given parameters}
// *
// * @param      new  The new elevator status
// */
func UpdateOrAddActiveElevator(new utils.MessageElevatorStatus) {

	utils.ElevatorsMutex.Lock()
	found := false
	for i, e := range utils.Elevators {
		if e.ID == new.Elevator.ID {
			utils.Elevators[i].IsActive = true
			utils.Elevators[i] = new.Elevator
			found = true
		}
	}
	if !found {
		utils.Elevators = append(utils.Elevators, new.Elevator)
	}
	utils.ElevatorsMutex.Unlock()
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the active elevators based on the given parameters}
// *
// * @param      status           The status of the elevator
// * @param      Orders           The orders for all elevators
// * @param      messageHandler   Channel for resending orders
// * @param      AllOrdersCh   	  The global update channel
// * @param      ButtonPressCh    The button press channel
// */

func UpdateElevatorStatusAndHandleOrders(status utils.Status, Orders map[int][3][utils.NumFloors]bool,
	messageHandler chan utils.Message, AllOrdersCh chan utils.GlobalOrderUpdate, ButtonPressCh chan elevio.ButtonEvent) {

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Function: UpdateActiveElevators")
	if !status.IsOnline {
		fmt.Println("Elevator ", status.ID, " is offline")
		utils.ElevatorsMutex.Lock()
		UpdateElevatorActiveStatus(status.ID, status.IsOnline)
		RedistributeHallOrders(status.ID, utils.Elevators, messageHandler, AllOrdersCh, ButtonPressCh)
		utils.ElevatorsMutex.Unlock()
	} else {
		utils.ElevatorsMutex.Lock()
		UpdateElevatorActiveStatus(status.ID, status.IsOnline)
		utils.ElevatorsMutex.Unlock()
		SendCabOrders(Orders, status.ID, messageHandler)
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the active status of the elevator based on the given parameters}
// *
// * @param      id      The elevator ID
// * @param      online  Indicates if the elevator is online
// */
func UpdateElevatorActiveStatus(id int, online bool) {
	for i := range utils.Elevators {
		if utils.Elevators[i].ID == id {
			utils.Elevators[i].IsActive = online
			break
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Redistributes the hall orders based on the given parameters}
// *
// * @param      id                The elevator ID
// * @param      elevators         The elevators
// * @param      messageHandler    Channel for resending orders
// * @param      AllOrdersCh    	The global update channel
// * @param      ButtonPressCh     The button press channel
// */
func RedistributeHallOrders(id int, elevators []utils.Elevator, messageHandler chan utils.Message,
	AllOrdersCh chan utils.GlobalOrderUpdate, ButtonPressCh chan elevio.ButtonEvent) {

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Master: ", utils.MasterID)
	if !utils.Master && utils.MasterID != utils.NotDefined {
		return
	}

	fmt.Println("Function: RedistributeHallOrders")

	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if utils.Orders[id][b][f] {
				switch utils.MasterID {
				case utils.NotDefined:
					ButtonPressCh <- elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}

				default:
					BestElevator := utils.ChooseElevator(utils.Order{Floor: f, Button: elevio.ButtonType(b)})
					msg := utils.PackMessage("MessageNewOrder", BestElevator.ID, utils.ID, utils.Order{Floor: f, Button: elevio.ButtonType(b)})
					fmt.Println("Redistributed hall orders")
					messageHandler <- msg
					time.Sleep(50 * time.Millisecond)
					SendToGlobalUpdateChannel(b, f, BestElevator.ID, id, AllOrdersCh)
				}
			}
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Sends the cab orders based on the given parameters}
// *
// * @param      CabOrders        The cab orders
// * @param      id               The elevator ID
// * @param      messageHandler   Channel for resending orders
// */
func SendCabOrders(CabOrders map[int][utils.NumOfElevators][utils.NumFloors]bool, id int, messageHandler chan utils.Message) {

	if !utils.Master {
		return
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Function: SendCabOrders")
	for f := 0; f < utils.NumFloors; f++ {
		if utils.Orders[id][2][f] {
			fmt.Println("Elevator restart. Sending cab order at floor", f, " to elevator ", id)
			msg := utils.PackMessage("MessageNewOrder", id, utils.ID, utils.Order{Floor: f, Button: elevio.BT_Cab})
			messageHandler <- msg
			time.Sleep(50 * time.Millisecond)
			continue
		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Gets the active elevators}
// *
// * @return     {The active elevators}
func GetActiveElevators() []int {
	var activeElevators []int
	utils.ElevatorsMutex.Lock()
	for _, e := range utils.Elevators {
		if e.IsActive {
			activeElevators = append(activeElevators, e.ID)
		}
	}
	utils.ElevatorsMutex.Unlock()
	return activeElevators
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Sends the order watcher update if the order has changed}
// *
// * @param      e             The elevator
// * @param      GlobalUpdate  The global update
// * @param      orderWatcher  The order watcher channel
// */
func SendWatcherUpdateIfChanged(e utils.Elevator, AllOrdersCh utils.GlobalOrderUpdate, orderWatcher chan utils.OrderWatcher) {

	isNew := AllOrdersCh.IsNew
	isComplete := AllOrdersCh.IsComplete
	order := AllOrdersCh.Order
	forElevatorID := AllOrdersCh.ForElevatorID

	fmt.Println("Change was true.")

	watcherUpdate := utils.OrderWatcher{
		Order:         order,
		ForElevatorID: forElevatorID,
		IsComplete:    isComplete,
		IsNew:         isNew,
		IsConfirmed:   false}

	orderWatcher <- watcherUpdate
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Sends the global update to the global update channel}
// *
// * @param      b               The button
// * @param      f               The floor
// * @param      BestElevatorID  The best elevator ID
// * @param      elevatorID      The elevator ID
// * @param      AllOrdersCh  The global update channel
// */

func SendToGlobalUpdateChannel(b int, f int, BestElevatorID int, elevatorID int, AllOrdersCh chan utils.GlobalOrderUpdate) {
	AllOrdersCh <- utils.GlobalOrderUpdate{
		Order:         utils.Order{Floor: f, Button: elevio.ButtonType(b)},
		ForElevatorID: BestElevatorID,
		IsComplete:    false,
		IsNew:         true}

	AllOrdersCh <- utils.GlobalOrderUpdate{
		Order:         utils.Order{Floor: f, Button: elevio.ButtonType(b)},
		ForElevatorID: elevatorID,
		IsComplete:    true,
		IsNew:         false}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Initializes the orders}
// *
func InitOrders() map[int][utils.NumButtons][utils.NumFloors]bool {
	Orders := make(map[int][utils.NumButtons][utils.NumFloors]bool)
	for i := 0; i < utils.NumOfElevators; i++ {
		Orders[i] = [utils.NumButtons][utils.NumFloors]bool{}
	}
	return Orders
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

// *
// * @brief      {Updates the orders based on the given parameters}
// *
// * @param      copy    The copy of the orders from the master
// * @param      Orders  The current orders
// *
// * @return     {The updated orders}
// */

func UpdateOrders(copy utils.MessageOrderWatcher, Orders map[int][3][utils.NumFloors]bool) map[int][3][utils.NumFloors]bool {

	// OrdersForSending converts the order matrix to a map with string keys.
	// It is used to send the order matrix over the network.

	OrdersReceived := utils.Map_StringToInt(copy.Orders)
	tempOrders := [3][utils.NumFloors]bool{}

	for id := 0; id < utils.NumOfElevators; id++ {
		if id == utils.ID {
			continue
		}
		for b := 0; b < utils.NumButtons; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				tempOrders[b][f] = OrdersReceived[id][b][f] || Orders[id][b][f]
			}
		}
		Orders[id] = tempOrders
		tempOrders = [3][utils.NumFloors]bool{}

	}
	return Orders
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
