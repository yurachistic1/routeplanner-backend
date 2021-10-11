package routeplanner

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func RoutePlannerAPI(w http.ResponseWriter, r *http.Request) {

	bbox := overpass.BBox(51.5211374, -0.1516939, 4)

	res, err := overpass.Query(overpass.KumiSys, bbox+query)

	if err != nil || res.Elements == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	graph := OSMToGraph(res)

	routes := routing.TopRoutes(51.5211374, -0.1516939, 5000, graph)

	// Send response back to client as JSON
	w.WriteHeader(http.StatusOK)
	response := routes
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		panic(err)
	}
}
