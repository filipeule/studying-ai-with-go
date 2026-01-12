package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	DFS = iota
	BFS
	GBFS
	ASTAR
	DIJKSTRA
)

type Point struct {
	Row   int
	Col   int
	Water bool
}

type Wall struct {
	State Point
	wall  bool
}

type Node struct {
	index               int
	State               Point
	Parent              *Node
	Action              string
	CostToGoal          int
	EstimatedCostToGoal float64
}

func (n *Node) ManhattanDistance(goal Point) int {
	return abs(n.State.Row-goal.Row) + abs(n.State.Col-goal.Col)
}

type Solution struct {
	Actions []string
	Cells   []Point
}

type Maze struct {
	Height      int
	Width       int
	Start       Point
	Goal        Point
	Walls       [][]Wall
	CurrentNode *Node
	Solution    Solution
	Explored    []Point
	Steps       int
	NumExplored int
	Debug       bool
	SearchType  int
	Animate     bool
}

func (g *Maze) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening %s: %s\n", filename, err)
	}
	defer f.Close()

	var fileContents []string

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("cannot open file %s: %s", filename, err)
		}

		fileContents = append(fileContents, line)
	}

	foundStart, foundEnd := false, false
	for _, line := range fileContents {
		if strings.Contains(line, "A") {
			foundStart = true
		}

		if strings.Contains(line, "B") {
			foundEnd = true
		}
	}

	if !foundStart {
		return errors.New("starting location not found")
	}

	if !foundEnd {
		return errors.New("ending location not found")
	}

	g.Height = len(fileContents)
	g.Width = len(fileContents[0])

	var rows [][]Wall
	for i, row := range fileContents {
		var cols []Wall
		for j, col := range row {
			curLetter := fmt.Sprintf("%c", col)

			var wall Wall
			wall.State.Row = i
			wall.State.Col = j

			switch curLetter {
			case "A":
				g.Start = Point{Row: i, Col: j}
				wall.wall = false
			case "B":
				g.Goal = Point{Row: i, Col: j}
				wall.wall = false
			case " ":
				wall.wall = false
			case "w":
				wall.wall = false
				wall.State.Water = true
			case "#":
				wall.wall = true
			default:
				continue
			}

			cols = append(cols, wall)
		}
		rows = append(rows, cols)
	}

	g.Walls = rows
	return nil
}

func (g *Maze) inSolution(x Point) bool {
	for _, step := range g.Solution.Cells {
		if step.Row == x.Row && step.Col == x.Col {
			return true
		}
	}
	return false
}

func (g *Maze) PrintMaze() {
	for r, row := range g.Walls {
		for c, col := range row {
			if col.wall {
				fmt.Print("◼")
			} else if g.Start.Row == col.State.Row && g.Start.Col == col.State.Col {
				fmt.Print("A")
			} else if g.Goal.Row == col.State.Row && g.Goal.Col == col.State.Col {
				fmt.Print("B")
			} else if g.inSolution(Point{r, c, false}) {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func init() {
	_ = os.Mkdir("./tmp", os.ModePerm)
	emptyTmp()
}

func main() {
	var m Maze
	var maze, searchType string

	flag.StringVar(&maze, "file", "maze.txt", "maze file")
	flag.StringVar(&searchType, "search", "dfs", "search type")
	flag.BoolVar(&m.Debug, "debug", false, "write debugging info")
	flag.BoolVar(&m.Animate, "animate", false, "produce animation")
	flag.Parse()

	err := m.Load(maze)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	startTime := time.Now()

	switch searchType {
	case "dfs":
		m.SearchType = DFS
		solveDFS(&m)
	case "bfs":
		m.SearchType = BFS
		solveBFS(&m)
	case "dijkstra":
		m.SearchType = DIJKSTRA
		solveDijkstra(&m)
	case "gbfs":
		m.SearchType = GBFS
		solveGBFS(&m)
	case "astar":
		m.SearchType = ASTAR
		solveAStar(&m)
	default:
		fmt.Println("invalid search type")
		os.Exit(1)
	}

	if len(m.Solution.Actions) > 0 {
		fmt.Println("solution:")
		// m.PrintMaze()
		fmt.Println("solution is", len(m.Solution.Cells), "steps")
		fmt.Println("time to solve:", time.Since(startTime))
		m.OutputImage("image.png")
	} else {
		fmt.Println("no solution")
	}

	fmt.Println("explored", len(m.Explored), "nodes")

	if m.Animate {
		fmt.Println("building animation...")
		m.OutputAnimatedImage()
		fmt.Println("done!")
	}
}

func solveDFS(m *Maze) {
	var s DepthFirstSearch
	s.Game = m
	fmt.Println("goal is", s.Game.Goal)
	s.Solve()
}

func solveBFS(m *Maze) {
	var s BreadthFirstSearch
	s.Game = m
	fmt.Println("goal is", s.Game.Goal)
	s.Solve()
}

func solveDijkstra(m *Maze) {
	var s DijkstraSearch
	s.Game = m
	fmt.Println("goal is", s.Game.Goal)
	s.Solve()
}

func solveGBFS(m *Maze) {
	var s GreedyBestFirstSearch
	s.Game = m
	fmt.Println("goal is", s.Game.Goal)
	s.Solve()
}

func solveAStar(m *Maze) {
	var s AStarSearch
	s.Game = m
	fmt.Println("goal is", s.Game.Goal)
	s.Solve()
}
