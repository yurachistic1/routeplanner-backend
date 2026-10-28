package routeplanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/gorilla/schema"

	"github.com/yurachistic1/routeplanner-backend/overpass"
	"github.com/yurachistic1/routeplanner-backend/routing"
	"github.com/yurachistic1/routeplanner-backend/tiles"
)

// Route distances in km. Routes can be up to MaxDistance where tiles have
// been generated, elsewhere map data comes from Overpass, which is too slow
// for more than OverpassMaxDistance.
const (
	MinDistance         = 1
	MaxDistance         = 35
	OverpassMaxDistance = 10
)

// tilesDir is the directory of tiles written by cmd/tilegen, set with the
// TILES_DIR environment variable.
var tilesDir = func() string {
	if dir := os.Getenv("TILES_DIR"); dir != "" {
		return dir
	}
	return "tiledata"
}()

var errNoTiles = fmt.Errorf("distances over %dkm are only available where tiles have been generated", OverpassMaxDistance)

const query = `
[out:json]
[timeout:25]
;
(
(
(
  (
	(
	  (
		(
		  (
			(
			  (
				way["highway"~"^(secondary|tertiary|unclassified|residential|living_street|pedestrian|track|footway|steps|path|crossing|trailhead|bridleway)$"];
				way["footway"~"^(sidewalk|crossing)$"];
				way
				  ["highway"]
				  ["sidewalk"~"^(both|right|yes|left)$"];
				way["townpath"="yes"];
				way["foot"~"^(yes|designated|permissive)$"];
				way["designation"="public_footpath"];
);
			  -
			  way["access"="customers"];
);
			-
			way["access"="private"];
);
		  -
		  way["area"="yes"];
);
		-
		way["foot"="no"];
);
	  -
	  way["indoor"="yes"];
);
	-
	way["tunnel"];
);
	- 
	way["route"="ferry"];
);
	- 
	way["sidewalk"~"^(no|none)$"];
);
  -
  (
	way
	  ["building"]
	  ["building"!="no"];
	node(w);
	way
	  ["highway"]
	  (bn);
);
);
out;
>;
out skel qt;
`

type CoordPair [2]float64

type Route struct {
	Path     []CoordPair `json:"path"`
	Distance float64     `json:"distance"`
}

type Responce []Route

type Request struct {
	Lat      float64 `schema:"lat,required"`
	Lon      float64 `schema:"lon,required"`
	Distance float64 `schema:"distance,required"`
}

// Handler function that is invoked by GCP.
func RoutePlannerAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "https://yurachistic1.github.io")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var decoder = schema.NewDecoder()

	// Parse the request from query string and report any parsing errors
	var req Request
	if err := decoder.Decode(&req, r.URL.Query()); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	// validate distance
	if req.Distance < MinDistance || req.Distance > MaxDistance {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", errors.New("invalid distance"))
		return
	}

	// load map data for the area the routes can reach
	graph, source, err := loadGraph(r, req.Lat, req.Lon, req.Distance)

	switch {
	case errors.Is(err, errNoTiles):
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", err)
		return
	case errors.Is(err, errOverpass):
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error: %s", err)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	w.Header().Set("X-Data-Source", source)

	// calculate routes
	routes := routing.TopRoutes(req.Lat, req.Lon, req.Distance*1000, graph)

	// Send response back to client as JSON
	w.WriteHeader(http.StatusOK)
	response := routesToResponce(routes)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		return
	}
}

var errOverpass = errors.New("overpass api")

// coverage is how close in metres to the start the tiles must have a node for
// them to be used. Starts just outside a generated region can overlap tiles
// at its edge that have nothing near them.
const coverage = 1000

// loadGraph returns the graph for routes of distance km from the
// coordinates, and where its data came from: "tiles" when tiles cover the
// start, otherwise "overpass".
func loadGraph(r *http.Request, lat, lon, distance float64) (routing.Graph, string, error) {
	side := distance * 0.8
	south, west, north, east := boundsAround(lat, lon, side)

	ts, err := tiles.Load(tilesDir, tiles.Covering(south, west, north, east, tiles.Zoom))
	if err != nil {
		return nil, "", err
	}
	if len(ts) > 0 {
		graph := TilesToGraph(ts, south, west, north, east)
		if hasNodeNear(graph, lat, lon, coverage) {
			return graph, "tiles", nil
		}
	}

	if distance > OverpassMaxDistance {
		return nil, "", errNoTiles
	}

	res, err := overpass.Query(r.Context(), overpass.Main, overpass.BBox(lat, lon, side)+query)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", errOverpass, err)
	}
	return OSMToGraph(res), "overpass", nil
}

// hasNodeNear reports whether the graph has a node within metres of the
// coordinates.
func hasNodeNear(graph routing.Graph, lat, lon, metres float64) bool {
	target := routing.Node{Lat: lat, Lon: lon}
	for _, n := range graph {
		if routing.Haversine(&target, n) <= metres {
			return true
		}
	}
	return false
}

// boundsAround returns a box with sides of side km centred on the
// coordinates, the same box overpass.BBox describes.
func boundsAround(lat, lon, side float64) (south, west, north, east float64) {
	dLat := (side / 2) / 111.32
	dLon := (side / 2) / (111.32 * math.Cos(lat*(math.Pi/180)))
	return lat - dLat, lon - dLon, lat + dLat, lon + dLon
}

// RoutesToResponse takes a routing.Routes object and condenses it to the most
// essential data needed in the server response.
func routesToResponce(routes routing.Routes) (res Responce) {

	// encode as [] rather than null when there are no routes
	res = Responce{}

	for _, val := range routes {
		route := Route{Path: []CoordPair{}, Distance: val.Length}
		for _, node := range val.Path {
			route.Path = append(route.Path, CoordPair{node.Lat, node.Lon})
		}

		res = append(res, route)
	}

	return
}
