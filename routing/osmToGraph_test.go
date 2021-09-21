package routing

import (
	"reflect"
	"testing"
)

func TestRemoveEdge(t *testing.T) {
	cases := []struct {
		in     Node
		target Id
		want   Node
	}{
		{Node{1, 0, 0, []Id{2, 3}, map[Id]Edge{2: {}, 3: {}}}, 4, Node{1, 0, 0, []Id{2, 3}, map[Id]Edge{2: {}, 3: {}}}},
		{Node{1, 0, 0, []Id{2, 3}, map[Id]Edge{2: {}, 3: {}}}, 3, Node{1, 0, 0, []Id{2}, map[Id]Edge{2: {}}}},
		{Node{1, 0, 0, []Id{}, nil}, 1, Node{1, 0, 0, []Id{}, nil}},
	}

	for _, c := range cases {

		original := c.in

		c.in.removeEdge(c.target)

		if !reflect.DeepEqual(c.in, c.want) {
			t.Errorf("%v.removeEdge(%v) == %v, want %v", original, c.target, c.in, c.want)
		}
	}
}

func TestRemoveDeadEnds(t *testing.T) {
	cases := []struct {
		in   Graph
		want Graph
	}{
		// case 1
		{
			// in
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3, 4}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2, 5}, nil},
				4: &Node{4, 0, 0, []Id{1}, nil},
				5: &Node{5, 0, 0, []Id{3}, nil},
			},
			// want
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2}, nil},
			},
		},

		// case 2
		{
			// in
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3, 4}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2}, nil},
				4: &Node{4, 0, 0, []Id{1, 5}, nil},
				5: &Node{5, 0, 0, []Id{4}, nil},
			},
			// want
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2}, nil},
			},
		},

		// case 3
		{
			// in
			Graph{
				1: &Node{1, 0, 0, []Id{2}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{2, 4}, nil},
				4: &Node{4, 0, 0, []Id{3}, nil},
			},
			// want
			Graph{},
		},

		// case 4
		{
			// in
			Graph{
				1: &Node{1, 0, 0, []Id{}, nil},
			},
			// want
			Graph{},
		},

		// case 5
		{
			// in
			Graph{},
			// want
			Graph{},
		},

		// case 6
		{
			// in
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3, 4}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2}, nil},
				4: &Node{4, 0, 0, []Id{1, 5, 6}, nil},
				5: &Node{5, 0, 0, []Id{4}, nil},
				6: &Node{6, 0, 0, []Id{4}, nil},
			},
			// want
			Graph{
				1: &Node{1, 0, 0, []Id{2, 3}, nil},
				2: &Node{2, 0, 0, []Id{1, 3}, nil},
				3: &Node{3, 0, 0, []Id{1, 2}, nil},
			},
		},
	}

	for _, c := range cases {

		original := c.in

		c.in.removeDeadEnds()

		if !reflect.DeepEqual(c.in, c.want) {
			t.Errorf("%v.removeDeadEnds() == %v, want %v", original, c.in, c.want)
		}
	}
}
