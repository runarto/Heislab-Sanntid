package watchdog

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func MasterBark(e utils.Elevator, m *utils.OrderWatcherArray, MasterBarkCh chan utils.Order) {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if utils.Master {

			currentTime := time.Now()
			m.WatcherMutex.Lock()

			for button := 0; button < utils.NumButtons-1; button++ {
				for floor := 0; floor < utils.NumFloors; floor++ {

					timeSent := m.HallOrderArray[button][floor].Time

					if currentTime.Sub(timeSent) > utils.MasterTimeout &&
						!m.HallOrderArray[button][floor].Completed &&
						m.HallOrderArray[button][floor].Active {

						m.HallOrderArray[button][floor].Time = time.Now()

						// Resend the order to the network
						order := utils.Order{
							Floor:  floor,
							Button: elevio.ButtonType(button),
						}

						MasterBarkCh <- order
					}
				}
			}
			m.WatcherMutex.Unlock()
		}
	}
}

func SlaveBark(e utils.Elevator, s *utils.OrderWatcherArray, SlaveBarkCh chan utils.Order) {

	fmt.Println("Barker started.")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if !utils.Master {

			currentTime := time.Now()
			s.WatcherMutex.Lock()

			for button := 0; button < utils.NumButtons-1; button++ {
				for floor := 0; floor < utils.NumFloors; floor++ {

					timeSent := s.HallOrderArray[button][floor].Time

					if currentTime.Sub(timeSent) > utils.SlaveTimeout &&
						s.HallOrderArray[button][floor].Active &&
						!s.HallOrderArray[button][floor].Confirmed {

						s.HallOrderArray[button][floor].Time = time.Now()

						order := utils.Order{
							Floor:  floor,
							Button: elevio.ButtonType(button),
						}

						SlaveBarkCh <- order
					}
				}
			}
			s.WatcherMutex.Unlock()
		}
	}
}

func Watchdog(e utils.Elevator, m *utils.OrderWatcherArray, s *utils.OrderWatcherArray, MasterBarkCh chan utils.Order,
	SlaveBarkCh chan utils.Order, ButtonCh chan elevio.ButtonEvent, ch chan interface{}) {

	fmt.Println("Watchdog started.")

	go MasterBark(e, m, MasterBarkCh)

	//go SlaveBark(e, s, SlaveBarkCh)

	for {

		select {

		case order := <-MasterBarkCh:

			fmt.Println("Master bark received, resending order", order)

			BestElevator := orders.ChooseElevator(order)

			if BestElevator.ID == e.ID {

				ButtonCh <- elevio.ButtonEvent{
					Floor:  order.Floor,
					Button: order.Button,
				}

			} else {

				msg := utils.PackMessage("MessageNewOrder", order, BestElevator.ID, e.ID)
				ch <- msg

			}

		case order := <-SlaveBarkCh:

			fmt.Println("Slave bark received, resending order to master", order)

			msg := utils.PackMessage("MessageNewOrder", order, utils.MasterID, e.ID)
			ch <- msg

		}
	}
}
