// Package overpass provides functions to query a specified
// overpass api instance and unmarshal the JSON response.
package overpass

import (
	_ "encoding/json"
	"time"
)

// Public Overpass API instances.
const (
	Main    = "https://lz4.overpass-api.de/api/interpreter"
	Russian = "https://overpass.openstreetmap.ru/api/interpreter"
	French  = "https://overpass.openstreetmap.fr/api/interpreter"
	Swiss   = "https://overpass.osm.ch/api/interpreter"
	KumiSys = "https://overpass.kumi.systems/api/interpreter"
	Taiwan  = "https://overpass.nchc.org.tw/api/interpreter"
)

// Response represents OSM data in JSON format as specified at
// http://overpass-api.de/output_formats.html
type Response struct {
	Version   float32
	Generator string
	Meta      Meta `json:"osm3s"`
	Elements  []Element
}

// Meta contains meta properties and is a field of the Response struct
type Meta struct {
	Timestamp time.Time
	Copyright string
}

// Element is a node, way or relation.
type Element struct {
	Type      string
	Id        int
	Timestamp time.Time
	Version   int
	Changeset int
	User      string
	Uid       int
	Tags      map[string]string

	// Node specific
	Lat float64
	Lon float64

	// Way specific
	Nodes []int

	// Relation specific
	Members []Member
}

// Member is a reference to component of a relation.
type Member struct {
	Type string
	Ref  int
	Role string
}
