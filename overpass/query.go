// Package overpass provides functions to query a specified
// overpass api instance and unmarshal the JSON response.
package overpass

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"time"
)

// Public Overpass API instance
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

// Meta contains meta properties and is a field of the Response struct.
type Meta struct {
	Timestamp time.Time `json:"timestamp_osm_base"`
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

// Query returns a Response value and an error.
// Api argument has to be of the form: address/api/interpreter.
// Query argument has to be an overpass QL statement with output format
//specified as JSON.
func Query(api string, query string) (response Response, err error) {

	overpassClient := http.Client{
		Timeout: time.Second * 60,
	}

	res, err := overpassClient.PostForm(
		api,
		url.Values{"data": []string{query}},
	)

	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return response, errors.New("overpass api error")
	}

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	json.Unmarshal(data, &response)

	return
}

// BBox returns a bbox setting as a string to be used in an overpass QL statement.
// It takes coordinates of the center of the bbox and desired side length in km.
//
// In a standard Overpass QL program, a bounding box is constructed with
//two decimal degree coordinate pairs in ISO 6709 standard order and format,
//and each value is separated with a comma. The values are, in order:
//southern-most latitude, western-most longitude, northern-most latitude,
//eastern-most longitude.
func BBox(lat float64, lon float64, side float64) (bbox string) {
	south := lat - ((side / 2) / 111.32)
	north := lat + ((side / 2) / 111.32)
	west := lon - ((side / 2) / (111.32 * math.Cos(lat*(math.Pi/180))))
	east := lon + ((side / 2) / (111.32 * math.Cos(lat*(math.Pi/180))))
	return fmt.Sprintf("[bbox:%f,%f,%f,%f]", south, west, north, east)
}
