package routeplanner

import (
	"math"
	"sort"
	"testing"

	"github.com/yurachistic1/routeplanner-backend/routing"
	"github.com/yurachistic1/routeplanner-backend/tiles"
	"github.com/yurachistic1/routeplanner-backend/walkfilter"
)

// gridTiles returns tiles for a size x size grid of nodes about 100m apart
// near Regent's Park, spanning the boundary between tiles 12/2045 and 12/2046,
// with a dead end spur off one corner. Node ids are row*100+col+1.
func gridTiles(size int) ([]*tiles.Tile, map[int64][2]float64) {
	coords := make(map[int64][2]float64)
	var edges []walkfilter.Edge
	id := func(row, col int) int64 { return int64(row*100 + col + 1) }

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			coords[id(row, col)] = [2]float64{51.5300 + float64(row)*0.0009, -0.1800 + float64(col)*0.0014}
			if col > 0 {
				edges = append(edges, walkfilter.Edge{A: id(row, col-1), B: id(row, col), Class: walkfilter.Preferred})
			}
			if row > 0 {
				edges = append(edges, walkfilter.Edge{A: id(row-1, col), B: id(row, col), Class: walkfilter.OK})
			}
		}
	}

	// spur 99998-99999 off the corner node 1
	coords[99998] = [2]float64{51.5290, -0.1800}
	coords[99999] = [2]float64{51.5280, -0.1800}
	edges = append(edges, walkfilter.Edge{A: 1, B: 99998}, walkfilter.Edge{A: 99998, B: 99999})

	return tiles.Split(edges, coords, tiles.Zoom), coords
}

func TestTilesToGraph(t *testing.T) {
	ts, coords := gridTiles(5)
	if len(ts) < 2 {
		t.Fatalf("test grid is in %d tiles, want it to cross a tile boundary", len(ts))
	}

	graph := TilesToGraph(ts, -90, -180, 90, 180)

	// all grid nodes, without the spur
	if len(graph) != 25 {
		t.Errorf("graph has %d nodes, want 25", len(graph))
	}
	for _, id := range []routing.Id{99998, 99999} {
		if _, ok := graph[id]; ok {
			t.Errorf("dead end node %d is still in the graph", id)
		}
	}

	// edges are joined up across the tile boundary and match OSMToGraph:
	// both directions, with distance and bearing
	for id, n := range graph {
		c := coords[int64(id)]
		if math.Abs(n.Lat-c[0]) > 1e-7 || math.Abs(n.Lon-c[1]) > 1e-7 {
			t.Errorf("node %d at %v,%v, want %v", id, n.Lat, n.Lon, c)
		}
		if len(n.Adjacent) != len(n.Edges) {
			t.Errorf("node %d has %d adjacent ids but %d edges", id, len(n.Adjacent), len(n.Edges))
		}
		for other, e := range n.Edges {
			back, ok := graph[other].Edges[id]
			if !ok {
				t.Errorf("edge %d-%d has no reverse", id, other)
				continue
			}
			if want := routing.Haversine(n, graph[other]); math.Abs(e.Distance-want) > 1e-9 || back.Distance != e.Distance {
				t.Errorf("edge %d-%d is %vm, want %vm both ways", id, other, e.Distance, want)
			}
			if want := routing.Bearing(n, graph[other]); math.Abs(e.Bearing-want) > 1e-9 {
				t.Errorf("edge %d-%d bearing %v, want %v", id, other, e.Bearing, want)
			}
		}
	}

	// interior nodes of a 5x5 grid have 4 neighbours, corners 2
	if got := len(graph[routing.Id(1*100+1+1)].Adjacent); got != 4 {
		t.Errorf("interior node has %d neighbours, want 4", got)
	}
	if got := len(graph[1].Adjacent); got != 2 {
		t.Errorf("corner node has %d neighbours, want 2 after removing the spur", got)
	}
}

func TestTilesToGraphCrops(t *testing.T) {
	ts, coords := gridTiles(5)

	// a box around the first 3 columns: edges from column 3 into column 4
	// are kept, but column 4 has no edges between its own nodes, as both
	// their ends are outside, so those become dead ends and are removed
	graph := TilesToGraph(ts, 51.5299, -0.1801, 51.5340, -0.1800+2*0.0014+0.0001)

	var cols []int
	seen := map[int]bool{}
	for id := range graph {
		col := int(id-1) % 100
		if !seen[col] {
			seen[col] = true
			cols = append(cols, col)
		}
	}
	sort.Ints(cols)

	if len(cols) != 3 || cols[2] != 2 {
		t.Errorf("cropped graph has columns %v, want 0-2", cols)
	}
	if _, ok := graph[routing.Id(4+1)]; ok {
		t.Errorf("node %v outside the box with no edge into it is in the graph", coords[5])
	}
}

func TestTilesToGraphEmpty(t *testing.T) {
	if graph := TilesToGraph(nil, 51, -1, 52, 0); len(graph) != 0 {
		t.Errorf("graph from no tiles has %d nodes, want 0", len(graph))
	}
}
