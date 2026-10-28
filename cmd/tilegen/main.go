// Command tilegen builds the tiles the route planner loads, from an OSM
// extract such as one from download.geofabrik.de.
//
// It has its own go.mod so that the OSM reading library, which needs
// CGO_ENABLED=0 or pkg-config to build, is not a dependency of the API.
//
// Usage, from this directory:
//
//	CGO_ENABLED=0 go run . -in ../../devtools/data/greater-london-latest.osm.pbf -out ../../tiledata
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"

	"github.com/yurachistic1/routeplanner-backend/tiles"
	"github.com/yurachistic1/routeplanner-backend/walkfilter"
)

// manifest describes a generated set of tiles.
type manifest struct {
	Version   int        `json:"version"`
	Zoom      int        `json:"zoom"`
	Source    string     `json:"source"`
	SourceAt  time.Time  `json:"source_timestamp,omitzero"`
	Generated time.Time  `json:"generated"`
	Tiles     []tileInfo `json:"tiles"`
}

// step is the network left after one filtering step, for the report.
type step struct {
	name  string
	edges []walkfilter.Edge
}

type tileInfo struct {
	tiles.Key
	Nodes int   `json:"nodes"`
	Edges int   `json:"edges"`
	Bytes int64 `json:"bytes"`
}

func main() {
	in := flag.String("in", "", "OSM .pbf extract to read")
	out := flag.String("out", "", "directory to write tiles to")
	flag.Parse()

	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*in, *out); err != nil {
		log.Fatal(err)
	}
}

func run(in, out string) error {
	start := time.Now()

	if err := prepare(out); err != nil {
		return err
	}

	// pass 1: classify ways and note the nodes kept ways use
	type keptWay struct {
		nodes []int64
		class walkfilter.Class
	}
	var ways []keptWay
	reasons := make(map[string]int)
	coords := make(map[int64][2]float64)

	err := scan(in, true, false, func(o osm.Object) {
		w := o.(*osm.Way)
		if w.Tags.Find("highway") == "" {
			return
		}

		class, reason := walkfilter.Classify(walkfilter.Tags(w.Tags.Map()))
		reasons[fmt.Sprintf("%-13s %s", class, reason)]++
		if class == walkfilter.Never {
			return
		}

		ids := make([]int64, len(w.Nodes))
		for i, n := range w.Nodes {
			ids[i] = int64(n.ID)
			coords[ids[i]] = [2]float64{math.NaN(), math.NaN()}
		}
		ways = append(ways, keptWay{ids, class})
	})
	if err != nil {
		return err
	}

	// pass 2: coordinates of those nodes, and barriers among them
	blocked := make(map[int64]bool)

	header, err := scanNodes(in, func(n *osm.Node) {
		id := int64(n.ID)
		if _, ok := coords[id]; !ok {
			return
		}
		coords[id] = [2]float64{n.Lat, n.Lon}

		if blocks, _ := walkfilter.BlocksFoot(walkfilter.Tags(n.Tags.Map())); blocks {
			blocked[id] = true
		}
	})
	if err != nil {
		return err
	}

	// segments between consecutive nodes, once each, keeping the most
	// suitable class when ways overlap
	index := make(map[[2]int64]int)
	var edges []walkfilter.Edge

	for _, w := range ways {
		for i := 1; i < len(w.nodes); i++ {
			a, b := w.nodes[i-1], w.nodes[i]
			ca, cb := coords[a], coords[b]
			if a == b || math.IsNaN(ca[0]) || math.IsNaN(cb[0]) {
				continue
			}

			key := [2]int64{min(a, b), max(a, b)}
			if j, ok := index[key]; ok {
				edges[j].Class = min(edges[j].Class, w.class)
				continue
			}
			index[key] = len(edges)
			edges = append(edges, walkfilter.Edge{A: a, B: b, Class: w.class, Length: haversine(ca, cb)})
		}
	}

	steps := []step{{"kept ways", edges}}

	edges = walkfilter.RemoveBarriers(edges, blocked)
	steps = append(steps, step{fmt.Sprintf("cut at %d barriers", len(blocked)), edges})

	edges = walkfilter.TrimCrossings(edges, walkfilter.MaxCrossing)
	steps = append(steps, step{"crossings trimmed", edges})

	edges = walkfilter.RemoveDeadEnds(edges)
	steps = append(steps, step{"dead ends removed", edges})

	// write tiles and the manifest
	m := manifest{
		Version:   tiles.Version,
		Zoom:      tiles.Zoom,
		Source:    filepath.Base(in),
		Generated: time.Now().UTC().Truncate(time.Second),
	}
	if header != nil {
		m.SourceAt = header.ReplicationTimestamp
	}

	var totalBytes int64
	for _, t := range tiles.Split(edges, coords, tiles.Zoom) {
		if err := tiles.Write(out, t); err != nil {
			return err
		}
		info, err := os.Stat(filepath.Join(out, t.Key.Path()))
		if err != nil {
			return err
		}
		totalBytes += info.Size()
		m.Tiles = append(m.Tiles, tileInfo{t.Key, len(t.IDs), len(t.EdgeA), info.Size()})
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}

	report(reasons, steps, m, totalBytes, time.Since(start))
	return nil
}

