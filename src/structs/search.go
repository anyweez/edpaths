package structs

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
)

type SpaceBucket struct {
	Systems []*SpaceSystem
	X       int
	Y       int
	Z       int
}

/**
 * Options provided to the routing algorithm. These can be either hard or soft constraints.
 */
type RoutingConstraints struct {
	MaxJump float64
	MaxHops int
}

type SpaceRoute struct {
	Origin      *SpaceSystem
	Destination *SpaceSystem
	Stops       []*SpaceSystem
	Distance    float64

	// Debug info, probably to be removed
	Checks int // number of sites that needed to be checked. fewer is faster.
}

const (
	UniverseMin = -16899.750
	UniverseMax = 65630.156
)

type SpaceGraph struct {
	Buckets [][][]*SpaceBucket
	Radius  float64
	systems map[SystemID]*SpaceSystem // access via Get()
}

/**
 * Identifies the bucket coordinates of the given system.
 */
func (graph *SpaceGraph) FindBucket(system *SpaceSystem) (SystemID, SystemID, SystemID) {
	bucketSize := math.Ceil((UniverseMax - UniverseMin) / float64(len(graph.Buckets)))

	x := int(math.Floor((system.X - UniverseMin) / bucketSize))
	y := int(math.Floor((system.Y - UniverseMin) / bucketSize))
	z := int(math.Floor((system.Z - UniverseMin) / bucketSize))

	return SystemID(x), SystemID(y), SystemID(z)
}

// TODO: find best route among combination of points

func (graph *SpaceGraph) FindRoute(from *SpaceSystem, to *SpaceSystem, cons *RoutingConstraints) *SpaceRoute {
	return graph._findRoute(from, to, cons, 0)
}

func (graph *SpaceGraph) FindPath(from *SpaceSystem, to *SpaceSystem, cons *RoutingConstraints) *SpaceRoute {
	visited := make(map[*SpaceSystem]bool)
	queued := make(map[*SpaceSystem]bool)
	path := make(map[*SpaceSystem]*SpaceSystem)
	available := NewDestinationQueue(to)

	heap.Push(&available, &SearchStop{Location: from, Hops: 0})
	queued[from] = true

	costFromOrigin := make(map[*SpaceSystem]float64)    // known
	costToDestination := make(map[*SpaceSystem]float64) // blend of known and heuristic

	costFromOrigin[from] = 0.0
	costToDestination[from] = TravelCost(from, to)

	checks := 0
	// Based on A* pseudocode from Wikipedia:
	//    https://en.wikipedia.org/wiki/A*_search_algorithm#Pseudocode\
	for available.Len() > 0 {
		current := heap.Pop(&available).(*SearchStop)

		// If we've exceeded the maximum number of hops, abandon this route and move on
		// to the next.
		if current.Hops > cons.MaxHops {
			continue
		}

		checks++

		visited[current.Location] = true // mark the current location as visited
		// fmt.Printf("Checking %+v (distance %.2f)\n", current.Location, TravelCost(current.Location, to))

		// 	// Return success!
		if current.Location.ID == to.ID {
			var unwound []*SpaceSystem
			current := to

			for current != nil {
				unwound = append([]*SpaceSystem{current}, unwound...)
				current = path[current]
			}

			distance := 0.0
			for i := 1; i < len(unwound); i++ {
				distance += unwound[i].DistanceTo(unwound[i-1])
			}

			return &SpaceRoute{
				Origin:      from,
				Destination: to,
				Distance:    distance,
				Stops:       unwound,
				Checks:      checks,
			}
		}

		// Investigate each neighbor if they haven't been investigated yet (if they have then we already found a
		// shorter way to get there and a loop isn't going to help).
		for _, near := range graph.Within(current.Location, cons.MaxJump) {
			if !visited[near] {
				score := costFromOrigin[current.Location] + current.Location.DistanceTo(near)

				// If its not already being searched, add it to the queue
				if !queued[near] {
					heap.Push(&available, &SearchStop{
						Location: near,
						Hops:     current.Hops + 1,
					})
					queued[near] = true
				} else if score >= costFromOrigin[near] {
					continue
				}

				// fmt.Printf("Juicy new path from %s to %s\n  So far: %.2f\n  Remaining: %.2f\n", current.Location.Name, near.Name, score, TravelCost(near, to))
				path[near] = current.Location
				costFromOrigin[near] = score
				costToDestination[near] = costFromOrigin[current.Location] + TravelCost(current.Location, to)
			}
		}
	}

	return nil
}

