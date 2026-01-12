package main

type PriorityQueueDijkstra []*Node

func (pq PriorityQueueDijkstra) Len() int {
	return len(pq)
}

func (pq PriorityQueueDijkstra) Less(i, j int) bool {
	return pq[i].CostToGoal < pq[j].CostToGoal
}

func (pq PriorityQueueDijkstra) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueueDijkstra) Push(x any) {
	n := len(*pq)
	item := x.(*Node)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueueDijkstra) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0:n-1]
	return item
}
