// Package routing provides functions to navigate on and process
// graphs in the context of routeplanner project.
package routing

import (
	"math"
	"math/rand"
	"sort"
)

// TopRoutes returns a slice of routes that are considered the best fit for the
// supplied criteria such as distance as well as implicit criteria such as No of
// turns and others. Lots of possible ones are generated and then best 25 are
// selected and returned.
func TopRoutes(lat, lon, distance float64, graph Graph) Routes {

	nodes := ClosestNodes(lat, lon, graph, 3)

	top := make(Routes, 0, 25)

	for _, start := range nodes {

		for i := 0; i < 360; i += 20 {

			for j := 0; j < 50; j++ {
				r1 := createRoute(start, distance, float64(i), graph, Clockwise)
				top = appendRoute(r1, top)
				r2 := createRoute(start, distance, float64(i), graph, Anticlockwise)
				top = appendRoute(r2, top)
			}
		}
	}

	for i, r := range top {
		top[i] = completeRoute(r, graph)
	}

	sort.Sort(top)

	return top
}

// AppendRoute is a custom append function for Routes type that keeps the slice
// ordered as well as attempting to keep all elements sufficiently distinct.
// AppendRoute does not allow exceeding the capacity of the original slice.
func appendRoute(route Route, routes Routes) Routes {

	if len(routes) == 0 {
		return append(routes, route)
	}

	if len(routes) == cap(routes) {
		lastTwo := Routes{routes[len(routes)-1], route}

		if lastTwo.Less(0, 1) {
			return routes
		}
	}

	for i := 0; i < len(routes); i++ {
		similarity := routeSimilarity(route, routes[i])

		if similarity > 70 {
			pair := Routes{route, routes[i]}
			if pair.Less(0, 1) {
				routes[i] = route

			}

			return routes
		}
	}

	if len(routes) < cap(routes) {
		routes = append(routes, route)
	} else {
		routes[len(routes)-1] = route
	}

	for i := len(routes) - 2; i >= 0; i-- {
		pair := Routes{routes[i], routes[i+1]}

		if pair.Less(0, 1) {
			return routes
		} else {
			routes.Swap(i, i+1)
		}

	}

	return routes
}

// Create route returns a circular Route of desired distance at a specified start location.
func createRoute(start *Node, distance, initBearing float64, g Graph, rot Rotation) Route {

	route := Route{
		Path:          make([]*Node, 1, 1000),
		Length:        0,
		DesiredLength: distance,
		Visited:       make(map[Id]int),
		RepeatVisits:  0,
		Turns:         0,
	}

	route.Path[0] = start

	var b float64 = initBearing
	var radius float64 = distance / (2 * math.Pi)

	// vars to figure out if next node counts as a turn or not
	var currentBearing float64 = initBearing
	var newBearing float64

	var previousNode *Node = &Node{Id: -1}

	for route.Length < distance*0.98 {
		currentNode := route.Path[len(route.Path)-1]

		route.Visited[currentNode.Id] += 1

		if route.Visited[currentNode.Id] > 1 {
			route.RepeatVisits += 1
		}

		if len(route.Path) > 1 {
			previousNode = route.Path[len(route.Path)-2]
			currentBearing = previousNode.Edges[currentNode.Id].Bearing

		}

		steer := pickAlongBearing(b, currentNode.Edges, previousNode.Id)
		straight := pickAlongBearing(currentBearing, currentNode.Edges, previousNode.Id)

		pick := 0
		choices := []Id{steer, straight}

		if straight != steer {
			pick = rand.Intn(2)

		}

		route.Path = append(route.Path, (g)[choices[pick]])
		route.Length += currentNode.Edges[choices[pick]].Distance

		newBearing = (g)[currentNode.Id].Edges[choices[pick]].Bearing

		if len(route.Path) > 1 && bearingDifference(currentBearing, newBearing) > 45 {
			route.Turns++
		}

		switch rot {
		case Clockwise:
			b = math.Mod(sectorAngle(route.Length, radius)+initBearing, 360)
		case Anticlockwise:
			b = math.Mod((-sectorAngle(route.Length, radius)+initBearing)+360, 360)
		}
	}

	return route
}

// PickAlongBearing selects a an edge (connected node id) that has the closest bearing
// to the target bearing.
func pickAlongBearing(target float64, vals map[Id]Edge, exclude Id) (closest Id) {

	minDifference := math.MaxFloat64

	for key, val := range vals {

		if key == exclude {
			continue
		}

		difference := bearingDifference(target, val.Bearing)
		if difference < minDifference {
			closest = key
			minDifference = difference
		}
	}

	return
}

// ClosestNode returns n closest node pointers to a given lat and lon coordinates.
func ClosestNodes(lat float64, lon float64, g Graph, n int) (closest []*Node) {

	closest = []*Node{}

	target := Node{Lat: lat, Lon: lon}
	pairs := sortByD{}

	for _, val := range g {
		distance := Haversine(&target, val)

		pairs = append(pairs, pair{val, distance})
	}

	sort.Sort(pairs)

	for i := 0; i < n; i++ {
		closest = append(closest, pairs[i*5].n)
	}
	return closest
}

type pair struct {
	n *Node
	d float64
}

type sortByD []pair

func (a sortByD) Len() int           { return len(a) }
func (a sortByD) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByD) Less(i, j int) bool { return a[i].d < a[j].d }
