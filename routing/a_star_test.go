package routing

import (
	"math"
	"testing"
)

// gridGraph returns a size x size grid of nodes roughly 100m apart, with ids
// row*size+col+1. Edges for which skip returns true are left out.
func gridGraph(size int, skip func(a, b Id) bool) Graph {
	g := make(Graph)
	id := func(row, col int) Id { return Id(row*size + col + 1) }

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			g[id(row, col)] = &Node{
				Id:       id(row, col),
				Lat:      51.5 + float64(row)*0.0009,
				Lon:      -0.1 + float64(col)*0.0014,
				Adjacent: []Id{},
				Edges:    make(map[Id]Edge),
			}
		}
	}

	connect := func(a, b Id) {
		if skip(a, b) {
			return
		}
		d := Haversine(g[a], g[b])
		g[a].Adjacent = append(g[a].Adjacent, b)
		g[a].Edges[b] = Edge{Distance: d}
		g[b].Adjacent = append(g[b].Adjacent, a)
		g[b].Edges[a] = Edge{Distance: d}
	}

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if col+1 < size {
				connect(id(row, col), id(row, col+1))
			}
			if row+1 < size {
				connect(id(row, col), id(row+1, col))
			}
		}
	}

	return g
}

// shortestDistance is a plain Dijkstra used to check aStar against.
func shortestDistance(g Graph, from, to Id) float64 {
	dist := map[Id]float64{from: 0}
	done := map[Id]bool{}

	for {
		current, best := Id(0), math.Inf(1)
		for id, d := range dist {
			if !done[id] && d < best {
				current, best = id, d
			}
		}
		if math.IsInf(best, 1) || current == to {
			return best
		}
		done[current] = true

		for _, id := range g[current].Adjacent {
			if d := best + g[current].Edges[id].Distance; d < getWithDefault(dist, id, math.Inf(1)) {
				dist[id] = d
			}
		}
	}
}

func TestAStar(t *testing.T) {
	const size = 6

	// wall between columns 2 and 3 open only on the top row, so the path
	// from bottom left to bottom right has to go around it
	wall := func(a, b Id) bool {
		colA, colB := int(a-1)%size, int(b-1)%size
		row := int(a-1) / size
		return colA == 2 && colB == 3 && row != size-1
	}

	cases := []struct {
		name     string
		skip     func(a, b Id) bool
		from, to Id
	}{
		{"open grid", func(a, b Id) bool { return false }, 1, size * size},
		{"around wall", wall, 1, size},
		{"same node", func(a, b Id) bool { return false }, 8, 8},
	}

	for _, c := range cases {
		g := gridGraph(size, c.skip)

		// aStar searches from the last node in the path back to the first
		result := aStar([]*Node{g[c.to], g[c.from]}, g)

		if len(result) == 0 || result[0] != c.from || result[len(result)-1] != c.to {
			t.Errorf("%s: aStar path %v does not run from %d to %d", c.name, result, c.from, c.to)
			continue
		}

		length := 0.0
		for i := 0; i < len(result)-1; i++ {
			edge, ok := g[result[i]].Edges[result[i+1]]
			if !ok {
				t.Fatalf("%s: aStar path %v uses missing edge %d-%d", c.name, result, result[i], result[i+1])
			}
			length += edge.Distance
		}

		if want := shortestDistance(g, c.from, c.to); math.Abs(length-want) > 1e-6 {
			t.Errorf("%s: aStar path length %.1fm, want %.1fm", c.name, length, want)
		}
	}
}

func TestAStarUnreachable(t *testing.T) {
	// cut the grid in half between columns 2 and 3
	g := gridGraph(4, func(a, b Id) bool {
		return int(a-1)%4 == 1 && int(b-1)%4 == 2
	})

	if result := aStar([]*Node{g[4], g[1]}, g); len(result) != 0 {
		t.Errorf("aStar across disconnected graph == %v, want []", result)
	}
}
