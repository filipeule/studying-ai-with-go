package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	var configFile, algorithm string
	var animate, cat, isHouse, useLogic bool

	flag.StringVar(&configFile, "file", "empty.json", "configuration file")
	flag.StringVar(&algorithm, "algorithm", "snake", "cleaning algorithm")
	flag.BoolVar(&animate, "animate", true, "animate while cleaning")
	flag.BoolVar(&cat, "cat", false, "add a cat to the room")
	flag.BoolVar(&isHouse, "house", false, "config file has multiple rooms")
	flag.BoolVar(&useLogic, "logic", false, "use propositional logic for cleaning decisions")
	flag.Parse()

	var house *House

	if !isHouse {
		// if not a house, we just have one room. create a house and assing one room to it
		// this way we can use the same loop for houses and for individual rooms
		var rooms []*Room

		// get a room from json config
		room := NewRoom(configFile, animate)
		rooms = append(rooms, room)

		var h House
		h.Rooms = rooms
		house = &h
	} else {
		// we are doing a complete house. just get a house from json config
		house = NewHouse(configFile, animate)
	}

	// add cats to rooms if necessary
	if cat {
		for _, room := range house.Rooms {
			room.Cat = NewCat(room)
		}
	}

	roomCount := 0

	if useLogic {
		// use propositional logic for cleaning
		fmt.Println("using propositional logic for cleaning decisions")
		robot := NewRobotWithLogic(1, 1)

		// assign a cleaning algorithm
		setUpAlgorithm(algorithm, robot.Robot)

		// scan the house
		roomNameToIndex := robot.ScanHouseWithLogic(house)

		fmt.Println("\nlogical state after scanning:")
		fmt.Printf("today is %s (weekday: %t)\n", time.Now().Weekday(), robot.World.IsWeekday)
		fmt.Printf("jack is home: %t\n", robot.World.Jack.IsHome)
		fmt.Printf("sarah is home: %t\n", robot.World.Sarah.IsHome)
		fmt.Printf("johnny is home: %t\n", robot.World.Johnny.IsHome)
		fmt.Printf("johnny's door is closed: %t\n", robot.World.Johnny.DoorClosed)

		if robot.World.Johnny.DoorClosed {
			fmt.Println("logic: will not vacuum johnny's room")
		} else {
			fmt.Println("logic: will vacuum johnny's room")
		}

		// determine cleaning priority based on logical rules
		cleaningPriority := robot.World.DetermineCleaningPriority()
		fmt.Println("\ndetermined cleaning priority based on propositional logic:")
		for i, roomName := range cleaningPriority {
			fmt.Printf("%d. %s\n", i+1, roomName)
		}

		for k, v := range roomNameToIndex {
			fmt.Println(k, "->", v)
		}

		fmt.Println("\npress enter to start cleaning...")
		fmt.Scanln()

		// clean the rooms in priority order
		for _, roomName := range cleaningPriority {
			// check to see if room exists
			roomIndex, exists := roomNameToIndex[roomName]
			if !exists {
				fmt.Printf("room '%s' not found in the house, skipping\n", roomName)
				continue
			}

			// get the room from house.Rooms by index
			room := house.Rooms[roomIndex]

			// reset robot position to 1,1
			robot.Position = Point{X: 1, Y: 1}
			robot.Path = []Point{{X: 1, Y: 1}}

			// clean the room
			robot.CleanRoom(room, robot.Robot)
			roomCount++
		}

	} else {
		// use the original cleaning approach without propositional logic, and for multiple rooms
		for _, room := range house.Rooms {
			// get a robot
			robot := NewRobot(1, 1)

			// assign a cleaning algorithm
			setUpAlgorithm(algorithm, robot)

			// clean the room
			robot.CleanRoom(room, robot)
			roomCount++
		}
	}

	fmt.Println()
	fmt.Printf("all done. cleaned a total of %d room(s)\n", roomCount)
}

func setUpAlgorithm(algorithm string, robot *Robot) {
	switch algorithm {
	case "random":
		robot.CleanRoom = CleanRoomRandomWalk
	case "slam":
		robot.CleanRoom = CleanRoomSlam
	case "spiral":
		robot.CleanRoom = CleanSpiralPattern
	case "snake":
		robot.CleanRoom = CleanRoomSnake
	default:
		robot.CleanRoom = CleanRoomSnake
	}
}
