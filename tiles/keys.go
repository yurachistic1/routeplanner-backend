// Package tiles stores the walkable graph of a region as small files, one per
// web map tile, so that a request only has to load the area it needs.
package tiles

import (
	"math"
	"path/filepath"
	"strconv"
)

// Zoom is the web map zoom level tiles are cut at. At zoom 12 a tile is about
// 6km across in London.
const Zoom = 12

// Key identifies a tile by zoom level and column and row, counted eastwards
// from 180°W and southwards from the top of the map.
type Key struct {
	Z int `json:"z"`
	X int `json:"x"`
	Y int `json:"y"`
}

// KeyFor returns the tile at zoom z that contains the coordinates.
func KeyFor(lat, lon float64, z int) Key {
	n := 1 << z
	r := lat * math.Pi / 180

	x := int(math.Floor((lon + 180) / 360 * float64(n)))
	y := int(math.Floor((1 - math.Log(math.Tan(r)+1/math.Cos(r))/math.Pi) / 2 * float64(n)))

	return Key{z, clamp(x, 0, n-1), clamp(y, 0, n-1)}
}

// Bounds returns the edges of the tile in degrees.
func (k Key) Bounds() (south, west, north, east float64) {
	n := float64(int(1) << k.Z)
	lon := func(x int) float64 { return float64(x)/n*360 - 180 }
	lat := func(y int) float64 { return math.Atan(math.Sinh(math.Pi*(1-2*float64(y)/n))) * 180 / math.Pi }

	return lat(k.Y + 1), lon(k.X), lat(k.Y), lon(k.X + 1)
}

// Covering returns the tiles at zoom z that overlap the bounding box.
func Covering(south, west, north, east float64, z int) []Key {
	sw, ne := KeyFor(south, west, z), KeyFor(north, east, z)

	keys := []Key{}
	for x := sw.X; x <= ne.X; x++ {
		for y := ne.Y; y <= sw.Y; y++ {
			keys = append(keys, Key{z, x, y})
		}
	}
	return keys
}

// Path returns the tile's file path relative to the tiles directory,
// z/x/y.bin.gz.
func (k Key) Path() string {
	return filepath.Join(strconv.Itoa(k.Z), strconv.Itoa(k.X), strconv.Itoa(k.Y)+".bin.gz")
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
