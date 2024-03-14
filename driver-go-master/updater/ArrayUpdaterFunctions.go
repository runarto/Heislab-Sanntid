package updater

import (
	"fmt"
	"strconv"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func UpdateGlobalOrderArray(GlobalUpdate utils.GlobalOrderUpdate, e utils.Elevator, orderWatcher chan utils.OrderWatcher, SendLights chan [2][utils.NumFloors]bool, ch chan interface{},
	IsOnlineCh chan bool, Orders *map[int][3][utils.NumFloors]bool, ordersUpdate chan map[int][3][utils.NumFloors]bool) {

	fmt.Println("Function: UpdateGlobalOrderArray")

	isNew := GlobalUpdate.IsNew
	o := GlobalUpdate.Order
	forElevatorID := GlobalUpdate.ForElevatorID

	change := false
	temp := (*Orders)[forElevatorID]

	switch isNew {
	case true:
		if !temp[o.Button][o.Floor] {
			fmt.Println("Update: New order")
			temp[o.Button][o.Floor] = true
			change = true
		}

	case false:

		if temp[o.Button][o.Floor] {
			fmt.Println("Update: Order complete")
			temp[o.Button][o.Floor] = false
			change = true
		}
	}

	(*Orders)[forElevatorID] = temp

	if utils.Master {
		fmt.Println(Orders)
	}

	if change {
		go SendOrderUpdateIfChanged(*Orders, ordersUpdate, forElevatorID)
		go SendWatcherUpdateIfChanged(e, GlobalUpdate, orderWatcher)
		go SendLightsIfChanged(e, *Orders, ch, SendLights)
	}
}

func Printlights(lights [2][utils.NumFloors]bool) {
	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			fmt.Print(lights[b][f], " ")
		}
		fmt.Println()
	}

}

func LightsToSend(Orders map[int][3][utils.NumFloors]bool) [2][utils.NumFloors]bool {
	var Lights [2][utils.NumFloors]bool
	for id := range Orders {
		for b := 0; b < 2; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if Orders[id][b][f] {
					Lights[b][f] = true
				}
			}
		}
	}
	return Lights
}

func UpdateWatcher(WatcherUpdate utils.OrderWatcher, o utils.Order, e utils.Elevator, m *utils.OrderWatcherArray,
	s *utils.OrderWatcherArray) {

	isNew := WatcherUpdate.IsNew
	isComplete := WatcherUpdate.IsComplete
	isConfirmed := WatcherUpdate.IsConfirmed

	if isNew && !isConfirmed && o.Button != utils.Cab {
		m.WatcherMutex.Lock()
		s.WatcherMutex.Lock()

		m.HallOrderArray[o.Button][o.Floor].Active = true
		m.HallOrderArray[o.Button][o.Floor].Completed = false
		m.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		s.HallOrderArray[o.Button][o.Floor].Active = true
		s.HallOrderArray[o.Button][o.Floor].Confirmed = false
		s.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		s.WatcherMutex.Unlock()
		m.WatcherMutex.Unlock()
	}

	if isComplete && o.Button != utils.Cab {
		m.WatcherMutex.Lock()

		m.HallOrderArray[o.Button][o.Floor].Active = false
		m.HallOrderArray[o.Button][o.Floor].Completed = true
		m.HallOrderArray[o.Button][o.Floor].Time = time.Now()

		m.WatcherMutex.Unlock()

	} else if isConfirmed && o.Button != utils.Cab {
		s.WatcherMutex.Lock()
		s.HallOrderArray[o.Button][o.Floor].Active = false
		s.HallOrderArray[o.Button][o.Floor].Confirmed = true
		s.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		s.WatcherMutex.Unlock()
	}
}

func UpdateAndSendNewState(e *utils.Elevator, s utils.Elevator, ch chan interface{}, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	Orders map[int][3][utils.NumFloors]bool, LocalOrders chan [3][utils.NumFloors]bool) {

	ReadAndSendOrdersDone(*e, s, ch, GlobalUpdateCh, Orders, LocalOrders)
	time.Sleep(100 * time.Millisecond)
	*e = s
	HandleActiveElevators(utils.MessageElevatorStatus{FromElevator: *e})
	msg := utils.PackMessage("MessageElevatorStatus", *e)
	ch <- msg

}

