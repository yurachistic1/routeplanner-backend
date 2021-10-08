package routing

import (
	"testing"
)

func TestRouteSimilarity(t *testing.T) {

	var (
		n1 = &Node{Id: 1}
		n2 = &Node{Id: 2}
		n3 = &Node{Id: 3}
		n4 = &Node{Id: 4}
		n5 = &Node{Id: 5}
		n6 = &Node{Id: 6}
	)

	var (
		route1 = Route{Path: []*Node{n1, n2, n3, n4}}
		route2 = Route{Path: []*Node{n1, n2}}
		route3 = Route{Path: []*Node{n5, n6}}
		route4 = Route{Path: []*Node{n1, n2, n5, n6}}
		route5 = Route{Path: []*Node{n3, n5, n6}}
	)

	cases := []struct {
		r1   Route
		r2   Route
		want int
	}{
		{route1, route2, 100},
		{route1, route4, 50},
		{route3, route5, 100},
		{route3, route2, 0},
		{route4, route5, 66},
	}

	for _, c := range cases {

		result := routeSimilarity(c.r1, c.r2)

		if result != c.want {
			t.Errorf("routeSimilarity(%v, %v) == %v, want %v",
				c.r1, c.r2, result, c.want)
		}
	}

}
