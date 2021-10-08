package routing

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/yurachistic1/routeplanner-backend/overpass"
)

func TopRoutes(lat, lon, distance float64, res overpass.Response) Routes {

	graph := OSMToGraph(res)

	start := ClosestNode(lat, lon, graph)

	top25 := make(Routes, 0, 25)

	for i := 0; i < 360; i += 15 {

		for j := 0; j < 50; j++ {
			r1 := createRoute(start, distance, float64(i), graph, Clockwise)
			top25 = appendRoute(r1, top25)
			r2 := createRoute(start, distance, float64(i), graph, Anticlockwise)
			top25 = appendRoute(r2, top25)
		}
	}

	return top25
}

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

		if similarity > 80 {
			pair := Routes{routes[i], route}
			if pair.Less(1, 0) {
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

	route := Route{Path: make([]*Node, 1, 1000), Length: 0, Turns: 0}
	route.Path[0] = start

	var b float64 = initBearing
	var radius float64 = distance / (2 * math.Pi)

	// vars to figure out if next node counts as a turn or not
	var currentBearing float64 = initBearing
	var newBearing float64

	var previousNode *Node = &Node{Id: -1}

	for route.Length < distance {
		currentNode := route.Path[len(route.Path)-1]

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

		if len(route.Path) > 1 && bearingDifference(currentBearing, newBearing) > 80 {
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

func (r *Route) ToPolyline() string {
	str := "var latlngs2 = [\n"

	for _, node := range r.Path {
		coordpair := fmt.Sprintf("[%f, %f],\n",
			node.Lat, node.Lon)
		str += coordpair
	}

	return str + "]"
}

// ClosestNode returns a node pointer with the shortest distance to given lat and lon coordinates.
func ClosestNode(lat float64, lon float64, g Graph) (closest *Node) {

	target := Node{Lat: lat, Lon: lon}
	minDistance := math.MaxFloat64

	for _, val := range g {
		distance := haversine(&target, val)

		if distance < minDistance {
			minDistance = distance
			closest = val
		}
	}
	return closest
}
