package animation

type jointValItem struct {
	val   float64
	joint int
}

type minJointValPriorityQueue []jointValItem

func (pq minJointValPriorityQueue) Len() int { return len(pq) }

func (pq minJointValPriorityQueue) Less(i, j int) bool {
	return pq[i].val < pq[j].val
}

func (pq minJointValPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *minJointValPriorityQueue) Push(x any) {
	item := x.(jointValItem)
	*pq = append(*pq, item)
}

func (pq *minJointValPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type maxJointValPriorityQueue []jointValItem

func (pq maxJointValPriorityQueue) Len() int { return len(pq) }

func (pq maxJointValPriorityQueue) Less(i, j int) bool {
	return pq[i].val > pq[j].val
}

func (pq maxJointValPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *maxJointValPriorityQueue) Push(x any) {
	item := x.(jointValItem)
	*pq = append(*pq, item)
}

func (pq *maxJointValPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
