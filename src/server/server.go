package main

import (
	"fmt"
	"structs" // local
)

func main() {
	db := structs.Connect()

	// db.ForEachSystem(func(system structs.SpaceSystem) {
	// 	fmt.Printf("%+v\n", system)
	// })
	// fmt.Printf("from:  %+v\n", start)
	// fmt.Printf("  to:  %+v\n", end)

	graph := structs.InitGraph(1000).Load(db)

	start := graph.Get(69374)
	end := graph.Get(704)

	fmt.Printf("from:  %+v\n @ (%d, %d, %d)\n", start, start.Bucket.X, start.Bucket.Y, start.Bucket.Z)
	fmt.Printf("  to:  %+v\n @ (%d, %d, %d)\n", end, end.Bucket.X, end.Bucket.Y, end.Bucket.Z)

	fmt.Println("Graph loaded. Attempting to find route.")
	route := graph.FindPath(start, end, &structs.RoutingConstraints{MaxJump: 18.0, MaxHops: 8})

	if route == nil {
		fmt.Println("No route found.")
	} else {
		/** Print out results **/
		fmt.Printf("Route [%s] => [%s]\n", start.Name, end.Name)
		fmt.Printf("%d jumps, %.2f light years\n", len(route.Stops), route.Distance)

		last := start
		for i, stop := range route.Stops {
			fmt.Printf(" %d: %s (%.2f ly)\n", i+1, stop.Name, stop.DistanceTo(last))
			last = stop
		}
	}
}
