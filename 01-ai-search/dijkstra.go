package main

import (
	"container/heap"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

type DijkstraSearch struct {
	Frontier PriorityQueueDijkstra
	Game     *Maze
}

func (ds *DijkstraSearch) GetFrontier() []*Node {
	return ds.Frontier
}

func (ds *DijkstraSearch) Add(i *Node) {
	i.CostToGoal = i.ManhattanDistance(ds.Game.Start)
	ds.Frontier.Push(i)
	heap.Init(&ds.Frontier)
}

func (ds *DijkstraSearch) ContainsState(i *Node) bool {
	for _, x := range ds.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (ds *DijkstraSearch) Empty() bool {
	return len(ds.Frontier) == 0
}

func (ds *DijkstraSearch) Remove() (*Node, error) {
	if len(ds.Frontier) > 0 {
		if ds.Game.Debug {
			fmt.Println("frontier before remove:")
			for _, x := range ds.Frontier {
				fmt.Println("node:", x.State)
			}
		}

		return heap.Pop(&ds.Frontier).(*Node), nil
	}
	return nil, errors.New("frontier is empty")
}

func (ds *DijkstraSearch) Solve() {
	fmt.Println("starting to solve maze using dijkstra search...")

	ds.Game.NumExplored = 0

	start := Node{
		State:  ds.Game.Start,
		Parent: nil,
		Action: "",
	}

	ds.Add(&start)
	ds.Game.CurrentNode = &start

	for {
		if ds.Empty() {
			return
		}

		currentNode, err := ds.Remove()
		if err != nil {
			fmt.Println(err)
			return
		}

		if ds.Game.Debug {
			fmt.Println("removed", currentNode.State)
			fmt.Println("---------")
			fmt.Println()
		}

		ds.Game.CurrentNode = currentNode
		ds.Game.NumExplored += 1

		// have we found the solution?
		if ds.Game.Goal == currentNode.State {
			var actions []string
			var cells []Point

			for {
				if currentNode.Parent != nil {
					actions = append(actions, currentNode.Action)
					cells = append(cells, currentNode.State)
					currentNode = currentNode.Parent
				} else {
					break
				}
			}

			slices.Reverse(actions)
			slices.Reverse(cells)

			ds.Game.Solution = Solution{
				Actions: actions,
				Cells:   cells,
			}
			ds.Game.Explored = append(ds.Game.Explored, currentNode.State)
			break
		}

		ds.Game.Explored = append(ds.Game.Explored, currentNode.State)

		if ds.Game.Animate {
			ds.Game.OutputImage(fmt.Sprintf("tmp/%06d.png", ds.Game.NumExplored))
		}

		for _, x := range ds.Neighbors(currentNode) {
			if !ds.ContainsState(x) {
				if !inExplored(x.State, ds.Game.Explored) {
					ds.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (ds *DijkstraSearch) Neighbors(node *Node) []*Node {
	row := node.State.Row
	col := node.State.Col

	candidates := []*Node{
		{State: Point{Row: row - 1, Col: col}, Parent: node, Action: "up"},
		{State: Point{Row: row + 1, Col: col}, Parent: node, Action: "down"},
		{State: Point{Row: row, Col: col - 1}, Parent: node, Action: "left"},
		{State: Point{Row: row, Col: col + 1}, Parent: node, Action: "right"},
	}

	var neighbors []*Node
	for _, x := range candidates {
		if 0 <= x.State.Row && x.State.Row < ds.Game.Height {
			if 0 <= x.State.Col && x.State.Col < ds.Game.Width {
				if !ds.Game.Walls[x.State.Row][x.State.Col].wall {
					neighbors = append(neighbors, x)
				}
			}
		}
	}

	// randomness
	for i := range neighbors {
		j := rand.Intn(i + 1)
		neighbors[i], neighbors[j] = neighbors[j], neighbors[i]
	}

	return neighbors
}
