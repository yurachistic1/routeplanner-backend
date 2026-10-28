package routeplanner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yurachistic1/routeplanner-backend/tiles"
)

func TestRoutePlannerAPI(t *testing.T) {
	// a 20x20 grid of tiles near Regent's Park, about 2.6km by 1.7km
	dir := t.TempDir()
	ts, _ := gridTiles(20)
	for _, tile := range ts {
		if err := tiles.Write(dir, tile); err != nil {
			t.Fatal(err)
		}
	}

	old := tilesDir
	tilesDir = dir
	defer func() { tilesDir = old }()

	request := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		RoutePlannerAPI(w, httptest.NewRequest(http.MethodGet, "/?"+query, nil))
		return w
	}

	t.Run("routes from tiles", func(t *testing.T) {
		w := request("lat=51.5385&lon=-0.1667&distance=2")

		if w.Code != http.StatusOK {
			t.Fatalf("status %d, want 200: %s", w.Code, w.Body)
		}

		var routes Responce
		if err := json.NewDecoder(w.Body).Decode(&routes); err != nil {
			t.Fatal(err)
		}
		if len(routes) == 0 {
			t.Fatal("no routes returned")
		}
		for _, r := range routes {
			if r.Distance < 1500 || r.Distance > 3000 {
				t.Errorf("route of %.0fm returned for a 2km request", r.Distance)
			}
			if r.Path[0] != r.Path[len(r.Path)-1] {
				t.Errorf("route starts at %v but ends at %v", r.Path[0], r.Path[len(r.Path)-1])
			}
		}
	})

	t.Run("distance too long", func(t *testing.T) {
		if w := request("lat=51.5385&lon=-0.1667&distance=36"); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status %d, want 422", w.Code)
		}
	})

	t.Run("start just outside the tiles", func(t *testing.T) {
		// 3km south of the grid: its box overlaps the grid's tiles, but there
		// is nothing near the start, so tiles must not be used
		w := request("lat=51.5100&lon=-0.1667&distance=15")

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status %d, want 422 as if there were no tiles: %s", w.Code, w.Body)
		}
	})

	t.Run("no tiles", func(t *testing.T) {
		// Manchester has no tiles, even for a short route
		w := request("lat=53.4808&lon=-2.2426&distance=2")

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status %d, want 422", w.Code)
		}
		if !strings.Contains(w.Body.String(), "no map data") {
			t.Errorf("error %q does not explain that there is no map data", w.Body)
		}
	})
}
