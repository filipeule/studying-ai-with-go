package main

import (
	"container/heap"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

type GreedyBestFirstSearch struct {
	Frontier PriorityQueueGBFS
	Game     *Maze
}

func (gs *GreedyBestFirstSearch) GetFrontier() []*Node {
	return gs.Frontier
}

func (gs *GreedyBestFirstSearch) Add(i *Node) {
	i.CostToGoal = i.ManhattanDistance(gs.Game.Goal)
	gs.Frontier.Push(i)
	heap.Init(&gs.Frontier)
}

func (gs *GreedyBestFirstSearch) ContainsState(i *Node) bool {
	for _, x := range gs.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (gs *GreedyBestFirstSearch) Empty() bool {
	return len(gs.Frontier) == 0
}

func (gs *GreedyBestFirstSearch) Remove() (*Node, error) {
	if len(gs.Frontier) > 0 {
		if gs.Game.Debug {
			fmt.Println("frontier before remove:")
			for _, x := range gs.Frontier {
				fmt.Println("node:", x.State)
			}
		}

		return heap.Pop(&gs.Frontier).(*Node), nil
	}
	return nil, errors.New("frontier is empty")
}

func (gs *GreedyBestFirstSearch) Solve() {
	fmt.Println("starting to solve maze using greedy best first search...")

	gs.Game.NumExplored = 0

	start := Node{
		State:  gs.Game.Start,
		Parent: nil,
		Action: "",
	}

	gs.Add(&start)
	gs.Game.CurrentNode = &start

	for {
		if gs.Empty() {
			return
		}

		currentNode, err := gs.Remove()
		if err != nil {
			fmt.Println(err)
			return
		}

		if gs.Game.Debug {
			fmt.Println("removed", currentNode.State)
			fmt.Println("---------")
			fmt.Println()
		}

		gs.Game.CurrentNode = currentNode
		gs.Game.NumExplored += 1

		// have we found the solution?
		if gs.Game.Goal == currentNode.State {
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

			gs.Game.Solution = Solution{
				Actions: actions,
				Cells:   cells,
			}
			gs.Game.Explored = append(gs.Game.Explored, currentNode.State)
			break
		}

		gs.Game.Explored = append(gs.Game.Explored, currentNode.State)

		if gs.Game.Animate {
			gs.Game.OutputImage(fmt.Sprintf("tmp/%06d.png", gs.Game.NumExplored))
		}

		for _, x := range gs.Neighbors(currentNode) {
			if !gs.ContainsState(x) {
				if !inExplored(x.State, gs.Game.Explored) {
					gs.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (gs *GreedyBestFirstSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < gs.Game.Height {
			if 0 <= x.State.Col && x.State.Col < gs.Game.Width {
				if !gs.Game.Walls[x.State.Row][x.State.Col].wall {
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
