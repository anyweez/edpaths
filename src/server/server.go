package main

import (
	"net/http"
	"strconv"
	"structs" // local

	"github.com/gin-gonic/gin"
)

type RouteResponse struct {
	Status int
	Route  *structs.SpaceRoute
}

func main() {
	db := structs.Connect()
	graph := structs.InitGraph(1000).Load(db)

	router := gin.Default()

	router.GET("/route", func(ctx *gin.Context) {
		if ctx.Query("from") == "" || ctx.Query("to") == "" {
			ctx.JSON(http.StatusBadRequest, RouteResponse{
				Status: http.StatusBadRequest,
			})

			return
		}

		startID, _ := strconv.Atoi(ctx.Query("from"))
		endID, _ := strconv.Atoi(ctx.Query("to"))

		start := graph.Get(structs.SystemID(startID)) // 69374
		end := graph.Get(structs.SystemID(endID))     // 704

		route := graph.FindPath(start, end, &structs.RoutingConstraints{MaxJump: 18.0, MaxHops: 8})

		if route == nil {
			ctx.JSON(http.StatusOK, RouteResponse{
				Status: http.StatusNotFound,
				Route:  nil,
			})
		} else {
			ctx.JSON(http.StatusOK, RouteResponse{
				Status: http.StatusOK,
				Route:  route,
			})
		}
	})

	router.Run()

	// fmt.Printf("from:  %+v\n @ (%d, %d, %d)\n", start, start.Bucket.X, start.Bucket.Y, start.Bucket.Z)
	// fmt.Printf("  to:  %+v\n @ (%d, %d, %d)\n", end, end.Bucket.X, end.Bucket.Y, end.Bucket.Z)

	// fmt.Println("Graph loaded. Attempting to find route.")

	// if route == nil {
	// 	fmt.Println("No route found.")
	// } else {
	// 	/** Print out results **/
	// 	fmt.Printf("Route [%s] => [%s]\n", start.Name, end.Name)
	// 	fmt.Printf("%d jumps, %.2f light years\n", len(route.Stops), route.Distance)

	// 	last := start
	// 	for i, stop := range route.Stops {
	// 		fmt.Printf(" %d: %s (%.2f ly)\n", i+1, stop.Name, stop.DistanceTo(last))
	// 		last = stop
	// 	}
	// }
}
