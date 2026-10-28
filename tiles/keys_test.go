package tiles

import (
	"math"
	"testing"
)

func TestKeyFor(t *testing.T) {
	cases := []struct {
		name     string
		lat, lon float64
		want     Key
	}{
		{"Regent's Park", 51.5313, -0.1570, Key{12, 2046, 1361}},
		{"Richmond Park", 51.4425, -0.2760, Key{12, 2044, 1363}},
		{"top left corner", 85.05, -180, Key{12, 0, 0}},
	}

	for _, c := range cases {
		if got := KeyFor(c.lat, c.lon, 12); got != c.want {
			t.Errorf("%s: KeyFor(%v, %v) == %v, want %v", c.name, c.lat, c.lon, got, c.want)
		}
	}
}

func TestBounds(t *testing.T) {
	s, w, n, e := Key{12, 2046, 1361}.Bounds()
	want := [4]float64{51.50874, -0.17578, 51.56341, -0.08789}

	for i, got := range [4]float64{s, w, n, e} {
		if math.Abs(got-want[i]) > 1e-5 {
			t.Errorf("Bounds of 12/2046/1361 == %v, want %v", [4]float64{s, w, n, e}, want)
			break
		}
	}

	// a point inside the tile maps back to it
	if k := KeyFor((s+n)/2, (w+e)/2, 12); k != (Key{12, 2046, 1361}) {
		t.Errorf("centre of 12/2046/1361 is in %v", k)
	}
}

func TestCovering(t *testing.T) {
	// the 28km box a 35km route from Regent's Park needs
	lat, lon, side := 51.5313, -0.1570, 28.0
	dLat := side / 2 / 111.32
	dLon := side / 2 / (111.32 * math.Cos(lat*math.Pi/180))

	keys := Covering(lat-dLat, lon-dLon, lat+dLat, lon+dLon, 12)

	if len(keys) != 30 {
		t.Fatalf("Covering returned %d tiles, want 30", len(keys))
	}
	first, last := keys[0], keys[len(keys)-1]
	if first != (Key{12, 2043, 1359}) || last != (Key{12, 2048, 1363}) {
		t.Errorf("Covering runs from %v to %v, want 12/2043/1359 to 12/2048/1363", first, last)
	}
}

func TestPath(t *testing.T) {
	if got := (Key{12, 2046, 1361}).Path(); got != "12/2046/1361.bin.gz" {
		t.Errorf("Path == %q, want %q", got, "12/2046/1361.bin.gz")
	}
}
