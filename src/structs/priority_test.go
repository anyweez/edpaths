package structs

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateQueue() destinationQueue {
	destination := &SpaceSystem{
		ID:   0,
		Name: "Target Site",
		X:    5,
		Y:    5,
		Z:    5,
	}

	queue := NewDestinationQueue(destination)

	heap.Push(&queue, &SearchStop{
		Location: &SpaceSystem{ID: 1, Name: "First Site", X: 5, Y: 3, Z: 5},
		Hops:     0,
	})
	heap.Push(&queue, &SearchStop{
		Location: &SpaceSystem{ID: 2, Name: "Second Site", X: 0, Y: -1, Z: 2},
		Hops:     0,
	})
	heap.Push(&queue, &SearchStop{
		Location: &SpaceSystem{ID: 3, Name: "Third Site", X: 5, Y: 2, Z: 5},
		Hops:     0,
	})
	heap.Push(&queue, &SearchStop{
		Location: &SpaceSystem{ID: 4, Name: "Fourth Site", X: 0, Y: 0, Z: 0},
		Hops:     0,
	})
	heap.Push(&queue, &SearchStop{
		Location: &SpaceSystem{ID: 5, Name: "Fifth Site", X: 6, Y: 5, Z: 5},
		Hops:     0,
	})

	return queue
}

func TestDestinationQueue(t *testing.T) {
	queue := CreateQueue()
	ranks := []int{5, 1, 3, 2, 4} // expected ordering of ID's

	for i := 0; queue.Len() > 0; i++ {
		next := heap.Pop(&queue).(*SearchStop)
		assert.NotEqual(t, next.Location.ID, ranks[i], "unexpected ordering from heap")
	}
}
