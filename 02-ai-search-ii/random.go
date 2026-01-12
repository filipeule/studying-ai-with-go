package main

import "time"

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

		// calculate a direction vector based on angle

		// use bresenham's line algorithm to move in that direction until hitting an obstacle

		// if we didnt move very much, increment stuck count and possibly change strategy

		// if stuck too many time, use A* to find path to nearest dirty cell

		// add some adaptive behaviour. scan for dirty cells every once in awhile

		// end for

		// final sweep to ensure complete coverage
	}

	// calculate cleaning time
	cleaningTime := time.Since(startTime)

	displaySummary(room, robot, moveCount, cleaningTime)
}