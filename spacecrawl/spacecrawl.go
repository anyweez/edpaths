package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anyweez/edpaths/structs" // local

	"github.com/fighterlyt/permutation"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
)

type RouteResponse struct {
	Status int
	Route  *structs.SpaceRoute
}

func main() {
	db := structs.Connect("sample")
	graph := structs.InitGraph(1000).Load(db)
	terms := structs.NewAutocomplete(db)

	router := gin.Default()
	router.Use(cors.Middleware(cors.Config{
		Origins:        "*",
		Methods:        "GET",
		RequestHeaders: "Origin, Authorization, Content-Type",
		ExposedHeaders: "",
		MaxAge:         50 * time.Second,
	}))

	/**
	 * Primary route; used for pathfinding between two destinations. This handler is fairly
	 * complex and should likely be refactored in the future.
	 *
	 * Currently it's considering three query params: `from` (id to start from), `to` (id to
	 * end on), and `visit`, a comma-delimited list of id's to visit in the middle. We find
	 * all permutations and solve each on its own goroutine, merging each of the subroutes
	 * together as we go. At the end we compare all of the combined routes to determine
	 * which is the shortest and return that.
	 */
	router.GET("/route", func(ctx *gin.Context) {
		if ctx.Query("from") == "" && ctx.Query("to") == "" {
			ctx.JSON(http.StatusBadRequest, RouteResponse{
				Status: http.StatusBadRequest,
			})

			return
		}
		fmt.Println("routing")

		var startID structs.SystemID // where to start
		var endID structs.SystemID   // where to end
		var visit []structs.SystemID // places to visit along the way

		if len(ctx.Query("from")) > 0 {
			id, _ := strconv.Atoi(ctx.Query("from"))
			startID = structs.SystemID(id)
		}

		if len(ctx.Query("to")) > 0 {
			id, _ := strconv.Atoi(ctx.Query("to"))
			endID = structs.SystemID(id)
		}

		if len(ctx.Query("visit")) > 0 {
			for _, rawID := range strings.Split(ctx.Query("visit"), ",") {
				id, _ := strconv.Atoi(rawID)
				visit = append(visit, structs.SystemID(id))
			}
		}

		// if we don't have any points, return error
		if len(visit) == 0 && startID == 0 && endID == 0 {
			ctx.JSON(http.StatusOK, RouteResponse{
				Status: http.StatusNotFound,
				Route:  nil,
			})
			return
		}

		// if we've only got an ending point, return it
		if len(visit) == 0 && endID == 0 {
			orig := graph.Get(startID).AsStop()
			orig.RequestedStop = true

			ctx.JSON(http.StatusOK, RouteResponse{
				Status: http.StatusOK,
				Route: &structs.SpaceRoute{
					Origin: orig,
				},
			})
			return
		}

		// if we've only got an ending point, return it
		if len(visit) == 0 && startID == 0 {
			dest := graph.Get(endID).AsStop()
			dest.RequestedStop = true

			ctx.JSON(http.StatusOK, RouteResponse{
				Status: http.StatusOK,
				Route: &structs.SpaceRoute{
					Destination: dest,
				},
			})
			return
		}

		routes := make(chan *structs.SpaceRoute, 10)
		var track sync.WaitGroup

		// Check all variants in parallel. We're looking for the combination that leads to the shortest overall
		// distance traveled.
		for _, variant := range getVariants(startID, endID, visit) {
			// Find routes for all legs of the journey and concat them together
			go func() {
				now := graph.Get(variant[0])
				next := graph.Get(variant[1])
				current := 1

				// Initial route; will be the base for all Merge() calls on this vairant.
				route := structs.SpaceRoute{
					Origin:      now.AsStop(),
					Destination: nil,
					Stops:       make([]*structs.SpaceStop, 0, len(visit)+2),
					Distance:    0,
					Checks:      0,
				}

				// Append the starting point (since it won't be copied from the first leg via Merge())
				route.Stops = append(route.Stops, &structs.SpaceStop{
					System:        now,
					RequestedStop: true,
				})

				for current < len(variant) {
					upcoming := graph.FindPath(now, next, &structs.RoutingConstraints{MaxJump: 18.0, MaxHops: 100})
					// Mark beginning and end as requested stops.
					upcoming.Stops[0].RequestedStop = true
					upcoming.Stops[len(upcoming.Stops)-1].RequestedStop = true

					route.Merge(upcoming)

					// move on to the next leg
					current++

					if current < len(variant) {
						now = next
						next = graph.Get(variant[current])
					}
				}

				routes <- &route
				track.Done()
			}()

			track.Add(2) // one for computation, one for comparison
		}

		// Close the routes channel when all results are in.
		go func() {
			track.Wait()
			close(routes)
		}()

		var route *structs.SpaceRoute
		for result := range routes {
			if route == nil {
				route = result
			} else if result.Distance < route.Distance {
				route = result
			}

			track.Done()
		}

		// Wait until all variants have been tested and compared against each other. Once that's
		// done `route` should be the shortest route to reach all provided points.
		track.Wait()

		// Choose the best of the available routes
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

	/**
	 * Secondary route: used for autocompleting system names.
	 */
	router.GET("/search", func(ctx *gin.Context) {
		query := ctx.Query("q")

		// Find terms
		ctx.JSON(http.StatusOK, terms.GetAll(query, 5))
	})

	router.Run()
}

// TODO: test this
func getVariants(start structs.SystemID, end structs.SystemID, visit []structs.SystemID) [][]structs.SystemID {
	var variants [][]structs.SystemID

	// If no visits are provided, start and end are required.
	if len(visit) == 0 {
		variants = append(variants, []structs.SystemID{start, end})
		return variants
	}

	gen, _ := permutation.NewPerm(visit, nil)

	for gen.Left() > 0 {
		next, _ := gen.Next()
		nextIds := next.([]structs.SystemID)

		if start != 0 {
			nextIds = append([]structs.SystemID{start}, nextIds...)
		}

		if end != 0 {
			nextIds = append(nextIds, end)
		}

		variants = append(variants, nextIds)
	}

	return variants
}
