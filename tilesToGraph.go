package routeplanner

import (
	"github.com/yurachistic1/routeplanner-backend/routing"
	"github.com/yurachistic1/routeplanner-backend/tiles"
)

// TilesToGraph takes tiles and returns a graph of the edges with at least one
// node inside the bounding box, and without dead ends.
func TilesToGraph(ts []*tiles.Tile, south, west, north, east float64) (graph routing.Graph) {

	graph = make(routing.Graph)
	coords, edges := tiles.Merge(ts)

	inside := func(c [2]float64) bool {
		return c[0] >= south && c[0] <= north && c[1] >= west && c[1] <= east
	}

	node := func(id int64) *routing.Node {
		n, ok := graph[routing.Id(id)]
		if !ok {
			n = &routing.Node{
				Id:       routing.Id(id),
				Lat:      coords[id][0],
				Lon:      coords[id][1],
				Adjacent: []routing.Id{},
				Edges:    make(map[routing.Id]routing.Edge)}
			graph[n.Id] = n
		}
		return n
	}

	for _, e := range edges {
		if e.A == e.B || (!inside(coords[e.A]) && !inside(coords[e.B])) {
			continue
		}
		connect(node(e.A), node(e.B))
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
