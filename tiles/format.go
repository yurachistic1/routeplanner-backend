package tiles

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yurachistic1/routeplanner-backend/walkfilter"
)

// Version of the file format. Tiles are always regenerated together, so
// files of another version are rejected rather than converted.
const Version = 1

const magic = "RPT1"

// maxCount guards against corrupt files claiming huge node or edge counts.
const maxCount = 50_000_000

// Tile holds the part of the graph in one tile as flat lists. Nodes are
// stored by position: IDs[i], Lats[i] and Lons[i] describe node i, and edge j
// connects nodes EdgeA[j] and EdgeB[j]. Edges crossing into a neighbouring
// tile are stored in both tiles, together with a copy of the node at the far
// end, so that merging tiles by node id joins them back up.
type Tile struct {
	Key   Key
	IDs   []int64 // OSM node ids
	Lats  []int32 // degrees × 1e7
	Lons  []int32 // degrees × 1e7
	EdgeA []uint32
	EdgeB []uint32
	Class []walkfilter.Class
}

type header struct {
	Magic   [4]byte
	Version uint16
	Zoom    uint8
	_       uint8
	X, Y    uint32
	Nodes   uint32
	Edges   uint32
}

// Encode writes the tile to w, gzip compressed.
func (t *Tile) Encode(w io.Writer) error {
	gz := gzip.NewWriter(w)

	h := header{
		Version: Version,
		Zoom:    uint8(t.Key.Z),
		X:       uint32(t.Key.X),
		Y:       uint32(t.Key.Y),
		Nodes:   uint32(len(t.IDs)),
		Edges:   uint32(len(t.EdgeA)),
	}
	copy(h.Magic[:], magic)

	for _, v := range []any{h, t.IDs, t.Lats, t.Lons, t.EdgeA, t.EdgeB, t.Class} {
		if err := binary.Write(gz, binary.LittleEndian, v); err != nil {
			return err
		}
	}

	return gz.Close()
}

// Decode reads a tile written by Encode.
func Decode(r io.Reader) (*Tile, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var h header
	if err := binary.Read(gz, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("reading tile header: %w", err)
	}
	if string(h.Magic[:]) != magic {
		return nil, errors.New("not a tile file")
	}
	if h.Version != Version {
		return nil, fmt.Errorf("tile format version %d, want %d: regenerate the tiles", h.Version, Version)
	}
	if h.Nodes > maxCount || h.Edges > maxCount {
		return nil, errors.New("tile header has invalid counts")
	}

	t := &Tile{
		Key:   Key{int(h.Zoom), int(h.X), int(h.Y)},
		IDs:   make([]int64, h.Nodes),
		Lats:  make([]int32, h.Nodes),
		Lons:  make([]int32, h.Nodes),
		EdgeA: make([]uint32, h.Edges),
		EdgeB: make([]uint32, h.Edges),
		Class: make([]walkfilter.Class, h.Edges),
	}

	for _, v := range []any{t.IDs, t.Lats, t.Lons, t.EdgeA, t.EdgeB, t.Class} {
		if err := binary.Read(gz, binary.LittleEndian, v); err != nil {
			return nil, fmt.Errorf("reading tile %v: %w", t.Key, err)
		}
	}

	for j := range t.EdgeA {
		if t.EdgeA[j] >= h.Nodes || t.EdgeB[j] >= h.Nodes {
			return nil, fmt.Errorf("tile %v has an edge to a missing node", t.Key)
		}
	}

	return t, nil
}

// Write saves the tile under dir at its Path.
func Write(dir string, t *Tile) (err error) {
	path := filepath.Join(dir, t.Key.Path())
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	w := bufio.NewWriter(f)
	if err := t.Encode(w); err != nil {
		return err
	}
	return w.Flush()
}

// Load reads the tiles for the keys from dir. Keys without a tile, such as
// outside the generated region, are skipped, so the result can be shorter
// than keys or empty.
func Load(dir string, keys []Key) ([]*Tile, error) {
	var tiles []*Tile

	for _, k := range keys {
		t, err := Read(dir, k)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		tiles = append(tiles, t)
	}

	return tiles, nil
}

// Read loads the tile k from dir. It returns an error satisfying
// errors.Is(err, fs.ErrNotExist) when there is no tile there, such as
// outside the generated region.
func Read(dir string, k Key) (*Tile, error) {
	f, err := os.Open(filepath.Join(dir, k.Path()))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t, err := Decode(bufio.NewReader(f))
	if err != nil {
		return nil, err
	}
	if t.Key != k {
		return nil, fmt.Errorf("file for tile %v contains tile %v", k, t.Key)
	}
	return t, nil
}
