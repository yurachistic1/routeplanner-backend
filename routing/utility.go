package routing

import "math"

// Haversine calculates haversine distance between two nodes.
func haversine(n1, n2 *Node) float64 {
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

// Bearing calculates bearing in degrees between two nodes.
func bearing(n1, n2 *Node) float64 {
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
