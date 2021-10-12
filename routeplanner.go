package routeplanner

import (
	"encoding/json"
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
				way["highway"~"^(primary| secondary|tertiary|unclassified|residential|living_street|pedestrian|track|footway|steps|path|crossing|trailhead|cycleway)$"];
				way["footway"~"^(sidewalk|crossing)$"];
				way
				  ["highway"]
				  ["sidewalk"~"^(both|right|yes|left)$"];
				way["townpath"="yes"];
				way["foot"~"^(yes|designated|permissive)$"];
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

type CoordPair struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Route struct {
	Path     []CoordPair
	Distance float64
}

type Responce []Route

type Request struct {
	Lat      float64 `schema:"lat,required"`
	Lon      float64 `schema:"lon,required"`
	Distance float64 `schema:"distance,required"`
}

func RoutePlannerAPI(w http.ResponseWriter, r *http.Request) {

	var decoder = schema.NewDecoder()

	// Parse the request from query string
	var req Request
	if err := decoder.Decode(&req, r.URL.Query()); err != nil {
		// Report any parsing errors
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	bbox := overpass.BBox(req.Lat, req.Lon, req.Distance*0.8)

	res, err := overpass.Query(overpass.KumiSys, bbox+query)

	if err != nil || res.Elements == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	graph := OSMToGraph(res)

	routes := routing.TopRoutes(req.Lat, req.Lon, req.Distance*1000, graph)

	// Send response back to client as JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	response := routesToResponce(routes)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		return
	}
}

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
