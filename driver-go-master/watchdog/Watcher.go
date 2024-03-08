package watchdog

import (
	"fmt"
	"time"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/orders"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func MasterBark(e utils.Elevator, c *utils.Channels, m *utils.OrderWatcherArray) {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if utils.Master {

			currentTime := time.Now()

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

						c.MasterBarkCh <- order
					}
				}
			}
		}
	}
}

func SlaveBark(e utils.Elevator, c *utils.Channels, s *utils.OrderWatcherArray) {

	fmt.Println("Barker started.")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if utils.Master {

			currentTime := time.Now()

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

						c.SlaveBarkCh <- order
					}
				}
			}
		}
	}
}

func Watchdog(c *utils.Channels, e utils.Elevator, m *utils.OrderWatcherArray, s *utils.OrderWatcherArray) {

	fmt.Println("Watchdog started.")

	go MasterBark(e, c, m)

	go SlaveBark(e, c, m)

	for {

		select {

		case order := <-c.MasterBarkCh:

			fmt.Println("Master bark received, resending order", order)

			BestElevator := orders.ChooseElevator(order)

			if BestElevator.ID == e.ID {

				c.ButtonCh <- elevio.ButtonEvent{
					Floor:  order.Floor,
					Button: order.Button,
				}

			} else {

				utils.CreateAndSendMessage(c, "MessageNewOrder", order, BestElevator.ID, e.ID)

			}

		case order := <-c.SlaveBarkCh:

			fmt.Println("Slave bark received, resending order to master", order)

			utils.CreateAndSendMessage(c, "MessageNewOrder", order, utils.MasterID, e.ID)
		}
	}
}
