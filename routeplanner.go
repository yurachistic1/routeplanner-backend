package routeplanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/gorilla/schema"

	"github.com/yurachistic1/routeplanner-backend/routing"
	"github.com/yurachistic1/routeplanner-backend/tiles"
)

// Route distances in km.
const (
	MinDistance = 1
	MaxDistance = 35
)

// tilesDir is the directory of tiles written by cmd/tilegen, set with the
// TILES_DIR environment variable.
var tilesDir = func() string {
	if dir := os.Getenv("TILES_DIR"); dir != "" {
		return dir
	}
	return "tiledata"
}()

var errNoTiles = errors.New("there is no map data for this area yet")

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

// allowedOrigin returns the origin allowed to call the API from a browser.
// ALLOWED_ORIGIN overrides it for local development, e.g. http://localhost:5173.
func allowedOrigin() string {
	if o := os.Getenv("ALLOWED_ORIGIN"); o != "" {
		return o
	}
	return "https://yurachistic1.github.io"
}

// Handler function that is invoked by GCP.
func RoutePlannerAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin())
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
	graph, err := loadGraph(req.Lat, req.Lon, req.Distance)

	switch {
	case errors.Is(err, errNoTiles):
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", err)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	// calculate routes
	routes := routing.TopRoutes(req.Lat, req.Lon, req.Distance*1000, graph)

	// Send response back to client as JSON
	w.WriteHeader(http.StatusOK)
	response := routesToResponce(routes)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		return
	}
}

// coverage is how close in metres to the start the tiles must have a node for
// them to be used. Starts just outside a generated region can overlap tiles
// at its edge that have nothing near them.
const coverage = 1000

// loadGraph returns the graph from the tiles for routes of distance km from
// the coordinates, or errNoTiles when the tiles don't cover the start.
func loadGraph(lat, lon, distance float64) (routing.Graph, error) {
	south, west, north, east := boundsAround(lat, lon, distance*0.8)

	ts, err := tiles.Load(tilesDir, tiles.Covering(south, west, north, east, tiles.Zoom))
	if err != nil {
		return nil, err
	}
	if len(ts) == 0 {
		return nil, errNoTiles
	}

	graph := TilesToGraph(ts, south, west, north, east)
	if !hasNodeNear(graph, lat, lon, coverage) {
		return nil, errNoTiles
	}
	return graph, nil
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
// coordinates.
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