/**
 * Return pointers to all SpaceSystem's within the specified radius of the origin. The origin currently needs
 * to be a SpaceSystem but this could conceivably work with any point.
 *
 * This currently only looks through the same bucket as the origin system. Need to expand that.
 */
func (graph *SpaceGraph) Within(origin *SpaceSystem, radius float64) []*SpaceSystem {
	var items []*SpaceSystem

	x, y, z := graph.FindBucket(origin)
	for _, loc := range graph.Buckets[x][y][z].Systems {
		if origin.DistanceTo(loc) < radius {
			items = append(items, loc)
		}
	}

	return items
}

func (graph *SpaceGraph) _findRoute(from *SpaceSystem, to *SpaceSystem, cons *RoutingConstraints, depth int) *SpaceRoute {
	// Base case: found the destination
	if from.ID == to.ID {
		return &SpaceRoute{
			Origin:      from,
			Destination: to,
			Distance:    0,
			Stops:       []*SpaceSystem{from},
		}
	} else if depth == cons.MaxHops {
		return nil
	}

	// Get all jump options
	options := graph.Within(from, cons.MaxJump)
	// Sort by distance of destination with `to`

	// Recursively call all options
	var best *SpaceRoute
	for _, dest := range options {
		route := graph._findRoute(dest, to, cons, depth+1)

		if best == nil || route.Distance < best.Distance {
			best = route
		}
	}

	return best
}

func (graph *SpaceGraph) Add(system *SpaceSystem) error {
	if system.X > UniverseMax || system.X < UniverseMin ||
		system.Y > UniverseMax || system.Y < UniverseMin ||
		system.Z > UniverseMax || system.Z < UniverseMin {
		return errors.New("System " + system.Name + " out of currently supported bounds.")
	}
	// Find coordinates for this system
	x, y, z := graph.FindBucket(system)

	// fmt.Printf("adding to (%d, %d, %d)\n", x, y, z)
	// Add to the cell's buckets
	system.Bucket = graph.Buckets[x][y][z]
	graph.Buckets[x][y][z].Systems = append(graph.Buckets[x][y][z].Systems, system)
	graph.systems[system.ID] = system

	return nil
}

func (graph *SpaceGraph) Get(id SystemID) *SpaceSystem {
	if system, exists := graph.systems[id]; exists {
		return system
	} else {
		return nil
	}
}

func InitGraph(radius float64) *SpaceGraph {
	graph := new(SpaceGraph)
	graph.Radius = radius
	graph.systems = make(map[SystemID]*SpaceSystem)

	count := int(math.Ceil((UniverseMax - UniverseMin) / radius))
	fmt.Printf("Initializing graph (%d^3)...\n", count)
	graph.Buckets = make([][][]*SpaceBucket, count)

	// Make all the space buckets!
	for i := 0; i < count; i++ {
		graph.Buckets[i] = make([][]*SpaceBucket, count)

		for j := 0; j < count; j++ {
			graph.Buckets[i][j] = make([]*SpaceBucket, 0, count)

			for k := 0; k < count; k++ {
				bucket := &SpaceBucket{
					X:       i,
					Y:       j,
					Z:       k,
					Systems: make([]*SpaceSystem, 0),
				}

				graph.Buckets[i][j] = append(graph.Buckets[i][j], bucket)
			}
		}
	}

	fmt.Println("Populating graph...")

	return graph
}

func (graph *SpaceGraph) Load(db *SpaceDB) *SpaceGraph {
	db.ForEachSystem(func(system SpaceSystem) {
		graph.Add(&system)
	})

	return graph
}

func (graph *SpaceGraph) LoadSample() *SpaceGraph {
	graph.Add(&SpaceSystem{ID: 1, Name: "First Site", X: 5, Y: 3, Z: 5})
	graph.Add(&SpaceSystem{ID: 2, Name: "Second Site", X: 0, Y: -1, Z: 2})
	graph.Add(&SpaceSystem{ID: 3, Name: "Third Site", X: 5, Y: 2, Z: 5})
	graph.Add(&SpaceSystem{ID: 4, Name: "Fourth Site", X: 0, Y: 0, Z: 0})
	graph.Add(&SpaceSystem{ID: 5, Name: "Fifth Site", X: 6, Y: 5, Z: 5})
	graph.Add(&SpaceSystem{ID: 6, Name: "Sixth Site", X: 2.5, Y: 2.5, Z: 2.5})

	return graph
}
