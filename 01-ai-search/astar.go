package main

import (
	"container/heap"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

const FloodedCost = 1000

type AStarSearch struct {
	Frontier PriorityQueueAStar
	Game     *Maze
}

func (as *AStarSearch) GetFrontier() []*Node {
	return as.Frontier
}

func (as *AStarSearch) Add(i *Node) {
	i.CostToGoal = i.ManhattanDistance(as.Game.Start)
	i.EstimatedCostToGoal = euclideanDist(i.State, as.Game.Goal) + float64(i.CostToGoal)

	// is this cell flooded?
	if i.State.Water {
		i.EstimatedCostToGoal += FloodedCost
	}

	as.Frontier.Push(i)
	heap.Init(&as.Frontier)
}

func (as *AStarSearch) ContainsState(i *Node) bool {
	for _, x := range as.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (as *AStarSearch) Empty() bool {
	return len(as.Frontier) == 0
}

func (as *AStarSearch) Remove() (*Node, error) {
	if len(as.Frontier) > 0 {
		if as.Game.Debug {
			fmt.Println("frontier before remove:")
			for _, x := range as.Frontier {
				fmt.Println("node:", x.State)
			}
		}

		return heap.Pop(&as.Frontier).(*Node), nil
	}
	return nil, errors.New("frontier is empty")
}

func (as *AStarSearch) Solve() {
	fmt.Println("starting to solve maze using A* search...")

	as.Game.NumExplored = 0

	start := Node{
		State:  as.Game.Start,
		Parent: nil,
		Action: "",
	}

	as.Add(&start)
	as.Game.CurrentNode = &start

	for {
		if as.Empty() {
			return
		}

		currentNode, err := as.Remove()
		if err != nil {
			fmt.Println(err)
			return
		}

		if as.Game.Debug {
			fmt.Println("removed", currentNode.State)
			fmt.Println("---------")
			fmt.Println()
		}

		as.Game.CurrentNode = currentNode
		as.Game.NumExplored += 1

		// have we found the solution?
		if as.Game.Goal == currentNode.State {
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

			as.Game.Solution = Solution{
				Actions: actions,
				Cells:   cells,
			}
			as.Game.Explored = append(as.Game.Explored, currentNode.State)
			break
		}

		as.Game.Explored = append(as.Game.Explored, currentNode.State)

		if as.Game.Animate {
			as.Game.OutputImage(fmt.Sprintf("tmp/%06d.png", as.Game.NumExplored))
		}

		for _, x := range as.Neighbors(currentNode) {
			if !as.ContainsState(x) {
				if !inExplored(x.State, as.Game.Explored) {
					as.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (as *AStarSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < as.Game.Height {
			if 0 <= x.State.Col && x.State.Col < as.Game.Width {
				if !as.Game.Walls[x.State.Row][x.State.Col].wall {
					if as.Game.Walls[x.State.Row][x.State.Col].State.Water {
						x.State.Water = true
					} 

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
