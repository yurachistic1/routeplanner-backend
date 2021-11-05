package routing

import (
	"fmt"
	"math"
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

// Route type stores information describing a route such as ordered slice of nodes that
// route, its length, number of turns and others.
type Route struct {
	Path          []*Node
	Length        float64
	DesiredLength float64
	Visited       map[Id]int
	RepeatVisits  int
	Turns         int
}

type Routes []Route

func (routes Routes) Len() int {
	return len(routes)
}

func (routes Routes) Swap(i, j int) {
	routes[i], routes[j] = routes[j], routes[i]
}

func (routes Routes) Less(i, j int) bool {

	turnsI := routes[i].Turns * 30
	dFromStartI := Haversine(routes[i].Path[0], routes[i].Path[len(routes[i].Path)-1]) / 3
	distanceDiffI := math.Abs(routes[i].DesiredLength-routes[i].Length) / 3
	repeatsI := (routes[i].RepeatVisits * 10000) / len(routes[i].Path)

	iScore := dFromStartI + float64(turnsI) + float64(repeatsI) + distanceDiffI

	turnsJ := routes[j].Turns * 30
	dFromStartJ := Haversine(routes[j].Path[0], routes[j].Path[len(routes[j].Path)-1]) / 3
	distanceDiffJ := math.Abs(routes[j].DesiredLength-routes[j].Length) / 3
	repeatsJ := (routes[j].RepeatVisits * 10000) / len(routes[j].Path)

	jScore := dFromStartJ + float64(turnsJ) + float64(repeatsJ) + distanceDiffJ

	return iScore < jScore
}
