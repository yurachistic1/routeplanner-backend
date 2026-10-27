package routing

import (
	"testing"
)

func TestPickAlongBearing(t *testing.T) {

	nodeA :=
		Node{1, 51.5307698, -0.1461484, []Id{2, 3},
			map[Id]Edge{2: {Bearing: 165}, 3: {Bearing: 344}}}

	nodeB :=
		Node{1, 51.5307698, -0.1461484, []Id{2, 3, 4, 5},
			map[Id]Edge{
				2: {Bearing: 165},
				3: {Bearing: 344},
				4: {Bearing: 12},
				5: {Bearing: 300},
			}}

	cases := []struct {
		in      Node
		target  float64
		exclude Id
		want    Id
	}{
		{nodeA, 0, -1, 3},
		{nodeA, 0, 3, 2},
		{nodeB, 330, -1, 3},
		{nodeB, 330, 3, 5},
	}

	for _, c := range cases {

		result := pickAlongBearing(c.target, c.in.Edges, c.exclude)

		if result != c.want {
			t.Errorf("pickAlongBearing(%v, %v, %v) == %v, want %v",
				c.target, c.in.Edges, c.exclude, result, c.want)
		}
	}
}

func TestClosestNodes(t *testing.T) {

	// graph of nodes 1..size spaced out northwards from the target
	graphOfSize := func(size int) Graph {
		g := make(Graph)
		for i := 1; i <= size; i++ {
			g[Id(i)] = &Node{Id: Id(i), Lat: 51.5 + float64(i)*0.001, Lon: -0.1}
		}
		return g
	}

	cases := []struct {
		size int
		want []Id
	}{
		{0, []Id{}},
		{2, []Id{1, 2}},
		{5, []Id{1, 2, 3}},
		{11, []Id{1, 6, 11}},
		{20, []Id{1, 6, 11}},
	}

	for _, c := range cases {

		result := ClosestNodes(51.5, -0.1, graphOfSize(c.size), 3)

		ids := []Id{}
		for _, node := range result {
			ids = append(ids, node.Id)
		}

		if len(ids) != len(c.want) {
			t.Errorf("ClosestNodes on graph of size %d == %v, want %v", c.size, ids, c.want)
			continue
		}
		for i := range ids {
			if ids[i] != c.want[i] {
				t.Errorf("ClosestNodes on graph of size %d == %v, want %v", c.size, ids, c.want)
				break
			}
		}
	}
}
