package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func CleanRoomRandomWalk(room *Room, robot *Robot) {
	startTime := time.Now()
	moveCount := 0

	// set variables
	maxMoves := room.Width * room.Height * 5
	stuckCount := 0
	maxStuckCount := 5 // max number of consecutive failed moves before changing strategy

	// clean current position
	Clean(robot, room)
	if room.Animate {
		room.Display(robot, false)
		time.Sleep(moveDelay)
	}

	for moveCount < maxMoves && room.CleanedCellCount < room.CleanableCellCount {
		// generate a random angle in radians
		angle := rand.Float64() * 2 * math.Pi

		// calculate a direction vector based on angle
		dx := math.Cos(angle)
		dy := math.Sin(angle)

		// use bresenham's line algorithm to move in that direction until hitting an obstacle
		moves := moveAtAngleUntilObstacle(room, robot, dx, dy)
		moveCount += moves

		// if we didnt move very much, increment stuck count and possibly change strategy
		if moves < 3 {
			stuckCount++

			// if stuck too many time, use A* to find path to nearest dirty cell
			if stuckCount >= maxStuckCount {
				stuckCount = 0
				dirtyCell := findNearestDirtyCell(room, robot.Position)
				if dirtyCell.X != -1 && dirtyCell.Y != -1 {
					path := AStar(room, robot.Position, dirtyCell)
					if len(path) > 1 {
						// move along the path
						for i := 1; i < len(path); i++ {
							robot.Position = path[i]
							robot.Path = append(robot.Path, path[i])
							Clean(robot, room)
							if room.Animate {
								room.Display(robot, false)
								time.Sleep(moveDelay)
							}
							moveCount++
						}
					}
				}
			}
		} else {
			stuckCount = 0
		}

		// add some adaptive behaviour. scan for dirty cells every once in awhile
		if moveCount%20 == 0 {
			if rand.Float64() < 0.3 { // 30% chance to target a specific dirty area
				dirtyCell := findNearestDirtyCell(room, robot.Position)
				if dirtyCell.X != -1 && dirtyCell.Y != -1 {
					path := AStar(room, robot.Position, dirtyCell)
					if len(path) > 1 {
						// move along path
						for i := 1; i < len(path); i++ {
							robot.Position = path[i]
							robot.Path = append(robot.Path, path[i])
							Clean(robot, room)
							if room.Animate {
								room.Display(robot, false)
								time.Sleep(moveDelay)
							}
							moveCount++
						}
					}
				}
			}
		}
	}

	fmt.Println("eita")

	// final sweep to ensure complete coverage
	for i := 1; i < room.Width-1; i++ {
		for j := 1; j < room.Height-1; j++ {
			if !room.Grid[i][j].Cleaned && !room.Grid[i][j].Obstacle {
				path := AStar(room, robot.Position, Point{X: i, Y: j})
				if len(path) == 0 {
					continue
				}

				for k := 1; k < len(path); k++ {
					robot.Position = path[k]
					robot.Path = append(robot.Path, path[k])
					Clean(robot, room)
					if room.Animate {
						room.Display(robot, false)
						time.Sleep(moveDelay)
					}
					moveCount++
				}
			}
		}
	}

	// calculate cleaning time
	cleaningTime := time.Since(startTime)

	displaySummary(room, robot, moveCount, cleaningTime)
}

func moveAtAngleUntilObstacle(room *Room, robot *Robot, dx, dy float64) int {
	moveCount := 0
	maxDistance := math.Max(float64(room.Width), float64(room.Height)) * 2
	startX, startY := robot.Position.X, robot.Position.Y

	endX := startX + int(dx*maxDistance)
	endY := startY + int(dy*maxDistance)

	points := bresenhamLine(startX, startY, endX, endY)

	// move along line until hitting an obstacle
	for i := 1; i < len(points); i++ {
		x, y := points[i].X, points[i].Y

		if !room.IsValid(x, y) {
			break
		}

		// move to new position
		robot.Position = Point{X: x, Y: y}
		robot.Path = append(robot.Path, robot.Position)

		Clean(robot, room)

		// animate if appropriate
		if room.Animate {
			room.Display(robot, false)
			time.Sleep(moveDelay)
		}

		moveCount++
	}

	return moveCount
}

func bresenhamLine(x0, y0, x1, y1 int) []Point {
	// initialize a slice to store all points on the line
	var points []Point

	// calculate the absolute difference between the endpoints
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)

	// determine the direction of movement, along each axis
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}

	// calculate the inicial error value
	err := dx - dy

	// for loop until we reach the endpoint
	for {
		// add the current point to our result
		points = append(points, Point{X: x0, Y: y0})

		// check to see if we've reached the endpoint
		if x0 == x1 && y0 == y1 {
			break
		}

		// calculate the error for the next step
		e2 := 2 * err

		// if moving in the x direction would keep us closer to the ideal line
		if e2 > -dy {
			// if we've reached the endpoint, stop
			if x0 == x1 {
				break
			}
			// update the error and move in the x direction
			err = dy
			x0 += sx
		}

		// if moving in the y direction would keep us closed to the ideal line
		if e2 < dx {
			// if we've reached the endpoint, stop
			if y0 == y1 {
				break
			}
			err += dx
			y0 += sy
		}
	}

	// return poins
	return points
}

func abs(x int) int {
	if x < 0 {
		return x * -1
	}
	return x
}

func findNearestDirtyCell(room *Room, position Point) Point {
	var nearestCell Point = Point{X: -1, Y: -1}
	minDistance := math.MaxFloat64

	for i := 1; i < room.Width-1; i++ {
		for j := 1; j < room.Height-1; j++ {
			distance := heuristic(position, Point{X: i, Y: j})
			if distance < minDistance {
				minDistance = distance
				nearestCell = Point{X: i, Y: j}
			}
		}
	}

	return nearestCell
}