func ReadAndSendOrdersDone(e utils.Elevator, s utils.Elevator, ch chan interface{}, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	Orders map[int][3][utils.NumFloors]bool, LocalOrders chan [3][utils.NumFloors]bool) {

	fmt.Println("Func: ReadAndSendOrdersDone")

	// fmt.Println("Old State")
	// utils.PrintLocalOrderArray(*e)
	// fmt.Println("New State")
	// utils.PrintLocalOrderArray(s)

	OldOrders := Orders[e.ID]
	NewOrders := s.LocalOrderArray

	LocalOrders <- NewOrders

	for b := 0; b < 3; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if OldOrders[b][f] && !NewOrders[b][f] {

				msg := utils.PackMessage("MessageOrderComplete", utils.Order{Floor: f, Button: elevio.ButtonType(b)}, utils.MasterID, e.ID)
				ch <- msg

				GlobalUpdateCh <- utils.GlobalOrderUpdate{
					Order:          utils.Order{Floor: f, Button: elevio.ButtonType(b)},
					ForElevatorID:  e.ID,
					FromElevatorID: e.ID,
					IsComplete:     true,
					IsNew:          false}
			}
		}
	}
}

func CopyMasterOrderWatcher(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray) {

	if !utils.Master {
		m.WatcherMutex.Lock()
		UpdateOrderWatcherArray(copy, m)
		m.WatcherMutex.Unlock()
	}

}

func UpdateOrderWatcherArray(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray) {

	HallOrders := Map_StringToInt(copy.Orders)
	size := len(HallOrders)

	for e := 0; e < size; e++ {
		for b := 0; b < 2; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if HallOrders[e][b][f] && !m.HallOrderArray[b][f].Active {
					m.HallOrderArray[b][f].Active = true
					m.HallOrderArray[b][f].Completed = false
					m.HallOrderArray[b][f].Time = time.Now()
				} else {
					m.HallOrderArray[b][f].Active = false
					m.HallOrderArray[b][f].Completed = true
				}
			}
		}
	}
}

func HandleActiveElevators(new utils.MessageElevatorStatus) {
	utils.ElevatorsMutex.Lock()
	found := false
	for i, e := range utils.Elevators {
		if e.ID == new.FromElevator.ID {
			utils.Elevators[i].IsActive = new.FromElevator.IsActive
			utils.Elevators[i] = new.FromElevator
			found = true
		}
	}
	if !found {
		utils.Elevators = append(utils.Elevators, new.FromElevator)
	}
	utils.ElevatorsMutex.Unlock()
}

func UpdateActiveElevators(status utils.Status, Orders map[int][3][utils.NumFloors]bool,
	ch chan interface{}, DoOrderCh chan utils.Order, MasterUpdateCh chan int, GlobalUpdateCh chan utils.GlobalOrderUpdate) {

	fmt.Println("Function: UpdateActiveElevators")
	if !status.IsOnline {
		utils.ElevatorsMutex.Lock()
		SearchForElevatorAndUpdate(status.ID, status.IsOnline)
		RedistributeHallOrders(status.ID, utils.Elevators, ch, DoOrderCh, GlobalUpdateCh)
		utils.ElevatorsMutex.Unlock()
	} else {
		utils.ElevatorsMutex.Lock()
		SearchForElevatorAndUpdate(status.ID, status.IsOnline)
		utils.ElevatorsMutex.Unlock()
		SendCabOrders(Orders, status.ID, ch)
	}

	DetermineMaster(utils.Elevators, MasterUpdateCh)
}

func SearchForElevatorAndUpdate(id int, online bool) {
	for i := range utils.Elevators {
		if utils.Elevators[i].ID == id {
			utils.Elevators[i].IsActive = online
			break
		}
	}
}

func RedistributeHallOrders(id int, elevators []utils.Elevator, ch chan interface{},
	DoOrderCh chan utils.Order, GlobalUpdateCh chan utils.GlobalOrderUpdate) {

	if !utils.Master {
		return
	}
	fmt.Println("Function: RedistributeHallOrders")
	var elevator utils.Elevator
	for _, e := range elevators {
		if e.ID == id {
			elevator = e
			break
		}
	}
	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if elevator.LocalOrderArray[b][f] {
				BestElevator := orders.ChooseElevator(utils.Order{Floor: f, Button: elevio.ButtonType(b)})
				if BestElevator.ID != utils.ID {
					msg := utils.PackMessage("MessageNewOrder", utils.Order{Floor: f, Button: elevio.ButtonType(b)}, BestElevator.ID, elevator.ID)
					ch <- msg
				} else {
					DoOrderCh <- utils.Order{Floor: f, Button: elevio.ButtonType(b)}
				}
				SendToGlobalUpdateChannel(b, f, BestElevator.ID, elevator.ID, GlobalUpdateCh)
			}
		}
	}
}

