package routing

import "fmt"

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

func (g Graph) String() string {
	str := ""

	for key, elem := range g {
		str += fmt.Sprintf("%d: %+v\n", key, *elem)
	}

	return str

}
