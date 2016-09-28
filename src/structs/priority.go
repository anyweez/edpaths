package structs

import "container/heap"

type SearchStop struct {
	Location *SpaceSystem
	Hops     int
}

func NewDestinationQueue(destination *SpaceSystem) destinationQueue {
	dest := destinationQueue{
		destination: destination,
		elements:    make([]*SearchStop, 0), // todo: customize cap?
	}

	heap.Init(&dest)

	return dest
}

/* TravelCost is the heuristic function for sorting queue */
func TravelCost(from *SpaceSystem, to *SpaceSystem) float64 {
	return from.DistanceTo(to)
}

type destinationQueue struct {
	destination *SpaceSystem
	elements    []*SearchStop
}

func (q destinationQueue) Len() int { return len(q.elements) }

func (q destinationQueue) Less(i int, j int) bool {
	return TravelCost(q.elements[i].Location, q.destination) < TravelCost(q.elements[j].Location, q.destination)
}

func (q destinationQueue) Swap(i int, j int) {
	q.elements[i], q.elements[j] = q.elements[j], q.elements[i]
}

func (q *destinationQueue) Push(in interface{}) {
	q.elements = append(q.elements, in.(*SearchStop))
}

func (q *destinationQueue) Pop() interface{} {
	next := q.elements[len(q.elements)-1]
	q.elements = q.elements[0 : len(q.elements)-1]

	return next
}