func SendCabOrders(CabOrders map[int][utils.NumOfElevators][utils.NumFloors]bool, id int, ch chan interface{}) {

	if !utils.Master {
		return
	}
	fmt.Println("Function: SendCabOrders")
	for f := 0; f < utils.NumFloors; f++ {
		if Orders[id][2][f] {
			msg := utils.PackMessage("MessageNewOrder", utils.Order{Floor: f, Button: elevio.BT_Cab}, id, utils.ID)
			ch <- msg
		}
	}
}

func GetActiveElevators() []int {
	var activeElevators []int
	utils.ElevatorsMutex.Lock()
	for _, e := range utils.Elevators {
		activeElevators = append(activeElevators, e.ID)
	}
	utils.ElevatorsMutex.Unlock()
	return activeElevators
}

func DetermineMaster(Elevators []utils.Elevator, MasterUpdateCh chan int) {

	fmt.Println("Function: DetermineMaster")
	fmt.Println("Master ID: ", utils.MasterID)

	if utils.NextMasterID != utils.MasterID {
		MasterUpdateCh <- utils.NextMasterID
		return
	}
}

func SendWatcherUpdateIfChanged(e utils.Elevator, GlobalUpdate utils.GlobalOrderUpdate, orderWatcher chan utils.OrderWatcher) {

	isNew := GlobalUpdate.IsNew
	isComplete := GlobalUpdate.IsComplete
	order := GlobalUpdate.Order
	forElevatorID := GlobalUpdate.ForElevatorID

	fmt.Println("Change was true.")

	watcherUpdate := utils.OrderWatcher{
		Order:         order,
		ForElevatorID: forElevatorID,
		IsComplete:    isComplete,
		IsNew:         isNew,
		IsConfirmed:   false}

	orderWatcher <- watcherUpdate
}

func SendLightsIfChanged(e utils.Elevator, Orders map[int][3][utils.NumFloors]bool,
	ch chan interface{}, SendLights chan [2][utils.NumFloors]bool) {

	if utils.Master {
		Lights := LightsToSend(Orders)
		SendLights <- Lights
	}
}

func SendOrderUpdateIfChanged(Orders map[int][3][utils.NumFloors]bool, OrdersUpdate chan map[int][3][utils.NumFloors]bool, forElevatorID int) {
	OrdersUpdate <- Orders
}

func SendToGlobalUpdateChannel(b int, f int, BestElevatorID int, elevatorID int, GlobalUpdateCh chan utils.GlobalOrderUpdate) {
	GlobalUpdateCh <- utils.GlobalOrderUpdate{
		Order:          utils.Order{Floor: f, Button: elevio.ButtonType(b)},
		ForElevatorID:  BestElevatorID,
		FromElevatorID: elevatorID,
		IsComplete:     false,
		IsNew:          true}

	GlobalUpdateCh <- utils.GlobalOrderUpdate{
		Order:          utils.Order{Floor: f, Button: elevio.ButtonType(b)},
		ForElevatorID:  elevatorID,
		FromElevatorID: elevatorID,
		IsComplete:     true,
		IsNew:          false}
}

func Map_IntToString(Orders map[int][utils.NumButtons][utils.NumFloors]bool) map[string][utils.NumButtons][utils.NumFloors]bool {

	// OrdersForSending converts the order matrix to a map with string keys.
	// It is used to send the order matrix over the network.

	OrdersForSending := make(map[string][utils.NumButtons][utils.NumFloors]bool)

	for id, orderMatrix := range Orders {
		OrdersForSending[fmt.Sprint(id)] = orderMatrix
	}

	return OrdersForSending

}

func Map_StringToInt(OrdersReceived map[string][utils.NumButtons][utils.NumFloors]bool) map[int][utils.NumButtons][utils.NumFloors]bool {

	// OrdersForSending converts the order matrix to a map with string keys.
	// It is used to send the order matrix over the network.

	Orders := make(map[int][utils.NumButtons][utils.NumFloors]bool)

	for id, orderMatrix := range OrdersReceived {
		intID, _ := strconv.Atoi(id)
		Orders[intID] = orderMatrix
	}

	return Orders
}

func InitOrders() map[int][utils.NumButtons][utils.NumFloors]bool {
	Orders := make(map[int][utils.NumButtons][utils.NumFloors]bool)
	for i := 0; i < utils.NumOfElevators; i++ {
		Orders[i] = [utils.NumButtons][utils.NumFloors]bool{}
	}
	return Orders
}

func UpdateOrders(copy utils.MessageOrderWatcher, Orders map[int][3][utils.NumFloors]bool) map[int][3][utils.NumFloors]bool {

	// OrdersForSending converts the order matrix to a map with string keys.
	// It is used to send the order matrix over the network.

	OrdersReceived := Map_StringToInt(copy.Orders)
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
