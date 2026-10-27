// Package walkfilter decides which OSM ways are suitable for running and how
// suitable they are, based on their tags.
package walkfilter

import (
	"strconv"
	"strings"
)

// Class describes how suitable a way is for running on.
type Class uint8

const (
	// Never ways are left out of the graph entirely.
	Never Class = iota
	// Preferred ways are paths, tracks and quiet streets.
	Preferred
	// OK ways are roads with a sidewalk, or quiet roads without one.
	OK
	// CrossingOnly ways are busier roads without a sidewalk. Only short
	// stretches of them are kept, see TrimCrossings.
	CrossingOnly
)

func (c Class) String() string {
	switch c {
	case Preferred:
		return "preferred"
	case OK:
		return "ok"
	case CrossingOnly:
		return "crossing only"
	default:
		return "never"
	}
}

// Tags are the tags of an OSM way.
type Tags map[string]string

// SlowSpeed is the speed limit in km/h at or under which a road without a
// sidewalk is still OK to run on. 50 km/h covers 30 mph.
const SlowSpeed = 50

// paths are highway types meant for people on foot.
var paths = set("footway", "path", "pedestrian", "track", "bridleway", "steps")

// streets are highway types that are quiet enough to run on regardless of
// sidewalks.
var streets = set("residential", "living_street")

// roadRule says how a road is treated when it has no sidewalk, or when
// sidewalks are not tagged, depending on whether its speed limit is at most
// SlowSpeed. Roads with a sidewalk are always OK.
type roadRule struct {
	noSidewalk, noSidewalkSlow Class
	unknown, unknownSlow       Class
}

var (
	trunk     = roadRule{Never, Never, Never, Never}
	primary   = roadRule{CrossingOnly, CrossingOnly, CrossingOnly, OK}
	secondary = roadRule{CrossingOnly, OK, CrossingOnly, OK}
	lane      = roadRule{CrossingOnly, OK, OK, OK}
)

// roads are highway types that are only OK with a sidewalk or, in some
// cases, a low speed limit.
var roads = map[string]roadRule{
	"trunk":          trunk,
	"trunk_link":     trunk,
	"primary":        primary,
	"primary_link":   primary,
	"secondary":      secondary,
	"secondary_link": secondary,
	"tertiary":       secondary,
	"tertiary_link":  secondary,
	"unclassified":   lane,
	"road":           lane,
}

// footAllowed are foot or access values that explicitly allow people on foot.
var footAllowed = set("yes", "designated", "permissive")

// Classify returns the class of a way and a short reason, used to report how
// many ways each rule affects.
func Classify(tags Tags) (Class, string) {

	highway, ok := tags["highway"]
	if !ok {
		return Never, "not a highway"
	}

	switch {
	case highway == "motorway" || highway == "motorway_link":
		return Never, "motorway"
	// squares and other pedestrian areas are kept so that paths leading into
	// them stay connected, routes then run around their edge
	case tags["area"] == "yes" && !oneOf(highway, "pedestrian", "footway"):
		return Never, "area"
	case isSet(tags["indoor"]):
		return Never, "indoor"
	case tags["route"] == "ferry":
		return Never, "ferry"
	case oneOf(tags["foot"], "no", "private", "use_sidepath", "discouraged"):
		return Never, "foot not allowed"
	case tags["foot"] == "customers":
		// paid entry or visitors only, such as cemeteries and gardens
		return Never, "customers only"
	case oneOf(tags["access"], "no", "private", "customers") && !footAllowed[tags["foot"]]:
		return Never, "private access"
	}

	// paths through tunnels are usually underpasses and are kept, roads
	// through tunnels are not pleasant to run in
	tunnel := isSet(tags["tunnel"])

	switch {
	case paths[highway], tags["towpath"] == "yes", tags["designation"] == "public_footpath":
		return Preferred, "path"
	case highway == "cycleway":
		// segregated=yes/no both mean the way is shared with people on foot,
		// it only says whether there is a line between them
		if footAllowed[tags["foot"]] || isSet(tags["segregated"]) || tags["segregated"] == "no" ||
			oneOf(tags["footway"], "sidewalk", "crossing") || sidewalkOf(tags) == yes {
			return Preferred, "cycleway open to foot"
		}
		return Never, "cycleway"
	case tunnel:
		return Never, "road tunnel"
	case streets[highway]:
		return Preferred, "quiet street"
	case highway == "service":
		if footAllowed[tags["foot"]] || sidewalkOf(tags) == yes {
			return OK, "service road open to foot"
		}
		return Never, "service road"
	}

	road, ok := roads[highway]
	if !ok {
		return Never, "other highway type"
	}

	switch sidewalkOf(tags) {
	case yes:
		return OK, "road with sidewalk"
	case no:
		if slow(tags) {
			return road.noSidewalkSlow, "slow road without sidewalk"
		}
		return road.noSidewalk, "road without sidewalk"
	default:
		if slow(tags) {
			return road.unknownSlow, "slow road, sidewalk unknown"
		}
		return road.unknown, "road, sidewalk unknown"
	}
}

type sidewalk int

const (
	unknown sidewalk = iota
	yes
	no
)

// sidewalkOf reads both the sidewalk=* tag and the newer per side
// sidewalk:both/left/right=* tags.
func sidewalkOf(tags Tags) sidewalk {
	present := func(v string) bool { return oneOf(v, "yes", "both", "left", "right", "separate") }
	absent := func(v string) bool { return oneOf(v, "no", "none") }

	switch {
	case present(tags["sidewalk"]), present(tags["sidewalk:both"]),
		present(tags["sidewalk:left"]), present(tags["sidewalk:right"]):
		return yes
	case absent(tags["sidewalk"]), absent(tags["sidewalk:both"]),
		absent(tags["sidewalk:left"]) && absent(tags["sidewalk:right"]):
		return no
	}
	return unknown
}

// slow reports whether the way has a speed limit of at most SlowSpeed.
// Missing or unparsable limits count as not slow.
func slow(tags Tags) bool {
	speed, ok := parseSpeed(tags["maxspeed"])
	return ok && speed <= SlowSpeed
}

// parseSpeed parses a maxspeed value such as "30 mph" or "48" into km/h.
func parseSpeed(v string) (float64, bool) {
	v = strings.TrimSpace(v)

	factor := 1.0
	if s, ok := strings.CutSuffix(v, "mph"); ok {
		v, factor = strings.TrimSpace(s), 1.609344
	}

	speed, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return speed * factor, true
}

func isSet(v string) bool { return v != "" && v != "no" }

func oneOf(v string, options ...string) bool {
	for _, o := range options {
		if v == o {
			return true
		}
	}
	return false
}

func set(values ...string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}
