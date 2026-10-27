package walkfilter

import (
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		tags Tags
		want Class
	}{
		// not for running
		{"no highway tag", Tags{"foot": "yes"}, Never},
		{"motorway", Tags{"highway": "motorway"}, Never},
		{"area", Tags{"highway": "pedestrian", "area": "yes"}, Never},
		{"indoor", Tags{"highway": "footway", "indoor": "yes"}, Never},
		{"indoor no", Tags{"highway": "footway", "indoor": "no"}, Preferred},
		{"ferry", Tags{"highway": "path", "route": "ferry"}, Never},
		{"foot no", Tags{"highway": "path", "foot": "no"}, Never},
		{"use sidepath", Tags{"highway": "secondary", "sidewalk": "both", "foot": "use_sidepath"}, Never},
		{"private", Tags{"highway": "track", "access": "private"}, Never},
		{"private but foot allowed", Tags{"highway": "track", "access": "private", "foot": "permissive"}, Preferred},
		{"customers", Tags{"highway": "footway", "access": "customers"}, Never},
		{"construction", Tags{"highway": "construction"}, Never},
		{"bus stop platform", Tags{"highway": "platform"}, Never},

		// paths
		{"footway", Tags{"highway": "footway"}, Preferred},
		{"sidewalk footway", Tags{"highway": "footway", "footway": "sidewalk"}, Preferred},
		{"steps", Tags{"highway": "steps"}, Preferred},
		{"track", Tags{"highway": "track"}, Preferred},
		{"underpass", Tags{"highway": "footway", "tunnel": "yes"}, Preferred},
		{"towpath", Tags{"highway": "service", "towpath": "yes"}, Preferred},
		{"public footpath", Tags{"highway": "unclassified", "designation": "public_footpath"}, Preferred},
		{"cycleway", Tags{"highway": "cycleway"}, Never},
		{"shared cycleway", Tags{"highway": "cycleway", "foot": "designated"}, Preferred},
		{"segregated cycleway", Tags{"highway": "cycleway", "segregated": "yes"}, Preferred},
		{"unsegregated cycleway", Tags{"highway": "cycleway", "segregated": "no"}, Preferred},
		{"cycleway with sidewalk", Tags{"highway": "cycleway", "sidewalk": "left"}, Preferred},
		{"cycleway on sidewalk", Tags{"highway": "cycleway", "footway": "sidewalk"}, Preferred},
		{"cycleway foot no", Tags{"highway": "cycleway", "segregated": "no", "foot": "no"}, Never},
		{"access no", Tags{"highway": "footway", "access": "no"}, Never},
		{"foot private", Tags{"highway": "footway", "foot": "private"}, Never},

		// streets and service roads
		{"residential", Tags{"highway": "residential"}, Preferred},
		{"residential without sidewalk", Tags{"highway": "residential", "sidewalk": "no"}, Preferred},
		{"living street", Tags{"highway": "living_street"}, Preferred},
		{"driveway", Tags{"highway": "service", "service": "driveway"}, Never},
		{"service with sidewalk", Tags{"highway": "service", "sidewalk": "left"}, OK},
		{"road tunnel", Tags{"highway": "residential", "tunnel": "yes"}, Never},
		{"tunnel no", Tags{"highway": "residential", "tunnel": "no"}, Preferred},

		// roads with sidewalks, in both tagging schemes
		{"primary sidewalk", Tags{"highway": "primary", "sidewalk": "both"}, OK},
		{"primary sidewalk:both", Tags{"highway": "primary", "sidewalk:both": "yes"}, OK},
		{"secondary sidewalk:left", Tags{"highway": "secondary", "sidewalk:left": "yes", "sidewalk:right": "no"}, OK},
		{"secondary separate sidewalk", Tags{"highway": "secondary", "sidewalk": "separate"}, OK},
		{"trunk sidewalk", Tags{"highway": "trunk", "sidewalk": "both"}, OK},

		// roads without sidewalks
		{"trunk", Tags{"highway": "trunk"}, Never},
		{"trunk foot yes", Tags{"highway": "trunk", "foot": "yes"}, Never},
		{"primary", Tags{"highway": "primary"}, CrossingOnly},
		{"primary unknown 30mph", Tags{"highway": "primary", "maxspeed": "30 mph"}, OK},
		{"primary no sidewalk 30mph", Tags{"highway": "primary", "sidewalk": "no", "maxspeed": "30 mph"}, CrossingOnly},
		{"b road no sidewalk 30mph", Tags{"highway": "secondary", "sidewalk": "no", "maxspeed": "30 mph"}, OK},
		{"b road no sidewalk", Tags{"highway": "secondary", "sidewalk": "no"}, CrossingOnly},
		{"b road sidewalk:both no", Tags{"highway": "tertiary", "sidewalk:both": "no"}, CrossingOnly},
		{"b road both sides no", Tags{"highway": "tertiary", "sidewalk:left": "no", "sidewalk:right": "no"}, CrossingOnly},
		{"b road one side unknown", Tags{"highway": "tertiary", "sidewalk:left": "no", "maxspeed": "60 mph"}, CrossingOnly},
		{"b road unknown", Tags{"highway": "secondary"}, CrossingOnly},
		{"b road unknown 30mph", Tags{"highway": "secondary", "maxspeed": "30 mph"}, OK},
		{"b road unknown 40mph", Tags{"highway": "tertiary", "maxspeed": "40 mph"}, CrossingOnly},
		{"b road unknown 50kmh", Tags{"highway": "tertiary", "maxspeed": "50"}, OK},
		{"b road national limit", Tags{"highway": "secondary", "maxspeed": "national"}, CrossingOnly},
		{"lane unknown", Tags{"highway": "unclassified"}, OK},
		{"lane no sidewalk", Tags{"highway": "unclassified", "sidewalk": "no"}, CrossingOnly},
		{"lane no sidewalk 20mph", Tags{"highway": "unclassified", "sidewalk": "none", "maxspeed": "20 mph"}, OK},
		{"slow trunk", Tags{"highway": "trunk", "maxspeed": "30 mph"}, Never},
	}

	for _, c := range cases {
		got, reason := Classify(c.tags)

		if got != c.want {
			t.Errorf("%s: Classify(%v) == %v (%s), want %v", c.name, c.tags, got, reason, c.want)
		}
	}
}

func TestParseSpeed(t *testing.T) {
	cases := []struct {
		in   string
		want float64
		ok   bool
	}{
		{"30 mph", 48.28032, true},
		{"20mph", 32.18688, true},
		{"48", 48, true},
		{" 50 ", 50, true},
		{"national", 0, false},
		{"GB:nsl_single", 0, false},
		{"", 0, false},
	}

	for _, c := range cases {
		got, ok := parseSpeed(c.in)

		if ok != c.ok || (ok && (got-c.want > 1e-9 || c.want-got > 1e-9)) {
			t.Errorf("parseSpeed(%q) == %v, %v, want %v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}