// prepare makes sure out is either new, empty, or a previous tilegen output,
// which is cleared so that tiles outside the new region do not linger.
func prepare(out string) error {
	entries, err := os.ReadDir(out)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(out, 0o755)
	}
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	manifestPath := filepath.Join(out, "manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("%s is not empty and has no manifest.json, so it is not tilegen output: not overwriting it", out)
	}
	if err := os.RemoveAll(filepath.Join(out, strconv.Itoa(tiles.Zoom))); err != nil {
		return err
	}
	return os.Remove(manifestPath)
}

func scan(path string, skipNodes, skipWays bool, f func(osm.Object)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	sc := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	defer sc.Close()
	sc.SkipNodes, sc.SkipWays, sc.SkipRelations = skipNodes, skipWays, true

	for sc.Scan() {
		f(sc.Object())
	}
	return sc.Err()
}

// scanNodes reads the nodes of the extract and returns its header, which
// says how up to date the data is.
func scanNodes(path string, f func(*osm.Node)) (*osmpbf.Header, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sc := osmpbf.New(context.Background(), file, runtime.GOMAXPROCS(-1))
	defer sc.Close()
	sc.SkipWays, sc.SkipRelations = true, true

	header, err := sc.Header()
	if err != nil {
		return nil, err
	}

	for sc.Scan() {
		f(sc.Object().(*osm.Node))
	}
	return header, sc.Err()
}

func report(reasons map[string]int, steps []step, m manifest, totalBytes int64, took time.Duration) {

	fmt.Println("ways by class and rule:")
	keys := make([]string, 0, len(reasons))
	for k := range reasons {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %8d  %s\n", reasons[k], k)
	}

	fmt.Println("\nnetwork after each step:")
	for _, s := range steps {
		km := make(map[walkfilter.Class]float64)
		for _, e := range s.edges {
			km[e.Class] += e.Length / 1000
		}
		fmt.Printf("  %-22s %9d edges  preferred %7.0f km  ok %6.0f km  crossing %4.0f km\n",
			s.name, len(s.edges), km[walkfilter.Preferred], km[walkfilter.OK], km[walkfilter.CrossingOnly])
	}

	fmt.Printf("\nwrote %d tiles, %.1f MB, from %s", len(m.Tiles), float64(totalBytes)/1e6, m.Source)
	if !m.SourceAt.IsZero() {
		fmt.Printf(" (data as of %s)", m.SourceAt.Format("2006-01-02"))
	}
	fmt.Printf(" in %v\n", took.Round(100*time.Millisecond))
}

func haversine(a, b [2]float64) float64 {
	lat1, lat2 := a[0]*math.Pi/180, b[0]*math.Pi/180
	dLat, dLon := lat2-lat1, (b[1]-a[1])*math.Pi/180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * 6371000 * math.Asin(math.Sqrt(h))
}
