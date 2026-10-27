package routeplanner

import (
	"testing"

	"github.com/yurachistic1/routeplanner-backend/overpass"
	"github.com/yurachistic1/routeplanner-backend/routing"
)

func TestOSMToGraph(t *testing.T) {

	node := func(id int, lat, lon float64) overpass.Element {
		return overpass.Element{Type: "node", Id: id, Lat: lat, Lon: lon}
	}
	way := func(nodes ...int) overpass.Element {
		return overpass.Element{Type: "way", Nodes: nodes}
	}

	// a square loop 1-2-3-4, with segment 1-2 also part of a second way, a
	// self loop and a way referencing node 99 which is missing from the response
	res := overpass.Response{Elements: []overpass.Element{
		node(1, 51.500, -0.100),
		node(2, 51.501, -0.100),
		node(3, 51.501, -0.099),
		node(4, 51.500, -0.099),
		way(1, 2, 3, 4, 1),
		way(2, 1),
		way(3, 3),
		way(4, 99),
	}}

	graph := OSMToGraph(res)

	if len(graph) != 4 {
		t.Fatalf("OSMToGraph returned %d nodes, want 4:\n%v", len(graph), graph)
	}

	for id, n := range graph {
		if len(n.Adjacent) != 2 || len(n.Edges) != 2 {
			t.Errorf("node %d has adjacent %v and %d edges, want 2 of each",
				id, n.Adjacent, len(n.Edges))
		}
	}

	if d := graph[routing.Id(1)].Edges[routing.Id(2)].Distance; d < 100 || d > 120 {
		t.Errorf("distance between nodes 1 and 2 == %v, want about 111m", d)
	}
}
