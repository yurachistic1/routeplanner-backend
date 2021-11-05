package routeplanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"

	"github.com/yurachistic1/routeplanner-backend/overpass"
	"github.com/yurachistic1/routeplanner-backend/routing"
)

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
				way["highway"~"^(tertiary|unclassified|residential|living_street|pedestrian|track|footway|steps|path|crossing|trailhead|bridleway)$"];
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
	if req.Distance < 1 || req.Distance > 10 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", errors.New("invalid distance"))
	}

	// request data from overpass api
	bbox := overpass.BBox(req.Lat, req.Lon, req.Distance*0.8)

	res, err := overpass.Query(overpass.KumiSys, bbox+query)

	if err != nil || res.Elements == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	// process and calculate routes
	graph := OSMToGraph(res)

	routes := routing.TopRoutes(req.Lat, req.Lon, req.Distance*1000, graph)

	// Send response back to client as JSON
	w.WriteHeader(http.StatusOK)
	response := routesToResponce(routes)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		return
	}
}

// RoutesToResponse takes a routing.Routes object and condenses it to the most
// essential data needed in the server response.
func routesToResponce(routes routing.Routes) (res Responce) {

	for _, val := range routes {
		route := Route{Path: []CoordPair{}, Distance: val.Length}
		for _, node := range val.Path {
			route.Path = append(route.Path, CoordPair{node.Lat, node.Lon})
		}

		res = append(res, route)
	}

	return
}
