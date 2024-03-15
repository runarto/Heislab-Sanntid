package crash

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

func Crash(e utils.Elevator) {
	fmt.Println("Elevator", e.ID, "crashed")
	SaveCabOrders(e)
	os.Exit(1)

}

func SaveCabOrders(e utils.Elevator) {

	var CabOrders [utils.NumFloors]bool

	for f := 0; f < utils.NumFloors; f++ {
		CabOrders[f] = e.LocalOrderArray[utils.Cab][f]
	}

	// Convert the cab orders to JSON
	data, err := json.Marshal(CabOrders)
	if err != nil {
		fmt.Println("Error marshaling cab orders:", err)
		return
	}

	// Create the file
	file, err := os.Create("crash/CabOrders.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

}

func CheckCrashDump() utils.Elevator {
	fmt.Println("Checking for crash dump")
	var CabOrders [utils.NumFloors]bool
	var orders [utils.NumButtons][utils.NumFloors]bool
	nullOrders := [utils.NumButtons][utils.NumFloors]bool{}

	// Open the file
	file, err := os.Open("crash/CabOrders.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return utils.Elevator{
			CurrentState:     utils.Still,
			CurrentFloor:     elevio.GetFloor(),
			CurrentDirection: elevio.MD_Stop,
			LocalOrderArray:  nullOrders,
			ID:               utils.ID,
			IsActive:         true,
		}
	}
	defer file.Close()

	// Read the file
	data := make([]byte, 100)
	count, err := file.Read(data)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return utils.Elevator{
			CurrentState:     utils.Still,
			CurrentFloor:     elevio.GetFloor(),
			CurrentDirection: elevio.MD_Stop,
			LocalOrderArray:  nullOrders,
			ID:               utils.ID,
			IsActive:         true,
		}
	}

	// Unmarshal the JSON data
	err = json.Unmarshal(data[:count], &CabOrders)
	if err != nil {
		fmt.Println("Error unmarshaling data:", err)
		return utils.Elevator{
			CurrentState:     utils.Still,
			CurrentFloor:     elevio.GetFloor(),
			CurrentDirection: elevio.MD_Stop,
			LocalOrderArray:  nullOrders,
			ID:               utils.ID,
			IsActive:         true,
		}
	}

	for f := 0; f < utils.NumFloors; f++ {
		orders[utils.Cab][f] = CabOrders[f]
	}

	fmt.Println("Crash dump found and loaded")
	return utils.Elevator{
		CurrentState:     utils.Still,
		CurrentFloor:     elevio.GetFloor(),
		CurrentDirection: elevio.MD_Stop,
		LocalOrderArray:  orders,
		ID:               utils.ID,
		IsActive:         true,
	}
}
