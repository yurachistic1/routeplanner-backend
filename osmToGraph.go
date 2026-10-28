package routeplanner

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
				Id:       routing.Id(node.Id),
				Lat:      node.Lat,
				Lon:      node.Lon,
				Adjacent: []routing.Id{},
				Edges:    make(map[routing.Id]routing.Edge)}
	}

	// connect all the nodes
	for _, way := range elementsGrouped.Ways {
		if len(way.Nodes) > 1 {
			for i := 0; i < len(way.Nodes)-1; i += 1 {
				n1, ok1 := graph[routing.Id(way.Nodes[i])]
				n2, ok2 := graph[routing.Id(way.Nodes[i+1])]

				// skip nodes missing from the response and self loops
				if !ok1 || !ok2 || n1.Id == n2.Id {
					continue
				}

				connect(n1, n2)
			}
		}
	}

	graph.RemoveDeadEnds()
	return graph
}

// connect adds edges both ways between two nodes, unless they are already
// connected by another way.
func connect(n1, n2 *routing.Node) {
	if _, exists := n1.Edges[n2.Id]; exists {
		return
	}

	distance := routing.Haversine(n1, n2)

	n1.Adjacent = append(n1.Adjacent, n2.Id)
	n1.Edges[n2.Id] =
		routing.Edge{Distance: distance, Bearing: routing.Bearing(n1, n2)}

	n2.Adjacent = append(n2.Adjacent, n1.Id)
	n2.Edges[n1.Id] =
		routing.Edge{Distance: distance, Bearing: routing.Bearing(n2, n1)}
}
