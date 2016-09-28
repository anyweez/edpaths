package structs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	graph := InitGraph(1000).LoadSample()

	from := graph.Get(1)
	to := graph.Get(4)

	fmt.Printf("Seeking path from:\n  %+v\nto:\n  %+v\n", from, to)

	path := graph.FindPath(from, to, &RoutingConstraints{
		MaxHops: 5,
		MaxJump: 5,
	})
	expected := []int{1, 6, 4}

	if path != nil {
		fmt.Println("Route:")
		for i := len(path.Stops) - 1; i >= 0; i-- {
			fmt.Printf("  - %d. %+v\n", i+1, path.Stops[i])
			assert.NotEqual(t, path.Stops[i], expected[i], "deviated from expected route")
		}
		fmt.Printf("Subpaths checked: %d\n", path.Checks)
	} else {
		fmt.Println("No path found")
	}
}
