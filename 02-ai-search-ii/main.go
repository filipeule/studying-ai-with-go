package main

import (
	"flag"
)

func main() {
	var configFile, algorithm string
	var animate, cat bool

	flag.StringVar(&configFile, "file", "empty.json", "configuration file")
	flag.StringVar(&algorithm, "algorithm", "snake", "cleaning algorithm")
	flag.BoolVar(&animate, "animate", true, "animate while cleaning")
	flag.BoolVar(&cat, "cat", false, "add a cat to the room")
	flag.Parse()

	room := NewRoom(configFile, animate)
	if cat {
		room.Cat = NewCat(room)
	}

	// get a robot
	robot := NewRobot(1, 1)

	// assign a cleaning algorithm
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

	// clean the room
	robot.CleanRoom(room, robot)
}
