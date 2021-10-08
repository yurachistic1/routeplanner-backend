package routing

import (
	"fmt"
)

// Rotation enum type
type Rotation int

// Rotation option
const (
	Clockwise Rotation = iota
	Anticlockwise
)

// Unique id identifying nodes
type Id int

// Map storing all the nodes
type Graph map[Id]*Node

func (g Graph) String() string {
	str := ""

	for key, elem := range g {
		str += fmt.Sprintf("%d: %+v\n", key, *elem)
	}

	return str

}

// removeDeadEnds removes all the nodes connected to 1 or fewer other nodes
// and repeats that process until all remaining nodes are connected to 2 or more nodes.
func (graph *Graph) RemoveDeadEnds() {

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

// ToPolyline converts graph to code that can be used to draw a polyline with leaflet js
// library. For testing purposes only.
func (graph *Graph) ToPolyline() string {

	str := "var latlngs = [\n"

	for _, val := range *graph {
		for _, id := range val.Adjacent {
			coordpair := fmt.Sprintf("[[%f, %f], [%f, %f]],\n",
				val.Lat, val.Lon, (*graph)[id].Lat, (*graph)[id].Lon)
			str += coordpair
		}
	}
	return str + "]"
}

// Node represents a vertex with latitude and longitude and stores a list of
// edges connected to it.
type Node struct {
	Id  Id
	Lat float64
	Lon float64

	Adjacent []Id
	Edges    map[Id]Edge
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

// Edge stores infomation on distance and and bearing between two nodes.
type Edge struct {
	Distance float64
	Bearing  float64
}

type Route struct {
	Path   []*Node
	Length float64
	Turns  int
}

type Routes []Route

func (routes Routes) Len() int {
	return len(routes)
}

func (routes Routes) Swap(i, j int) {
	routes[i], routes[j] = routes[j], routes[i]
}

func (routes Routes) Less(i, j int) bool {

	//return routes[i].Turns < routes[j].Turns

	turnsI := routes[i].Turns
	dFromStartI := Haversine(routes[i].Path[0], routes[i].Path[len(routes[i].Path)-1])

	turnsJ := routes[j].Turns
	dFromStartJ := Haversine(routes[j].Path[0], routes[j].Path[len(routes[j].Path)-1])

	return (float64(turnsI)*50 + dFromStartI) < (float64(turnsJ)*50 + dFromStartJ)
}
