package walkfilter

import (
	"reflect"
	"sort"
	"testing"
)

func TestTrimCrossings(t *testing.T) {
	const P, O, X = Preferred, OK, CrossingOnly

	cases := []struct {
		name  string
		edges []Edge
		want  []int64 // A ends of the kept edges
	}{
		{
			// path - 2 x 50m of road - path: kept
			"short crossing",
			[]Edge{{1, 2, P, 100}, {2, 3, X, 50}, {3, 4, X, 50}, {4, 5, P, 100}},
			[]int64{1, 2, 3, 4},
		},
		{
			// path - 3 x 100m of road - path: dropped
			"long run along road",
			[]Edge{{1, 2, P, 100}, {2, 3, X, 100}, {3, 4, X, 100}, {4, 5, X, 100}, {5, 6, P, 100}},
			[]int64{1, 5},
		},
		{
			// a 600m road that paths join every 150m is kept in pieces
			"paths joining along road",
			[]Edge{
				{1, 2, X, 150}, {2, 3, X, 150}, {3, 4, X, 150}, {4, 5, X, 150},
				{10, 2, P, 80}, {11, 3, O, 80}, {12, 4, P, 80},
				{13, 1, P, 80}, {14, 5, P, 80},
			},
			[]int64{1, 2, 3, 4, 10, 11, 12, 13, 14},
		},
		{
			// two roads meeting away from any path form one stretch
			"junction of two roads",
			[]Edge{{1, 2, X, 120}, {2, 3, X, 120}, {4, 2, X, 10}, {9, 1, P, 50}, {3, 8, P, 50}},
			[]int64{3, 9},
		},
		{
			"no crossing edges",
			[]Edge{{1, 2, P, 1000}, {2, 3, O, 1000}},
			[]int64{1, 2},
		},
	}

	for _, c := range cases {
		kept := TrimCrossings(c.edges, MaxCrossing)

		got := []int64{}
		for _, e := range kept {
			got = append(got, e.A)
		}
		sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: kept edges starting at %v, want %v", c.name, got, c.want)
		}
	}
}
