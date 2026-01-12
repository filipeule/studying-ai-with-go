package main

type PriorityQueueAStar []*Node

func (pq PriorityQueueAStar) Len() int {
	return len(pq)
}

func (pq PriorityQueueAStar) Less(i, j int) bool {
	return int(pq[i].EstimatedCostToGoal) < int(pq[j].EstimatedCostToGoal)
}

func (pq PriorityQueueAStar) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueueAStar) Push(x any) {
	n := len(*pq)
	item := x.(*Node)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueueAStar) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0:n-1]
	return item
}
