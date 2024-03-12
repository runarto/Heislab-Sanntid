package updater

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func UpdateGlobalOrderArray(GlobalUpdate utils.GlobalOrderUpdate, e utils.Elevator, orderWatcher chan utils.OrderWatcher, LocalLightsCh chan [2][utils.NumFloors]bool, ch chan interface{},
	IsOnlineCh chan bool, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool, HallOrders *map[int][2][utils.NumFloors]bool) {

	isNew := GlobalUpdate.IsNew
	o := GlobalUpdate.Order
	forElevatorID := GlobalUpdate.ForElevatorID
	fromElevatorID := GlobalUpdate.FromElevatorID

	change := false
	temp := (*HallOrders)[forElevatorID]

	switch isNew {
	case true:
		if o.Button == utils.Cab && !isOrderActive(o, fromElevatorID, CabOrders, temp) {
			CabOrders[fromElevatorID][o.Floor] = true
			change = true
		} else if o.Button != utils.Cab && !isOrderActive(o, forElevatorID, CabOrders, temp) {
			temp[o.Button][o.Floor] = true
			change = true
		}

	case false:

		if o.Button == utils.Cab && isOrderActive(o, fromElevatorID, CabOrders, temp) {
			CabOrders[fromElevatorID][o.Floor] = false
			change = true
		} else if o.Button != utils.Cab && isOrderActive(o, forElevatorID, CabOrders, temp) {
			temp[o.Button][o.Floor] = false
			change = true
		}
	}

	go SendWatcherUpdateIfChanged(change, e, GlobalUpdate, orderWatcher)

	if *HallOrders == nil {
		*HallOrders = make(map[int][2][utils.NumFloors]bool)
	}
	(*HallOrders)[fromElevatorID] = temp

	go SendLightsIfChange(change, e, HallOrders, ch, LocalLightsCh)
}

func Printlights(lights [2][utils.NumFloors]bool) {
	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			fmt.Print(lights[b][f], " ")
		}
		fmt.Println()
	}

}

func LightsToSend(HallOrders map[int][2][utils.NumFloors]bool) [2][utils.NumFloors]bool {
	var Lights [2][utils.NumFloors]bool
	size := len(HallOrders)
	for id := 0; id < size; id++ {
		for b := 0; b < 2; b++ {
			for f := 0; f < utils.NumFloors; f++ {
				if HallOrders[id][b][f] {
					Lights[b][f] = true
					break
				}
			}
		}
	}
	return Lights
}

func isOrderActive(o utils.Order, id int, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool, HallOrders [2][utils.NumFloors]bool) bool {

	if o.Button == utils.Cab {
		return CabOrders[id][o.Floor]

	} else {
		return HallOrders[o.Button][o.Floor]
	}
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

	} else if isNew && isConfirmed && o.Button != utils.Cab {
		s.WatcherMutex.Lock()
		s.HallOrderArray[o.Button][o.Floor].Active = false
		s.HallOrderArray[o.Button][o.Floor].Confirmed = true
		s.HallOrderArray[o.Button][o.Floor].Time = time.Now()
		s.WatcherMutex.Unlock()
	}

	m.WatcherMutex.Lock()
	fmt.Println("Master order watcher")
	printOrderWatcher(m)
	m.WatcherMutex.Unlock()

}

func UpdateAndSendNewState(e *utils.Elevator, s utils.Elevator, ch chan interface{}, GlobalUpdateCh chan utils.GlobalOrderUpdate, HallOrders map[int][2][utils.NumFloors]bool) {

	ReadAndSendOrdersDone(*e, s, ch, GlobalUpdateCh, HallOrders)
	time.Sleep(100 * time.Millisecond)
	*e = s
	HandleActiveElevators(utils.MessageElevatorStatus{FromElevator: *e})
	msg := utils.PackMessage("MessageElevatorStatus", *e)
	ch <- msg

}

func ReadAndSendOrdersDone(e utils.Elevator, s utils.Elevator, ch chan interface{}, GlobalUpdateCh chan utils.GlobalOrderUpdate,
	HallOrders map[int][2][utils.NumFloors]bool) {

	fmt.Println("Func: ReadAndSendOrdersDone")

	// fmt.Println("Old State")
	// utils.PrintLocalOrderArray(*e)
	// fmt.Println("New State")
	// utils.PrintLocalOrderArray(s)

	OldHallOrders := HallOrders[e.ID]
	OldCabOrders := CabOrders[e.ID]

	NewCabOrders := s.LocalOrderArray[2]
	NewHallOrders := s.LocalOrderArray[0:2]

	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if OldHallOrders[b][f] && !NewHallOrders[b][f] {

				msg := utils.PackMessage("MessageOrderComplete", utils.Order{Floor: f, Button: elevio.ButtonType(b)}, utils.MasterID, e.ID)
				ch <- msg

				GlobalUpdateCh <- utils.GlobalOrderUpdate{
					Order:         utils.Order{Floor: f, Button: elevio.ButtonType(b)},
					ForElevatorID: e.ID,
					IsComplete:    true,
					IsNew:         false}
			}
		}
	}

	for f := 0; f < utils.NumFloors; f++ {
		if OldCabOrders[f] && !NewCabOrders[f] {

			msg := utils.PackMessage("MessageOrderComplete", utils.Order{Floor: f, Button: elevio.BT_Cab}, utils.MasterID, e.ID)
			ch <- msg

			GlobalUpdateCh <- utils.GlobalOrderUpdate{
				Order:         utils.Order{Floor: f, Button: elevio.BT_Cab},
				ForElevatorID: e.ID,
				IsComplete:    true,
				IsNew:         false}
		}
	}

}

