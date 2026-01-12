package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	// Display characters
	charRobot          = "🔴"
	charWall           = "🟦"
	charFurniture      = "🪑"
	charClean          = "🧼"
	charDirty          = "🟫"
	charPath           = "🟢"
	charCat            = "🐱" // Display character for cat
	catStopProbability = 0.1 // Probability of cat stopping
	catStopDuration    = 5   // Duration cat stays still (in animation frames)
	moveDelay          = 50 * time.Millisecond
	cellSize           = 10
)

type Point struct {
	X, Y int
}

type Cell struct {
	Type         string // wall, furniturem clean, dirty, bike
	Cleaned      bool
	Obstacle     bool
	ObstacleName string
}

type Furniture struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

type RoomConfig struct {
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	Furniture []Furniture `json:"furniture"`
}

type Room struct {
	Grid               [][]Cell
	Width              int
	Height             int
	CleanableCellCount int
	CleanedCellCount   int
	Animate            bool
}

func NewRoom(configFile string, animate bool) *Room {
	// load from json config
	roomConfig, err := LoadRoomConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// convert dimensions of the room to grid cells
	gridWidth := roomConfig.Width / cellSize
	gridHeight := roomConfig.Height / cellSize

	// create grid
	grid := make([][]Cell, gridWidth)
	for i := range grid {
		grid[i] = make([]Cell, gridHeight)
		for j := range grid[i] {
			grid[i][j] = Cell{Type: "dirty", Cleaned: false, Obstacle: false}
		}
	}

	// add walls
	for i := 0; i < gridWidth; i++ {
		grid[i][0] = Cell{Type: "wall", Cleaned: false, Obstacle: true, ObstacleName: "wall"}
		grid[i][gridHeight-1] = Cell{Type: "wall", Cleaned: false, Obstacle: true, ObstacleName: "wall"}
	}

	for j := 0; j < gridHeight; j++ {
		grid[0][j] = Cell{Type: "wall", Cleaned: false, Obstacle: true, ObstacleName: "wall"}
		grid[gridWidth-1][j] = Cell{Type: "wall", Cleaned: false, Obstacle: true, ObstacleName: "wall"}
	}

	// add furniture
	for _, f := range roomConfig.Furniture {
		x := f.X / cellSize
		y := f.Y / cellSize

		width := f.Width / cellSize
		height := f.Height / cellSize

		for i := x; i < x+width; i++ {
			for j := y; j < y+height; j++ {
				grid[i][j] = Cell{Type: "furniture", Cleaned: false, Obstacle: true, ObstacleName: f.Name}
			}
		}
	}

	// count cleanable cells
	cleanableCellCount := 0
	for i := 0; i < gridWidth; i++ {
		for j := 0; j < gridHeight; j++ {
			if !grid[i][j].Obstacle {
				cleanableCellCount++
			}
		}
	}

	return &Room{
		Grid: grid,
		Width: gridWidth,
		Height: gridHeight,
		CleanableCellCount: cleanableCellCount,
		CleanedCellCount: 0,
		Animate: animate,
	}
}

func (room *Room) Display(robot *Robot, showPath bool) {
	// in windows, we can use github.com/inancgumus/screen
	// call screen.Clean()

	// clear the screen
	fmt.Print("\033[H\033[2J")

	for j := range room.Height {
		for i := range room.Width {
			if robot.Position.X == i && robot.Position.Y == j {
				fmt.Print(charRobot)
			} else if showPath && isInPath(Point{X: i, Y: j}, robot.Path) {
				fmt.Print(charPath)
			} else {
				cell := room.Grid[i][j]
				switch cell.Type {
				case "wall":
					fmt.Print(charWall)
				case "furniture":
					fmt.Print(charFurniture)
				case "clean":
					fmt.Print(charClean)
				case "dirty":
					fmt.Print(charDirty)
				}
			}
		}
		fmt.Println()
	}

	// display cleaning progress
	percentCleaned := float64(room.CleanedCellCount) / float64(room.CleanableCellCount) * 100
	fmt.Printf(
		"cleaning progress: %.2f%% (%d/%d) cells cleaned\n",
		percentCleaned,
		room.CleanedCellCount,
		room.CleanableCellCount,
	)
}

func (room *Room) IsValid(x, y int) bool {
	return x >= 0 && x < room.Width && y >= 0 && y < room.Height && !room.Grid[x][y].Obstacle
}

func LoadRoomConfig(filename string) (*RoomConfig, error) {
	// read the json file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading json file: %w", err)
	}

	// parse the json file
	var config RoomConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}

	return &config, nil
}

func isInPath(point Point, path []Point) bool {
	for _, p := range path {
		if p.X == point.X && p.Y == point.Y {
			return true
		}
	}

	return false
}

func displaySummary(room *Room, robot *Robot, moveCount int, cleaningTime time.Duration) {
	// display the final room state with the robot's path
	fmt.Println("\nFinal room state with robot's path")
	room.Display(robot, true)

	// cleaning summary information
	fmt.Println("\n=========== Cleaning Summary ===========")
	fmt.Println()
	fmt.Printf(
		"room size: %d x %d (%d cm x %d cm)\n",
		room.Width,
		room.Height,
		room.Width*cellSize,
		room.Height*cellSize,
	)

	// calculate coverage percentage
	percentCleaned := float64(room.CleanedCellCount) / float64(room.CleanableCellCount) * 100
	fmt.Printf(
		"coverage: %.2f%% (%d/%d cells cleaned)\n",
		percentCleaned,
		room.CleanedCellCount,
		room.CleanableCellCount,
	)

	// display time and moves
	fmt.Printf("total moves: %d\n", moveCount)
	fmt.Printf("cleaning time: %v\n", cleaningTime)

	// calculate efficiency (cells cleaned per move)
	efficiency := float64(room.CleanedCellCount) / float64(moveCount)
	fmt.Printf("efficiency: %.2f cells cleaned per move\n", efficiency)

	fmt.Println()
	fmt.Println("========================================")
}