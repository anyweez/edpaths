package structs

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"math/rand"
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
	Origin      *SpaceStop
	Destination *SpaceStop
	Stops       []*SpaceStop
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

		// Return success! We've reached our destination
		if current.Location.ID == to.ID {
			var unwound []*SpaceStop
			var prev *SpaceSystem
			current := to

			for current != nil {
				stop := current.AsStop()
				unwound = append([]*SpaceStop{stop}, unwound...)

				prev = current
				current = path[current]

				if current != nil {
					stop.DistanceFromPrev = current.DistanceTo(prev)
				}
			}

			distance := 0.0
			for i := 1; i < len(unwound); i++ {
				distance += unwound[i].DistanceFromPrev
				// distance += unwound[i].DistanceTo(unwound[i-1])
			}

			return &SpaceRoute{
				Origin:      from.AsStop(),
				Destination: to.AsStop(),
				Distance:    distance,
				Stops:       unwound,
				Checks:      checks,
			}
		}

		// Investigate each neighbor if they haven't been investigated yet (if they have then we already found a
		// shorter way to get there and a loop isn't going to help).
		for _, near := range graph.Proximity(current.Location, cons.MaxJump) {
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
func (graph *SpaceGraph) Proximity(origin *SpaceSystem, radius float64) []*SpaceSystem {
	var items []*SpaceSystem

	// Get all nearby buckets (including the current one) and scan all systems in each bucket.
	buckets := graph.NearbyBuckets(origin, radius)
	for _, bucket := range buckets {
		for _, loc := range bucket.Systems {
			if origin.DistanceTo(loc) < radius {
				items = append(items, loc)
			}
		}
	}

	return items
}

/**
 * Return all buckets within the radius of the origin, including the bucket the origin is currently in.
 * It should only return each bucket one time, and buckets will be pointers to the actual buckets.
 */
func (graph *SpaceGraph) NearbyBuckets(origin *SpaceSystem, radius float64) []*SpaceBucket {
	start := graph.GetBucket(graph.FindBucket(origin))
	buckets := []*SpaceBucket{start}

	// Get the bucket in the -x direction
	xDown := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X - radius,
		Y: origin.Y,
		Z: origin.Z,
	}))

	if xDown != start {
		buckets = append(buckets, xDown)
	}

	// Get the bucket in the +x direction
	xUp := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X + radius,
		Y: origin.Y,
		Z: origin.Z,
	}))

	if xUp != start {
		buckets = append(buckets, xUp)
	}

	// Get the bucket in the -y direction
	yDown := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X,
		Y: origin.Y - radius,
		Z: origin.Z,
	}))

	if yDown != start {
		buckets = append(buckets, yDown)
	}

	// Get the bucket in the +y direction
	yUp := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X,
		Y: origin.Y + radius,
		Z: origin.Z,
	}))

	if yUp != start {
		buckets = append(buckets, yUp)
	}

	// Get the bucket in the -z direction
	zDown := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X,
		Y: origin.Y,
		Z: origin.Z - radius,
	}))

	if zDown != start {
		buckets = append(buckets, zDown)
	}

	// Get the bucket in the +z direction
	zUp := graph.GetBucket(graph.FindBucket(&SpaceSystem{
		X: origin.X,
		Y: origin.Y,
		Z: origin.Z + radius,
	}))

	if zUp != start {
		buckets = append(buckets, zUp)
	}

	return buckets
}

/**
 * Simple function to provide an abstraction over the bucket storage schema.
 */
func (graph *SpaceGraph) GetBucket(x SystemID, y SystemID, z SystemID) *SpaceBucket {
	return graph.Buckets[x][y][z]
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
	system.Bucket = graph.GetBucket(x, y, z)
	system.Bucket.Systems = append(system.Bucket.Systems, system)
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

/**
 * This function is NOT performant and should not be used in production algorithms.
 * Currently exists for testing only.
 */
func (graph *SpaceGraph) GetRandom() *SpaceSystem {
	target := rand.Int() % len(graph.systems)

	current := 0
	for _, system := range graph.systems {
		if current == target {
			return system
		}

		current++
	}

	return nil
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

	return graph
}

func (graph *SpaceGraph) Load(db *SpaceDB) *SpaceGraph {
	fmt.Println("Populating graph...")

	count := 0
	// TODO: object copy here is probably grotesquely inefficient
	db.ForEachSystem(func(system *SpaceSystem) {
		graph.Add(system)
		count++
	})

	fmt.Printf("Loaded %d systems.\n", count)

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

/**
 * Merge two routes together. The callee is modified to include data from the extension (parameter).
 */
func (route *SpaceRoute) Merge(next *SpaceRoute) {
	// Copy all stops except the first (which should be the last of `route`)
	route.Stops = append(route.Stops, next.Stops[1:len(next.Stops)]...)
	route.Destination = next.Destination

	route.Distance += next.Distance
	route.Checks += next.Checks
}