func CopyMasterOrderWatcher(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool) {

	if !utils.Master {
		m.WatcherMutex.Lock()
		UpdateOrderWatcherArray(copy, m, CabOrders)
		m.WatcherMutex.Unlock()
	}

}

func UpdateOrderWatcherArray(copy utils.MessageOrderWatcher, m *utils.OrderWatcherArray, CabOrders *[utils.NumOfElevators][utils.NumFloors]bool) {

	for b := 0; b < 2; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			if copy.HallOrders[b][f] {
				m.HallOrderArray[b][f].Active = true
				m.HallOrderArray[b][f].Completed = false
				m.HallOrderArray[b][f].Time = time.Now()
			} else {
				m.HallOrderArray[b][f].Active = false
				m.HallOrderArray[b][f].Completed = true
			}
		}
	}

	for e := 0; e < utils.NumOfElevators; e++ {
		for f := 0; f < utils.NumFloors; f++ {
			if copy.CabOrders[e][f] {
				CabOrders[e][f] = true
			} else {
				CabOrders[e][f] = false
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

func UpdateActiveElevators(status utils.Status, CabOrders [utils.NumOfElevators][utils.NumFloors]bool,
	ch chan interface{}, DoOrderCh chan utils.Order, MasterUpdateCh chan int) {

	fmt.Println("Function: UpdateActiveElevators")
	if !status.IsOnline {
		utils.ElevatorsMutex.Lock()
		SearchForElevatorAndUpdate(status.ID, status.IsOnline)
		RedistributeHallOrders(status.ID, utils.Elevators, ch, DoOrderCh)
		utils.ElevatorsMutex.Unlock()
	} else {
		utils.ElevatorsMutex.Lock()
		SearchForElevatorAndUpdate(status.ID, status.IsOnline)
		utils.ElevatorsMutex.Unlock()
		SendCabOrders(CabOrders, status.ID, ch)
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
	DoOrderCh chan utils.Order) {

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
			}
		}
	}
}

func SendCabOrders(CabOrders [utils.NumOfElevators][utils.NumFloors]bool, id int, ch chan interface{}) {

	if !utils.Master {
		return
	}
	fmt.Println("Function: SendCabOrders")
	for f := 0; f < utils.NumFloors; f++ {
		if CabOrders[id][f] {
			msg := utils.PackMessage("MessageNewOrder", utils.Order{Floor: f, Button: elevio.BT_Cab}, id, utils.ID)
			ch <- msg
		}
	}
}

func printOrderWatcher(m *utils.OrderWatcherArray) {
	fmt.Println("HALL ORDERS")
	for i := 0; i < 2; i++ {
		for j := 0; j < utils.NumFloors; j++ {
			fmt.Println(m.HallOrderArray[i][j].Active, "")
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

func SendWatcherUpdateIfChanged(change bool, e utils.Elevator, GlobalUpdate utils.GlobalOrderUpdate, orderWatcher chan utils.OrderWatcher) {

	isNew := GlobalUpdate.IsNew
	isComplete := GlobalUpdate.IsComplete
	o := GlobalUpdate.Order
	forElevatorID := GlobalUpdate.ForElevatorID
	fromElevatorID := GlobalUpdate.FromElevatorID

	if change && fromElevatorID == e.ID {

		fmt.Println("Change was true.")

		watcherUpdate := utils.OrderWatcher{
			Order:         o,
			ForElevatorID: forElevatorID,
			IsComplete:    isComplete,
			IsNew:         isNew,
			IsConfirmed:   false}

		orderWatcher <- watcherUpdate
	}
}

func SendLightsIfChange(change bool, e utils.Elevator, HallOrders *map[int][2][utils.NumFloors]bool,
	ch chan interface{}, LocalLightsCh chan [2][utils.NumFloors]bool) {

	if utils.Master && change {

		fmt.Println("Sending lights from master")
		Lights := LightsToSend(*HallOrders)
		LocalLightsCh <- Lights

		msg := utils.PackMessage("MessageLights", Lights, e.ID)
		ch <- msg

		Printlights(Lights)
	}
}
