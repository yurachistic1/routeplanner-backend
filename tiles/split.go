package tiles

import (
	"math"
	"sort"

	"github.com/yurachistic1/routeplanner-backend/walkfilter"
)

// Split divides edges between tiles at zoom z. Each node goes in the tile its
// coordinates fall in, and each edge in the tiles of both its nodes, with a
// copy of the node at the far end when that is in another tile. coords holds
// the lat and lon of every node the edges use. Tiles are returned sorted by
// key.
func Split(edges []walkfilter.Edge, coords map[int64][2]float64, z int) []*Tile {
	type builder struct {
		tile  *Tile
		index map[int64]uint32
	}
	builders := make(map[Key]*builder)

	get := func(k Key) *builder {
		b, ok := builders[k]
		if !ok {
			b = &builder{&Tile{Key: k}, make(map[int64]uint32)}
			builders[k] = b
		}
		return b
	}

	node := func(b *builder, id int64) uint32 {
		if i, ok := b.index[id]; ok {
			return i
		}
		c := coords[id]
		i := uint32(len(b.tile.IDs))
		b.index[id] = i
		b.tile.IDs = append(b.tile.IDs, id)
		b.tile.Lats = append(b.tile.Lats, int32(math.Round(c[0]*1e7)))
		b.tile.Lons = append(b.tile.Lons, int32(math.Round(c[1]*1e7)))
		return i
	}

	add := func(k Key, e walkfilter.Edge) {
		b := get(k)
		b.tile.EdgeA = append(b.tile.EdgeA, node(b, e.A))
		b.tile.EdgeB = append(b.tile.EdgeB, node(b, e.B))
		b.tile.Class = append(b.tile.Class, e.Class)
	}

	for _, e := range edges {
		a, b := coords[e.A], coords[e.B]
		ka, kb := KeyFor(a[0], a[1], z), KeyFor(b[0], b[1], z)

		add(ka, e)
		if kb != ka {
			add(kb, e)
		}
	}

	tiles := make([]*Tile, 0, len(builders))
	for _, b := range builders {
		tiles = append(tiles, b.tile)
	}
	sort.Slice(tiles, func(i, j int) bool {
		a, b := tiles[i].Key, tiles[j].Key
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Y < b.Y
	})

	return tiles
}

// Merge combines tiles into one set of nodes and edges, dropping the
// duplicates that edges crossing tile boundaries create. Coordinates are in
// degrees. Edge lengths are not stored in tiles, so Length is left at 0.
func Merge(tiles []*Tile) (coords map[int64][2]float64, edges []walkfilter.Edge) {
	coords = make(map[int64][2]float64)
	seen := make(map[[2]int64]bool)

	for _, t := range tiles {
		for i, id := range t.IDs {
			if _, ok := coords[id]; !ok {
				coords[id] = [2]float64{float64(t.Lats[i]) / 1e7, float64(t.Lons[i]) / 1e7}
			}
		}

		for j := range t.EdgeA {
			a, b := t.IDs[t.EdgeA[j]], t.IDs[t.EdgeB[j]]
			key := [2]int64{min(a, b), max(a, b)}
			if seen[key] {
				continue
			}
			seen[key] = true
			edges = append(edges, walkfilter.Edge{A: a, B: b, Class: t.Class[j]})
		}
	}

	return coords, edges
}
