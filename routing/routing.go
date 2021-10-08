package routing

import (
	"fmt"
	"math"
	"math/rand"
)

// Create route returns a circular Route of desired distance at a specified start location.
func CreateRoute(start *Node, distance, initBearing float64, g Graph, rot Rotation) Route {

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
