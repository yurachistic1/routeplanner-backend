// Package routing provides functions to navigate on and process
// graphs in the context of routeplanner project.
package routing

import (
	"fmt"

	"github.com/yurachistic1/routeplanner-backend/overpass"
)

// osmToGraph takes an overpass Responce object and returns a graph of nodes and
// edges between them without dead ends.
func OSMToGraph(res overpass.Response) (graph Graph) {

	graph = make(Graph)
	elementsGrouped := struct {
		Nodes []overpass.Element
		Ways  []overpass.Element
	}{[]overpass.Element{}, []overpass.Element{}}

	// group elements
	for _, element := range res.Elements {
		switch element.Type {
		case "way":
			elementsGrouped.Ways = append(elementsGrouped.Ways, element)
		case "node":
			elementsGrouped.Nodes = append(elementsGrouped.Nodes, element)
		}
	}

	// insert all the nodes into the graph
	for _, node := range elementsGrouped.Nodes {
		graph[Id(node.Id)] =
			&Node{Id(node.Id), node.Lat, node.Lon, []Id{}, make(map[Id]Edge)}
	}

	// connect all the nodes
	for _, way := range elementsGrouped.Ways {
		if len(way.Nodes) > 1 {
			for i := 0; i < len(way.Nodes)-1; i += 1 {
				var n1 *Node = graph[Id(way.Nodes[i])]
				var n2 *Node = graph[Id(way.Nodes[i+1])]

				graph[n1.Id].Adjacent = append(graph[n1.Id].Adjacent, n2.Id)
				graph[n1.Id].Edges[n2.Id] = Edge{haversine(n1, n2), bearing(n1, n2)}

				graph[n2.Id].Adjacent = append(graph[n2.Id].Adjacent, n1.Id)
				graph[n2.Id].Edges[n1.Id] = Edge{haversine(n1, n2), bearing(n2, n1)}

			}
		}
	}

	graph.removeDeadEnds()
	return graph
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
