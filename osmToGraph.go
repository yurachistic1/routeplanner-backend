package main

import (
	"github.com/yurachistic1/routeplanner-backend/overpass"
	"github.com/yurachistic1/routeplanner-backend/routing"
)

// osmToGraph takes an overpass Responce object and returns a graph of nodes and
// edges between them without dead ends.
func OSMToGraph(res overpass.Response) (graph routing.Graph) {

	graph = make(routing.Graph)
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

	// insert all the nodes into the Graph
	for _, node := range elementsGrouped.Nodes {
		graph[routing.Id(node.Id)] =
			&routing.Node{
				routing.Id(node.Id),
				node.Lat, node.Lon,
				[]routing.Id{},
				make(map[routing.Id]routing.Edge)}
	}

	// connect all the nodes
	for _, way := range elementsGrouped.Ways {
		if len(way.Nodes) > 1 {
			for i := 0; i < len(way.Nodes)-1; i += 1 {
				var n1 *routing.Node = graph[routing.Id(way.Nodes[i])]
				var n2 *routing.Node = graph[routing.Id(way.Nodes[i+1])]

				graph[n1.Id].Adjacent = append(graph[n1.Id].Adjacent, n2.Id)
				graph[n1.Id].Edges[n2.Id] =
					routing.Edge{routing.Haversine(n1, n2), routing.Bearing(n1, n2)}

				graph[n2.Id].Adjacent = append(graph[n2.Id].Adjacent, n1.Id)
				graph[n2.Id].Edges[n1.Id] =
					routing.Edge{routing.Haversine(n1, n2), routing.Bearing(n2, n1)}

			}
		}
	}

	graph.RemoveDeadEnds()
	return graph
}
