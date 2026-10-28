package tiles

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/yurachistic1/routeplanner-backend/walkfilter"
)

// testNetwork returns a ladder of nodes running east along two rows near
// Regent's Park, crossing the boundary between tiles 12/2045 and 12/2046 at
// lon -0.17578.
func testNetwork() (map[int64][2]float64, []walkfilter.Edge) {
	coords := make(map[int64][2]float64)
	var edges []walkfilter.Edge

	const n = 20
	for i := 0; i < n; i++ {
		lon := -0.1800 + float64(i)*0.0005
		coords[int64(1000+i)] = [2]float64{51.5300, lon}
		coords[int64(2000+i)] = [2]float64{51.5310, lon}

		// rungs between the rows
		edges = append(edges, walkfilter.Edge{A: int64(1000 + i), B: int64(2000 + i), Class: walkfilter.Preferred})
		if i > 0 {
			edges = append(edges,
				walkfilter.Edge{A: int64(1000 + i - 1), B: int64(1000 + i), Class: walkfilter.OK},
				walkfilter.Edge{A: int64(2000 + i - 1), B: int64(2000 + i), Class: walkfilter.CrossingOnly})
		}
	}

	return coords, edges
}

func edgeSet(edges []walkfilter.Edge) map[[2]int64]walkfilter.Class {
	set := make(map[[2]int64]walkfilter.Class)
	for _, e := range edges {
		set[[2]int64{min(e.A, e.B), max(e.A, e.B)}] = e.Class
	}
	return set
}

func TestSplitMerge(t *testing.T) {
	coords, edges := testNetwork()

	split := Split(edges, coords, Zoom)
	if len(split) != 2 {
		t.Fatalf("Split made %d tiles, want 2", len(split))
	}

	// edges crossing the boundary are in both tiles
	total := 0
	for _, tile := range split {
		total += len(tile.EdgeA)
	}
	if total <= len(edges) {
		t.Errorf("tiles hold %d edges in total, want more than %d because of boundary copies", total, len(edges))
	}

	// round trip every tile through the file format before merging
	var decoded []*Tile
	for _, tile := range split {
		var buf bytes.Buffer
		if err := tile.Encode(&buf); err != nil {
			t.Fatal(err)
		}
		d, err := Decode(&buf)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(d, tile) {
			t.Errorf("tile %v changed in encoding", tile.Key)
		}
		decoded = append(decoded, d)
	}

	gotCoords, gotEdges := Merge(decoded)

	if !reflect.DeepEqual(edgeSet(gotEdges), edgeSet(edges)) {
		t.Errorf("merged tiles have edges %v, want %v", edgeSet(gotEdges), edgeSet(edges))
	}
	if len(gotEdges) != len(edges) {
		t.Errorf("merged tiles have %d edges, want %d without duplicates", len(gotEdges), len(edges))
	}

	if len(gotCoords) != len(coords) {
		t.Errorf("merged tiles have %d nodes, want %d", len(gotCoords), len(coords))
	}
	for id, want := range coords {
		got := gotCoords[id]
		if math.Abs(got[0]-want[0]) > 1e-7 || math.Abs(got[1]-want[1]) > 1e-7 {
			t.Errorf("node %d at %v, want %v", id, got, want)
		}
	}
}

func TestSplitSingleTile(t *testing.T) {
	coords, edges := testNetwork()

	// keep only the part east of the boundary
	var east []walkfilter.Edge
	for _, e := range edges {
		if coords[e.A][1] > -0.175 && coords[e.B][1] > -0.175 {
			east = append(east, e)
		}
	}

	split := Split(east, coords, Zoom)
	if len(split) != 1 || split[0].Key != (Key{12, 2046, 1361}) {
		t.Fatalf("Split made tiles %v, want just 12/2046/1361", split)
	}
	if len(split[0].EdgeA) != len(east) {
		t.Errorf("tile has %d edges, want %d", len(split[0].EdgeA), len(east))
	}
}

