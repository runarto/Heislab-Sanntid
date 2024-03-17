package watchdog

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Checks if a hall order has been confirmed by slave within a certain timeout period and sends a bark signal if not}
//*
//* @param      m             The master order watcher array
//* @param      MasterBarkCh  Channel that resends order for re-delegation
// */

func MasterBark(m *utils.OrderWatcherArray, MasterBarkCh chan utils.Order) {

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

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Checks if a hall order has been confirmed by master within a certain timeout period and sends a bark signal if not}
//*
//* @param      s             The slave order watcher array
//* @param      SlaveBarkCh   Channel that resends order to master
// */

func SlaveBark(s *utils.OrderWatcherArray, SlaveBarkCh chan utils.Order) {

	fmt.Println("Barker started.")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if !utils.Master && utils.MasterID != utils.NotDefined {

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

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Watches for barks from master and slave and resends the order to the network}
//*
//* @param      e             The elevator
//* @param      m             The master order watcher array
//* @param      s             The slave order watcher array
//* @param      MasterBarkCh  Channel that resends order for re-delegation
//* @param      SlaveBarkCh   Channel that resends order to master
//* @param      messageHandler  Channel for sending orders to the network
// */

func Watchdog(e utils.Elevator, m *utils.OrderWatcherArray, s *utils.OrderWatcherArray, MasterBarkCh chan utils.Order,
	SlaveBarkCh chan utils.Order, messageHandler chan utils.Message) {

	fmt.Println("Watchdog started.")

	go MasterBark(m, MasterBarkCh)

	go SlaveBark(s, SlaveBarkCh)

	for {

		select {

		case order := <-MasterBarkCh:

			fmt.Println("Master bark received, resending order", order)

			BestElevator := utils.ChooseElevator(order)

			msg := utils.PackMessage("MessageNewOrder", BestElevator.ID, e.ID, order)
			messageHandler <- msg

		case order := <-SlaveBarkCh:

			fmt.Println("Slave bark received, resending order to master", order)

			msg := utils.PackMessage("MessageNewOrder", utils.MasterID, e.ID, order)
			messageHandler <- msg

		}
	}
}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
