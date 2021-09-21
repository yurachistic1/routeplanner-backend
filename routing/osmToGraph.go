// Package routing provides functions to navigate on and process
// graphs in the context of routeplanner project.
package routing

import (
	"math"

	"github.com/yurachistic1/routeplanner-backend/overpass"
)

// Unique id identifying nodes
type Id int

// Map storing all the nodes
type Graph map[Id]*Node

// Node represents a vertex with latitude and longitude and stores a list of
// edges connected to it.
type Node struct {
	Id  Id
	Lat float64
	Lon float64

	Adjacent []Id
	Edges    map[Id]Edge
}

// Edge stores infomation on distance and and bearing between two nodes.
type Edge struct {
	distance float64
	bearing  float64
}

// removeEdge deletes a target id from a list of adjacent nodes stored
// in the node.Adjacent property as well as from node.Edges map.
func (node *Node) removeEdge(target Id) {
	l := node.Adjacent
	for i, val := range l {
		if val == target {
			l[i] = l[len(l)-1]
			node.Adjacent = l[:len(l)-1]
			delete(node.Edges, target)
			return
		}
	}
}

// Haversine calculates haversine distance between two nodes.
func haversine(n1, n2 *Node) float64 {
	const earthRadius = 6371000 // meters
	// lat lon in radians
	lat1Rad := n1.Lat * math.Pi / 180
	lat2Rad := n2.Lat * math.Pi / 180
	lon1Rad := n1.Lon * math.Pi / 180
	lon2Rad := n2.Lon * math.Pi / 180
	x := (lon2Rad - lon1Rad) * math.Cos((lat1Rad+lat2Rad)/2)
	y := (lat2Rad - lat1Rad)
	d := math.Sqrt(x*x+y*y) * earthRadius

	return d
}

// Bearing calculates bearing in degrees between two nodes.
func bearing(n1, n2 *Node) float64 {
	// lat lon in radians
	lat1Rad := n1.Lat * math.Pi / 180
	lat2Rad := n2.Lat * math.Pi / 180
	lon1Rad := n1.Lon * math.Pi / 180
	lon2Rad := n2.Lon * math.Pi / 180

	y := math.Sin(lon2Rad-lon1Rad) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(lon2Rad-lon1Rad)

	bearingRad := math.Atan2(y, x)
	bearingDeg := math.Mod((bearingRad*180/math.Pi + 360), 360) // in degrees

	return bearingDeg
}

// removeDeadEnds removes all the nodes connected to 1 or fewer other nodes
// and repeats that process until all remaining nodes are connected to 2 or more nodes.
func (graph *Graph) removeDeadEnds() {

	for id, node := range *graph {
		switch len(node.Adjacent) {
		case 0:
			delete(*graph, id)
		case 1:
			delete(*graph, id)
			next := (*graph)[node.Adjacent[0]]
			for {
				numEdges := len(next.Adjacent)
				if numEdges == 1 {
					delete(*graph, next.Id)
					break
				} else if numEdges == 2 {
					delete(*graph, next.Id)
					next.removeEdge(id)
					id = next.Id
					next = (*graph)[next.Adjacent[0]]
				} else {
					next.removeEdge(id)
					break
				}
			}
		}
	}
}

// osmToGraph takes an overpass Responce object and returns a graph of nodes and
// edges between them.
func osmToGraph(res overpass.Response) (graph Graph) {

	graph = make(Graph)
	elementsGrouped := struct {
		Nodes []overpass.Element
		Ways  []overpass.Element
	}{[]overpass.Element{}, []overpass.Element{}}

	for _, element := range res.Elements {
		switch element.Type {
		case "way":
			elementsGrouped.Ways = append(elementsGrouped.Ways, element)
		case "node":
			elementsGrouped.Nodes = append(elementsGrouped.Nodes, element)
		}
	}

	for _, node := range elementsGrouped.Nodes {
		graph[Id(node.Id)] =
			&Node{Id(node.Id), node.Lat, node.Lon, []Id{}, make(map[Id]Edge)}
	}

	for _, way := range elementsGrouped.Ways {
		if len(way.Nodes) > 1 {
			for i := 0; i < len(way.Nodes)-1; i += 2 {
				var n1 *Node = graph[Id(way.Nodes[i])]
				var n2 *Node = graph[Id(way.Nodes[i+1])]

				graph[n1.Id].Adjacent = append(graph[n1.Id].Adjacent, n2.Id)
				graph[n1.Id].Edges[n2.Id] = Edge{haversine(n1, n2), bearing(n1, n2)}

				graph[n2.Id].Adjacent = append(graph[n1.Id].Adjacent, n1.Id)
				graph[n2.Id].Edges[n1.Id] = Edge{haversine(n1, n2), bearing(n2, n1)}

			}
		}
	}
	return graph
}
