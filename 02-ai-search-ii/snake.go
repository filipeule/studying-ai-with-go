package main

import "time"

func CleanRoomSnake(room *Room, robot *Robot) {
	// initialize start time and movecount
	startTime := time.Now()
	moveCount := 0

	// generate the snaking pattern
	coveragePoints := generateSnakingPattern(room)

	// clean cell
	Clean(robot, room)

	if room.Animate {
		room.Display(robot, room.Cat, false)
		time.Sleep(moveDelay)
	}

	// visit each point in the coverage pattern (for loop)
	for _, point := range coveragePoints {
		// move the cat
		MoveCat(room.Cat, room)

		// skip cell if already clean
		if room.Grid[point.X][point.Y].Cleaned {
			continue
		}

		// find path to the next point
		path := AStar(room, robot.Position, point)

		// if no path found, try the next point
		if len(path) == 0 {
			continue
		}

		// move along the path (for loop)
		for i := 1; i < len(path); i++ {
			// update robot position
			robot.Position = path[i]
			robot.Path = append(robot.Path, path[i])

			// clean
			Clean(robot, room)
			MoveCat(room.Cat, room)

			// display the room
			if room.Animate {
				room.Display(robot, room.Cat, false)
				time.Sleep(moveDelay)
			}

			// increment movecount
			moveCount++
		}
	}

	// do final sweep
	finalCleanup(room, robot, &moveCount)

	cleaningTime := time.Since(startTime)

	displaySummary(room, robot, moveCount, cleaningTime)
}

func generateSnakingPattern(room *Room) []Point {
	var points []Point
	var directionX = 1

	for y := 1; y < room.Height-1; y++ {
		if directionX == 1 {
			// moving left to right
			for x := 1; x < room.Width-1; x++ {
				if !room.Grid[x][y].Obstacle {
					points = append(points, Point{X: x, Y: y})
				}
			}
		} else {
			// move right to left
			for x := room.Width - 2; x >= 1; x-- {
				if !room.Grid[x][y].Obstacle {
					points = append(points, Point{X: x, Y: y})
				}
			}
		}

		directionX *= -1
	}

	return points
}
