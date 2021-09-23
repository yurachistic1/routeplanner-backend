package routing

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
