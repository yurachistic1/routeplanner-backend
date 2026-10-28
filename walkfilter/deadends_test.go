package walkfilter

import (
	"reflect"
	"sort"
	"testing"
)

func TestRemoveDeadEnds(t *testing.T) {
	cases := []struct {
		name  string
		edges []Edge
		want  [][2]int64
	}{
		{
			// triangle 1-2-3 with a spur 3-4-5
			"spur off a loop",
			[]Edge{{A: 1, B: 2}, {A: 2, B: 3}, {A: 3, B: 1}, {A: 3, B: 4}, {A: 4, B: 5}},
			[][2]int64{{1, 2}, {2, 3}, {3, 1}},
		},
		{
			// two spurs meeting at 4, which then becomes a dead end too
			"spurs leave a dead end junction",
			[]Edge{{A: 1, B: 2}, {A: 2, B: 3}, {A: 3, B: 1}, {A: 3, B: 4}, {A: 4, B: 5}, {A: 4, B: 6}},
			[][2]int64{{1, 2}, {2, 3}, {3, 1}},
		},
		{
			"a line has nothing left",
			[]Edge{{A: 1, B: 2}, {A: 2, B: 3}},
			[][2]int64{},
		},
		{
			// the duplicate 1-2 must not count as a second neighbour
			"duplicate segment and self loop",
			[]Edge{{A: 1, B: 2}, {A: 2, B: 1}, {A: 2, B: 2}},
			[][2]int64{},
		},
	}

	for _, c := range cases {
		got := [][2]int64{}
		for _, e := range RemoveDeadEnds(c.edges) {
			got = append(got, [2]int64{e.A, e.B})
		}
		sort.Slice(got, func(i, j int) bool { return got[i][0] < got[j][0] })
		sort.Slice(c.want, func(i, j int) bool { return c.want[i][0] < c.want[j][0] })

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: RemoveDeadEnds kept %v, want %v", c.name, got, c.want)
		}
	}
}