func TestWriteRead(t *testing.T) {
	coords, edges := testNetwork()
	dir := t.TempDir()

	for _, tile := range Split(edges, coords, Zoom) {
		if err := Write(dir, tile); err != nil {
			t.Fatal(err)
		}
		got, err := Read(dir, tile.Key)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tile) {
			t.Errorf("tile %v changed when written and read back", tile.Key)
		}
	}

	if _, err := Read(dir, Key{12, 1, 1}); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Read of a missing tile returned %v, want fs.ErrNotExist", err)
	}
}

func TestLoad(t *testing.T) {
	coords, edges := testNetwork()
	dir := t.TempDir()

	written := Split(edges, coords, Zoom)
	for _, tile := range written {
		if err := Write(dir, tile); err != nil {
			t.Fatal(err)
		}
	}
	west, east := written[0].Key, written[1].Key

	got, err := Load(dir, []Key{west, {12, 1, 1}, east})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Key != west || got[1].Key != east {
		t.Errorf("Load returned %d tiles, want the 2 that exist, in key order", len(got))
	}

	// a corrupt tile is an error, not skipped
	os.WriteFile(filepath.Join(dir, east.Path()), []byte("not a tile"), 0o644)
	if _, err := Load(dir, []Key{east}); err == nil {
		t.Error("Load of a corrupt tile returned no error")
	}
}

func TestDecodeRejectsBadFiles(t *testing.T) {
	coords, edges := testNetwork()
	tile := Split(edges, coords, Zoom)[0]

	// encode writes gzip, so build broken files the same way
	encode := func(edit func(h *header, t *Tile)) *bytes.Buffer {
		h := header{Version: Version, Zoom: Zoom, X: uint32(tile.Key.X), Y: uint32(tile.Key.Y),
			Nodes: uint32(len(tile.IDs)), Edges: uint32(len(tile.EdgeA))}
		copy(h.Magic[:], magic)
		broken := *tile
		broken.EdgeA = append([]uint32{}, tile.EdgeA...)
		edit(&h, &broken)

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		for _, v := range []any{h, broken.IDs, broken.Lats, broken.Lons, broken.EdgeA, broken.EdgeB, broken.Class} {
			binary.Write(gz, binary.LittleEndian, v)
		}
		gz.Close()
		return &buf
	}

	cases := []struct {
		name string
		in   *bytes.Buffer
	}{
		{"not gzip", bytes.NewBufferString("hello")},
		{"wrong magic", encode(func(h *header, _ *Tile) { copy(h.Magic[:], "XXXX") })},
		{"old version", encode(func(h *header, _ *Tile) { h.Version = Version + 1 })},
		{"huge count", encode(func(h *header, _ *Tile) { h.Nodes = maxCount + 1 })},
		{"edge to missing node", encode(func(_ *header, t *Tile) { t.EdgeA[0] = uint32(len(t.IDs)) })},
		{"truncated", encode(func(h *header, _ *Tile) { h.Edges += 10 })},
	}

	for _, c := range cases {
		if _, err := Decode(c.in); err == nil {
			t.Errorf("%s: Decode returned no error", c.name)
		}
	}
}

func TestSplitSorted(t *testing.T) {
	coords := map[int64][2]float64{
		1: {51.53, -0.30}, 2: {51.53, -0.29},
		3: {51.53, -0.10}, 4: {51.53, -0.09},
		5: {51.44, -0.30}, 6: {51.44, -0.29},
	}
	edges := []walkfilter.Edge{{A: 3, B: 4}, {A: 5, B: 6}, {A: 1, B: 2}}

	var keys []Key
	for _, tile := range Split(edges, coords, Zoom) {
		keys = append(keys, tile.Key)
	}
	if !sort.SliceIsSorted(keys, func(i, j int) bool {
		return keys[i].X < keys[j].X || (keys[i].X == keys[j].X && keys[i].Y < keys[j].Y)
	}) {
		t.Errorf("Split returned tiles in order %v, want sorted by x then y", keys)
	}
}
