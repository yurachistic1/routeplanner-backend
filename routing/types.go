package routing

import (
	"fmt"
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
	Distance float64
	Bearing  float64
}

type Route struct {
	Path   []*Node
	Length float64
	Turns  int
}

type Routes []Route

func (g Graph) String() string {
	str := ""

	for key, elem := range g {
		str += fmt.Sprintf("%d: %+v\n", key, *elem)
	}

	return str

}

func (routes Routes) Len() int {
	return len(routes)
}

func (routes Routes) Swap(i, j int) {
	routes[i], routes[j] = routes[j], routes[i]
}

func (routes Routes) Less(i, j int) bool {

	//return routes[i].Turns < routes[j].Turns

	turnsI := routes[i].Turns
	dFromStartI := haversine(routes[i].Path[0], routes[i].Path[len(routes[i].Path)-1])

	turnsJ := routes[j].Turns
	dFromStartJ := haversine(routes[j].Path[0], routes[j].Path[len(routes[j].Path)-1])

	return (float64(turnsI)*50 + dFromStartI) < (float64(turnsJ)*50 + dFromStartJ)
}
