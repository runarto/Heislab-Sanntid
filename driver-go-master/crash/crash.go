package crash

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/runarto/Heislab-Sanntid/elevio"
	"github.com/runarto/Heislab-Sanntid/utils"
)

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Saves the cab orders to a file}
//*
//* @param      e     {The elevator}
// */

func Crash(e utils.Elevator) {
	fmt.Println("Elevator", e.ID, "crashed")
	SaveCabOrders(e)
	os.Exit(1)

}

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Saves the cab orders to a file}
//*
//* @param      e     {The elevator}
// */

func SaveCabOrders(e utils.Elevator) {

	var CabOrders [utils.NumButtons][utils.NumFloors]bool

	for b := 2; b < utils.NumButtons; b++ {
		for f := 0; f < utils.NumFloors; f++ {
			CabOrders[b][f] = e.LocalOrderArray[b][f]
		}
	}

	// Convert the cab orders to JSON
	data, err := json.Marshal(CabOrders)
	if err != nil {
		fmt.Println("Error marshaling cab orders:", err)
		return
	}

	// Create the file
	fileString := "crash/CabOrders" + strconv.Itoa(utils.ID) + ".json"
	file, err := os.Create(fileString)
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

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//*
//* @brief      {Checks for a crash dump and loads elevator state from it if it exists}
//*
//* @return     {The elevator}
// */

func CheckCrashDump() utils.Elevator {
	fmt.Println("Checking for crash dump")
	var orders [utils.NumButtons][utils.NumFloors]bool
	nullOrders := [utils.NumButtons][utils.NumFloors]bool{}

	// Open the file
	fileString := "crash/CabOrders" + strconv.Itoa(utils.ID) + ".json"
	file, err := os.Open(fileString)
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
	err = json.Unmarshal(data[:count], &orders)
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

//*--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
