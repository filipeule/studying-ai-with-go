package main

import (
	"fmt"
	"time"
)

type PersonStatus struct {
	Name       string
	IsHome     bool
	Room       string // person's room name
	DoorClosed bool
}

type LogicalWorld struct {
	Jack      PersonStatus
	Sarah     PersonStatus
	Johnny    PersonStatus
	IsWeekday bool
	Objects   map[string]bool
}

func NewLogicalWorld() *LogicalWorld {
	// get current day to determine if it's a weekday
	today := time.Now()
	weekday := today.Weekday()

	isWeekday := weekday >= time.Monday && weekday <= time.Friday

	return &LogicalWorld{
		Jack: PersonStatus{
			Name:   "Jack",
			IsHome: false,
			Room:   "Jack's Room",
		},
		Sarah: PersonStatus{
			Name:   "Sarah",
			IsHome: false,
			Room:   "Sarah's Room",
		},
		Johnny: PersonStatus{
			Name:       "Johnny",
			IsHome:     false,
			Room:       "Johnny's Room",
			DoorClosed: false,
		},
		IsWeekday: isWeekday,
		Objects:   make(map[string]bool),
	}
}

func (world *LogicalWorld) UpdateObjectFound(objectName string) {
	world.Objects[objectName] = true

	// apply our object to person identification rules

	// rule 1: if a backpack is found, then jack is home
	if objectName == "backpack" {
		if !world.Jack.IsHome {
			fmt.Println("logic: backpack found, deducing jack is home")
		}
		world.Jack.IsHome = true
	}

	// rule 2: if a bicycle is found, then sarah is home
	if objectName == "bicycle" {
		if !world.Sarah.IsHome {
			fmt.Println("logic: bicycle found, deducing sarah is home")
		}
		world.Sarah.IsHome = true
	}

	// rule 3: if a skateboard is found, then johnny is home
	if objectName == "skateboard" {
		if !world.Johnny.IsHome {
			fmt.Println("logic: skateboard found, deducing johnny is home")
		}
		world.Johnny.IsHome = true
	}
}

// UpdateDoorStatus updates whether johnny' door is open or closed
func (world *LogicalWorld) UpdateDoorStatus(doorName string, isClosed bool) {
	if doorName == "Johnny's Door" {
		world.Johnny.DoorClosed = isClosed
		fmt.Printf("logic: johnny's door is now %s\n", map[bool]string{true: "closed", false: "open"}[isClosed])
	}
}

// DetermineCleaningPriority decide on the order we clean rooms
func (world *LogicalWorld) DetermineCleaningPriority() []string {
	availableRooms := []string{
		"Kitchen",
		"Living Room",
		"Jack's Room",
		"Sarah's Room",
		"Johnny's Room",
	}

	// rule: if no one is home, then vacuum all rooms starting with the kitchen
	if !world.Jack.IsHome && !world.Sarah.IsHome && !world.Johnny.IsHome {
		fmt.Println("logic: no one is home, so vacuuming all rooms starting with the kitchen")
		return availableRooms
	}

	// initialize a priority list with all available rooms
	priorityList := make([]string, 0)
	skipRooms := make(map[string]bool)

	// rule: if sarah is home, then don't vacuum the living room
	if world.Sarah.IsHome {
		fmt.Println("logic: sarah is home, skipping the living room")
		skipRooms["Living Room"] = true
	}

	// rule: if johnny is home and his door is closed, then skip johnny's room
	if world.Johnny.IsHome && world.Johnny.DoorClosed {
		fmt.Println("logic: johnny is home and his door is closed, skipping his room")
		skipRooms["Johnny's Room"] = true
	}

	// rule: if jack is home and it's a weekday, do his room last
	jackRoomLast := world.Jack.IsHome && world.IsWeekday

	// build our priority list. add kitchen first and filter out skipped rooms
	priorityList = append(priorityList, "Kitchen")

	// add all other room's except Jack's (if it needs to be last) and skipped rooms
	for _, room := range availableRooms {
		if room == "Kitchen" {
			continue // already added
		}

		if room == "Jack's Room" && jackRoomLast {
			continue // will be added last
		}

		if skipRooms[room] {
			continue // skip this room
		}

		priorityList = append(priorityList, room)
	}

	// add jack's room if it should be last
	if jackRoomLast && !skipRooms["Jack's Room"] {
		priorityList = append(priorityList, "Jack's Room")
	}

	return priorityList
}

// RobotWithLogic is a type for a robot with logic
type RobotWithLogic struct {
	*Robot // embedding the original robot
	World  *LogicalWorld
}

// NewRobotWithLogic is a factory method for robot with logic
func NewRobotWithLogic(startX, startY int) *RobotWithLogic {
	return &RobotWithLogic{
		Robot: NewRobot(startX, startY),
		World: NewLogicalWorld(),
	}
}

func (robot *RobotWithLogic) ScanHouseWithLogic(house *House) map[string]int {
	// create a map which maps room indices to room names
	roomNameToIndex := make(map[string]int)

	// identify all rooms and generate our mapping
	for i, room := range house.Rooms {
		roomName := ""

		for x := range room.Width {
			for y := range room.Height {
				if room.Grid[x][y].Type == "furniture" {
					if roomName == "" {
						switch room.Grid[x][y].ObstacleName {
						case "bed":
							if roomName == "" && roomNameToIndex["Jack's Room"] == 0 {
								roomName = "Jack's Room"
							} else if roomNameToIndex["Sarah's Room"] == 0 && roomNameToIndex["Johnny's Room"] == 0 {
								roomName = "Sarah's Room"
							} else {
								roomName = "Johnny's Room"
							}
						case "desk":
							if roomName == "" {
								roomName = "Study"
							}
						case "sofa", "tv":
							roomName = "Living Room"
						case "stove", "fridge", "sink":
							roomName = "Kitchen"
						}

						// johnny's door
						if room.Grid[x][y].ObstacleName == "johnny's door" {
							// change to true to close door
							robot.World.UpdateDoorStatus("johnny's door", false)
						}
					}
				}
			}
		}

		// if we can't determine the name of the room, give it a default name
		if roomName == "" {
			roomName = fmt.Sprintf("Room %d", i)
		}

		roomNameToIndex[roomName] = i
		fmt.Printf("identified room %s (index %d)\n", roomName, i)
	}

	// scan the house for objects to build our logical world
	fmt.Println("robot is scanning the house for objects...")

	for _, room := range house.Rooms {
		for x := range room.Width {
			for y := range room.Height {
				if room.Grid[x][y].Type == "furniture" && room.Grid[x][y].ObstacleName != "" {
					robot.World.UpdateObjectFound(room.Grid[x][y].ObstacleName)
				}
			}
		}
	}

	return roomNameToIndex
}
