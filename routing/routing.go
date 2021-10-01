package routing

import (
	"fmt"
	"math"
)

func CreateRoute(lat, lon, distance, initBearing float64, g Graph) Route {
	start := closestNode(lat, lon, g)

	route := Route{Path: []*Node{g[start]}, Length: 0, Turns: 0}

	var b float64 = initBearing
	var r float64 = distance / (2 * math.Pi)

	// vars to figure out if next node counts as a turn or not
	var currentBearing float64
	var newBearing float64

	for route.Length < distance {
		previousNode := Id(-1)
		currentNode := route.Path[len(route.Path)-1]

		if len(route.Path) > 1 {
			previousNode = route.Path[len(route.Path)-2].Id
			currentBearing = g[previousNode].Edges[currentNode.Id].Bearing

		}

		nextId := pickAlongBearing(b, currentNode.Edges, previousNode)

		route.Path = append(route.Path, g[nextId])
		route.Length += currentNode.Edges[nextId].Distance

		newBearing = g[currentNode.Id].Edges[nextId].Bearing

		if len(route.Path) > 1 && bearingDifference(currentBearing, newBearing) > 80 {
			route.Turns++
		}

		b = math.Mod(sectorAngle(route.Length, r)+initBearing, 360)
	}

	return route
}

// PickAlongBearing selects a an edge (connected node id) that has the closest bearing
// to the target bearing and returns a boolean indicating whether it was a turn or not.
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

// ClosestNode returns a node id with the shortest distance to given lat and lon coordinates.
func closestNode(lat float64, lon float64, g Graph) Id {

	target := Node{Lat: lat, Lon: lon}
	minDistance := math.MaxFloat64
	closest := Id(-1)

	for key, val := range g {
		distance := haversine(&target, val)

		if distance < minDistance {
			minDistance = distance
			closest = key
		}
	}
	return closest
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
