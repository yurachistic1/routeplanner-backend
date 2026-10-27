package walkfilter

import (
	"reflect"
	"testing"
)

func TestBlocksFoot(t *testing.T) {
	cases := []struct {
		name string
		tags Tags
		want bool
	}{
		{"not a barrier", Tags{"access": "private"}, false},
		{"plain gate", Tags{"barrier": "gate"}, false},
		{"gate closed to cars", Tags{"barrier": "gate", "motor_vehicle": "no"}, false},
		{"private gate", Tags{"barrier": "gate", "access": "private"}, true},
		{"no access gate", Tags{"barrier": "gate", "access": "no"}, true},
		{"private gate open to foot", Tags{"barrier": "gate", "access": "private", "foot": "permissive"}, false},
		{"foot no", Tags{"barrier": "kissing_gate", "foot": "no"}, true},
		{"foot private", Tags{"barrier": "gate", "foot": "private"}, true},
		{"customers only", Tags{"barrier": "turnstile", "foot": "customers"}, true},
		{"locked", Tags{"barrier": "gate", "locked": "yes"}, true},
		{"locked at night", Tags{"barrier": "gate", "locked": "yes", "opening_hours": "07:00-dusk"}, false},
		{"locked with side gate", Tags{"barrier": "gate", "locked": "yes", "foot": "yes"}, false},
		{"unlocked", Tags{"barrier": "gate", "locked": "no"}, false},
		{"bollard", Tags{"barrier": "bollard"}, false},
	}

	for _, c := range cases {
		got, reason := BlocksFoot(c.tags)

		if got != c.want {
			t.Errorf("%s: BlocksFoot(%v) == %v (%s), want %v", c.name, c.tags, got, reason, c.want)
		}
	}
}

func TestRemoveBarriers(t *testing.T) {
	// path 1-2-3-4 with a locked gate at node 3, and a branch 2-5
	edges := []Edge{{1, 2, Preferred, 50}, {2, 3, Preferred, 50}, {3, 4, Preferred, 50}, {2, 5, OK, 50}}

	got := RemoveBarriers(edges, map[int64]bool{3: true})
	want := []Edge{{1, 2, Preferred, 50}, {2, 5, OK, 50}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RemoveBarriers == %v, want %v", got, want)
	}
}
