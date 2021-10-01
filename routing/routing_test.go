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
