package routing

import "math"

// Haversine calculates haversine distance between two nodes in meters.
func Haversine(n1, n2 *Node) float64 {
	const earthRadius = 6371000 // meters
	// lat lon in radians
	lat1Rad := n1.Lat * math.Pi / 180
	lat2Rad := n2.Lat * math.Pi / 180
	lon1Rad := n1.Lon * math.Pi / 180
	lon2Rad := n2.Lon * math.Pi / 180
	x := (lon2Rad - lon1Rad) * math.Cos((lat1Rad+lat2Rad)/2)
	y := (lat2Rad - lat1Rad)
	d := math.Sqrt(x*x+y*y) * earthRadius

	return d
}

// Bearing calculates bearing in degrees between two nodes. Range 0 ... 360.
func Bearing(n1, n2 *Node) float64 {
	// lat lon in radians
	lat1Rad := n1.Lat * math.Pi / 180
	lat2Rad := n2.Lat * math.Pi / 180
	lon1Rad := n1.Lon * math.Pi / 180
	lon2Rad := n2.Lon * math.Pi / 180

	y := math.Sin(lon2Rad-lon1Rad) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(lon2Rad-lon1Rad)

	bearingRad := math.Atan2(y, x)
	bearingDeg := math.Mod((bearingRad*180/math.Pi + 360), 360) // in degrees

	return bearingDeg
}

// BearingDifference calculates difference between two bearings in degrees.
// Return value is in the range between 0 and 180.
func bearingDifference(b1, b2 float64) float64 {
	d1 := b1 - b2
	d2 := b2 - b1

	if d1 < 0 {
		d1 += 360
	}

	if d2 < 0 {
		d2 += 360
	}

	if d1 < d2 {
		return d1
	} else {
		return d2
	}
}

// SectorAngle returns angle of a sector given length of the sector arc.
func sectorAngle(arcL, radius float64) float64 {
	return (arcL / (math.Pi * 2 * radius)) * 360
}

// RouteSimilarity estimates percent of nodes in the shorter route that are
// shared with the longer route.
func routeSimilarity(r1, r2 Route) (percentSimilar int) {

	var short, long map[Id]int
	var overlapCount int = 0

	if len(r1.Visited) < len(r2.Visited) {
		short, long = r1.Visited, r2.Visited
	} else {
		short, long = r2.Visited, r1.Visited
	}

	for key, _ := range short {
		_, ok := long[key]

		if ok {
			overlapCount++
		}
	}

	return (100 * overlapCount) / len(long)

}
