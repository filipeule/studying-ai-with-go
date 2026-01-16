package main

import "math/rand"

type Cat struct {
	Position   Point
	Active     bool
	StopTimer  int
	Path       []Point
	DirectionX int
	DirectionY int
}

func NewCat(room *Room) *Cat {
	var startX, startY int
	for {
		startX = rand.Intn(room.Width-4) +2
		startY = rand.Intn(room.Height-4) + 2

		if !room.Grid[startX][startY].Obstacle {
			break
		}
	}

	return &Cat{
		Position: Point{X: startX, Y: startY},
		Active: true,
		StopTimer: 0,
		Path: []Point{{X: startX, Y: startY}},
		DirectionX: rand.Intn(3) - 1,
		DirectionY: rand.Intn(3) - 1,
	}
}

func MoveCat(cat *Cat, room *Room) {
	if cat != nil {
		// check to see if the cat is currently stopped
		if cat.StopTimer > 0 {
			cat.StopTimer--
		}

		// change for the cat to stop
		if rand.Float64() < catStopProbability {
			cat.StopTimer = catStopDuration
			return
		}

		// change direction randomly
		if rand.Float64() < 0.2 {
			cat.DirectionX = rand.Intn(3) - 1
			cat.DirectionY = rand.Intn(3) - 1
		}

		// make sure cat doesnt just stay still when moving
		if cat.DirectionX == 0 && cat.DirectionY == 0 {
			cat.DirectionX = rand.Intn(3) - 1
			if cat.DirectionX == 0 {
				cat.DirectionY = rand.Intn(2) * 2 - 1
			} else {
				cat.DirectionY = rand.Intn(3) - 1
			}
		}

		// calculate the new position
		newX := cat.Position.X + cat.DirectionX
		newY := cat.Position.Y + cat.DirectionY

		if room.IsValid(newX, newY) {
			// update cat position
			cat.Position = Point{X: newX, Y: newY}
			cat.Path = append(cat.Path, cat.Position)
		}

		if room.Grid[newX][newY].Cleaned {
			room.Grid[newX][newY].Cleaned = false
			room.Grid[newX][newY].Type = "dirty"
			room.CleanedCellCount--
		}
	}
}

func IsAdjacentToCat(robot *Robot, cat *Cat) bool {
	if cat != nil {
		dx := abs(robot.Position.X - cat.Position.X)
		dy := abs(robot.Position.Y - cat.Position.Y)
		return dx <= 1 && dy <= 1
	}
	return false
}