package routing

import (
	"fmt"
	"math"
)

func CreateRoute(lat, lon, distance float64, g Graph) Route {
	start := closestNode(lat, lon, g)

	fmt.Println(start)

	route := Route{Path: []*Node{g[start]}, Length: 0, Turns: 0}

	var b float64 = 0
	var r float64 = distance / (2 * math.Pi)

	for route.Length < distance {
		previousNode := Id(-1)
		currentNode := route.Path[len(route.Path)-1]

		fmt.Printf("%+v\n", currentNode)

		if len(route.Path) > 1 {
			previousNode = route.Path[len(route.Path)-2].Id
		}

		nextId, _ := pickAlongBearing(b, currentNode.Edges, previousNode)

		route.Path = append(route.Path, g[nextId])
		route.Length += currentNode.Edges[nextId].Distance

		fmt.Println(route.Length, r)
		fmt.Println(sectorAngle(route.Length, r))
		b = math.Mod(sectorAngle(route.Length, r), 360)
	}

	return route
}

// PickAlongBearing selects a an edge (connected node id) that has the closest bearing
// to the target bearing and returns a boolean indicating whether it was a turn or not.
func pickAlongBearing(target float64, vals map[Id]Edge, exclude Id) (closest Id, turned bool) {

	minDifference := math.MaxFloat64

	for key, val := range vals {

		if key == exclude {
			continue
		}

		difference := bearingDifference(target, val.Bearing)
		if difference < minDifference {
			closest = key
			minDifference = difference
			turned = difference > 80
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
