package main

import "time"

func CleanSpiralPattern(room *Room, robot *Robot) {
	startTime := time.Now()
	moveCount := 0

	// find the center of the room
	centerX := room.Width / 2
	centerY := room.Height / 2

	// find a valid point near the center
	centerPoint := findNearestCleanablePoint(room, Point{X: centerX, Y: centerY})

	// find path to center (A*)
	pathToCenter := AStar(room, robot.Position, centerPoint)

	// move to the center point
	if len(pathToCenter) > 1 {
		for i := 1; i < len(pathToCenter); i++ {
			robot.Position = pathToCenter[i]
			robot.Path = append(robot.Path, robot.Position)
			Clean(robot, room)

			if room.Animate {
				room.Display(robot, false)
				time.Sleep(moveDelay)
			}

			moveCount++
		}
	}

	// create a spiral pattern
	spiralPoints := generateSpiralPattern(room, centerPoint)

	// follow the spiral pattern (for loop)
	for _, point := range spiralPoints {
		// skip if cell is already clean, or an obstacle
		if room.Grid[point.X][point.Y].Cleaned || room.Grid[point.X][point.Y].Obstacle {
			continue
		}

		// find path to the next point (A*)
		path := AStar(room, robot.Position, point)

		if len(path) <= 1 {
			continue
		}

		// move along path
		for i := 1; i < len(path); i++ {
			robot.Position = path[i]
			robot.Path = append(robot.Path, robot.Position)
			Clean(robot, room)

			if room.Animate {
				room.Display(robot, false)
				time.Sleep(moveDelay)
			}
			moveCount++
		}
	}

	// final cleanup
	finalCleanup(room, robot, &moveCount)

	// calculate cleaning time
	cleaningTime := time.Since(startTime)

	// display final statistics
	displaySummary(room, robot, moveCount, cleaningTime)
}

func findNearestCleanablePoint(room *Room, target Point) Point {
	if room.IsValid(target.X, target.Y) && !room.Grid[target.X][target.Y].Obstacle {
		return target
	}

	// search for a valid point in expanding circles
	for radius := 1; radius < room.Width || radius < room.Height; radius++ {
		// check all points at the current radius
		for dx := -radius; dx <= radius; dx ++ {
			for dy := -radius; dy <= radius; dy++ {
				if abs(dx) != radius && abs(dy) != radius {
					continue
				}

				x, y := target.X + dx, target.Y + dy

				// check to see if this point is valid and not a obstacle
				if room.IsValid(x, y) && !room.Grid[x][y].Obstacle {
					return Point{X: x, Y: y}
				}
			}
		}
	}

	// if no valid point, return the starting point
	return Point{X: 1, Y: 1}
}

func generateSpiralPattern(room *Room, center Point) []Point {
	var points []Point

	// maximum possible spiral size
	maxSize := max(room.Width, room.Height)

	// set delta x and delta y
	dx := []int{1, 0, -1, 0}
	dy := []int{0, 1, 0, -1}

	// start at center
	x, y := center.X, center.Y
	dir := 0 // start moving right

	// set spiral parameters
	step := 1
	stepCount := 0
	dirChanges := 0

	// generate the spiral pattern
	for range maxSize * maxSize {
		// add current point if valid
		if room.IsValid(x, y) {
			points = append(points, Point{X: x, Y: y})
		}

		// take a step
		x += dx[dir]
		y += dy[dir]
		stepCount++

		// check to see if we need to change direction
		if stepCount == step {
			dir = (dir + 1) % 4
			stepCount = 0
			dirChanges++

			// increase step size after every two direction changes
			if dirChanges == 2 {
				step++
				dirChanges = 0
			}
		}

		// break if we are out of bounds
		if x < 0 || x >= room.Width || y < 0 || y >= room.Height {
			break
		}
	}

	return points
}

func finalCleanup(room *Room, robot *Robot, moveCount *int) {
	for i := 1; i < room.Width - 1; i++ {
		for j := 1; j < room.Height - 1; j++ {
			if !room.Grid[i][j].Obstacle && !room.Grid[i][j].Cleaned {
				// find path to cell
				path := AStar(room, robot.Position, Point{X: i, Y: j})

				if len(path) <= 1 {
					continue
				}

				// move along path
				for k := 1; k < len(path); k++ {
					robot.Position = path[k]
					robot.Path = append(robot.Path, robot.Position)
					Clean(robot, room)

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