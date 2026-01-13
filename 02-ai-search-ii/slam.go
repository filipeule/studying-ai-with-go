package main

import (
	"math"
	"time"
)

func CleanRoomSlam(room *Room, robot *Robot) {
	// set start time and move count
	startTime := time.Now()
	moveCount := 0

	// initialize the robots internal map
	robotMap := initializeRobotMap(room.Width, room.Height)

	// initialize visited cells (tracking)
	visited := make(map[Point]bool)

	// initialize a frontier
	frontier := make(map[Point]bool)

	// mark starting position as visited and update map for the first time
	visited[robot.Position] = true
	updateRobotMap(robot.Position, robotMap, room)

	// clean the current position
	Clean(robot, room)

	// add neighbors to the frontier
	addNeighborsToFrontier(robot.Position, robotMap, frontier, visited, room)

	// display the initial state
	if room.Animate {
		room.Display(robot, false)
		time.Sleep(moveDelay)
	}

	// for - if the frontier is not empty and the room is not 100% clean
	for len(frontier) > 0 && room.CleanedCellCount < room.CleanableCellCount {
		// get closest frontier point
		target := getClosestFrontierPoint(robot.Position, frontier)

		// if no valid target, break
		if target.X == -1 && target.Y == -1 {
			break
		}

		// remove target from frontier
		delete(frontier, target)

		// find path to target using astar
		path := AStar(room, robot.Position, target)

		// if no path found, go to next frontier point
		if len(path) <= 1 {
			continue
		}

		// move along the path
		for i := 1; i < len(path); i++ {
			// update robot position
			robot.Position = path[i]
			robot.Path = append(robot.Path, robot.Position)

			// clean position
			Clean(robot, room)

			// mark as visited
			visited[robot.Position] = true

			// update map (internal) based upon what we can see from the current position
			updateRobotMap(robot.Position, robotMap, room)

			// update frontier with newly discovered cells
			addNeighborsToFrontier(robot.Position, robotMap, frontier, visited, room)

			// display the room
			if room.Animate {
				room.Display(robot, false)
				time.Sleep(moveDelay)
			}

			moveCount++
		}

		// every 10 moves, do a more thorough frontier check
		if moveCount%10 == 0 {
			updateAllFrontiers(robotMap, frontier, visited, room)
		}

		// chech if we have sufficient coverage - break
		if float64(room.CleanedCellCount) / float64(room.CleanableCellCount) > 0.95 {
			break
		}
	}

	// final cleanup phase
	cleanRemainingCells(room, robot, &moveCount)

	// calculate cleaning time
	cleaningTime := time.Since(startTime)

	// display final statistics
	displaySummary(room, robot, moveCount, cleaningTime)
}

func initializeRobotMap(width, height int) [][]int {
	// 0 = unknown, 1 = free, 2 = obstacle, 3 = cleaned
	robotMap := make([][]int, width)
	for i := range robotMap {
		robotMap[i] = make([]int, height)
	}

	return robotMap
}

func updateRobotMap(position Point, robotMap [][]int, room *Room) {
	if room.Grid[position.X][position.Y].Cleaned {
		robotMap[position.X][position.Y] = 3
	} else {
		robotMap[position.X][position.Y] = 1
	}

	// scan surroundings
	for _, dir := range directions {
		newX, newY := position.X+dir[0], position.Y+dir[1]

		// check if position is within bounds
		if newX >= 0 && newX < len(robotMap) && newY >= 0 && newY < len(robotMap[0]) {
			if room.Grid[newX][newY].Obstacle {
				robotMap[newX][newY] = 2
			} else if robotMap[newX][newY] == 0 {
				robotMap[newX][newY] = 1
			} else if room.Grid[newX][newY].Cleaned {
				robotMap[newX][newY] = 3
			}
		}
	}
}

func addNeighborsToFrontier(
	position Point,
	robotMap [][]int,
	frontier map[Point]bool,
	visited map[Point]bool,
	room *Room,
) {
	// chech adjacent cells
	for _, dir := range directions {
		newX, newY := position.X+dir[0], position.Y+dir[1]
		newPoint := Point{X: newX, Y: newY}

		// check if position is valid, not visited, not an obstacle and not already in frontier
		if newX >= 0 && newX <= len(robotMap) && newY >= 0 && newY < len(robotMap[0]) &&
			!visited[newPoint] && !frontier[newPoint] && room.IsValid(newX, newY) {
			// add to frontier
			frontier[newPoint] = true
		}
	}
}

func getClosestFrontierPoint(position Point, frontier map[Point]bool) Point {
	closestPoint := Point{X: -1, Y: -1}
	minDistance := math.MaxFloat64

	for point := range frontier {
		distance := heuristic(position, point)
		if distance < minDistance {
			minDistance = distance
			closestPoint = point
		}
	}

	return closestPoint
}

func updateAllFrontiers(robotMap [][]int, frontier map[Point]bool, visited map[Point]bool, room *Room) {
	for x := 1; x < room.Width-1; x++ {
		for y := 1; y < room.Height-1; y++ {
			// if a cell is free, but not visited, add to frontier
			point := Point{X: x, Y: y}
			if robotMap[x][y] == 1 && !visited[point] && !frontier[point] && !room.Grid[x][y].Obstacle {
				// check to see if it is accesible (has at least 1 visited neighbor)
				for _, dir := range directions {
					nx, ny := x+dir[0], y+dir[1]
					neighborPoint := Point{X: nx, Y: ny}
					if nx >= 0 && nx < room.Width && ny < room.Height && ny >= 0 && visited[neighborPoint] {
						frontier[point] = true
						break
					}
				}
			}
		}
	}
}

func cleanRemainingCells(room *Room, robot *Robot, moveCount *int) {
	// find all cells that should be cleanable but havent been cleaned
	for i := 1; i < room.Width-1; i++ {
		for j := 1; j < room.Height-1; j++ {
			// if the cell is not an obstacle, not cleaned, and known to the robot
			if !room.Grid[i][j].Obstacle && !room.Grid[i][j].Cleaned {
				path := AStar(room, robot.Position, Point{X: i, Y: j})
				if len(path) <= 1 {
					continue
				}

				// move along the path
				for k := 1; k < len(path); k++ {
					// update robot position
					robot.Position = path[k]
					robot.Path = append(robot.Path, path[k])

					// clean position
					Clean(robot, room)

					// display room
					if room.Animate {
						room.Display(robot, false)
						time.Sleep(moveDelay)
					}

					*moveCount++
				}
			}
		}
	}
}