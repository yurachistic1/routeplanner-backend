package walkfilter

// MaxCrossing is the longest stretch of CrossingOnly road in metres that is
// kept, enough to get from a path on one side of a road to one further along.
const MaxCrossing = 200

// Edge is a segment between two OSM nodes.
type Edge struct {
	A, B   int64
	Class  Class
	Length float64 // metres
}

// TrimCrossings returns the edges with CrossingOnly stretches longer than
// maxLength removed. A stretch is a run of CrossingOnly edges between nodes
// where Preferred or OK edges join, so a road that paths meet every so often
// is kept in short pieces while long runs along it are dropped.
func TrimCrossings(edges []Edge, maxLength float64) []Edge {

	// nodes where a stretch ends because a better way joins
	anchors := make(map[int64]bool)
	for _, e := range edges {
		if e.Class != CrossingOnly {
			anchors[e.A] = true
			anchors[e.B] = true
		}
	}

	// group CrossingOnly edges into stretches: edges sharing a node that is
	// not an anchor belong to the same stretch
	stretches := newUnionFind()
	firstAt := make(map[int64]int)

	for i, e := range edges {
		if e.Class != CrossingOnly {
			continue
		}
		stretches.add(i)
		for _, n := range [2]int64{e.A, e.B} {
			if anchors[n] {
				continue
			}
			if j, ok := firstAt[n]; ok {
				stretches.union(i, j)
			} else {
				firstAt[n] = i
			}
		}
	}

	length := make(map[int]float64)
	for i, e := range edges {
		if e.Class == CrossingOnly {
			length[stretches.find(i)] += e.Length
		}
	}

	kept := make([]Edge, 0, len(edges))
	for i, e := range edges {
		if e.Class == CrossingOnly && length[stretches.find(i)] > maxLength {
			continue
		}
		kept = append(kept, e)
	}

	return kept
}

// unionFind groups edge indexes into disjoint sets.
type unionFind map[int]int

func newUnionFind() unionFind { return make(unionFind) }

func (u unionFind) add(i int) {
	if _, ok := u[i]; !ok {
		u[i] = i
	}
}

func (u unionFind) find(i int) int {
	for u[i] != i {
		u[i] = u[u[i]]
		i = u[i]
	}
	return i
}

func (u unionFind) union(i, j int) {
	u[u.find(i)] = u.find(j)
}
