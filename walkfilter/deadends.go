package walkfilter

// RemoveDeadEnds returns the edges left after repeatedly removing nodes with
// at most one neighbour, the same as routing.Graph.RemoveDeadEnds. Running it
// on the whole region makes tiles smaller; routing still runs it again on
// each request, since cutting out an area creates new dead ends at its edges.
func RemoveDeadEnds(edges []Edge) []Edge {
	neighbours := make(map[int64][]int64)
	seen := make(map[[2]int64]bool)

	for _, e := range edges {
		a, b := min(e.A, e.B), max(e.A, e.B)
		if a == b || seen[[2]int64{a, b}] {
			continue
		}
		seen[[2]int64{a, b}] = true
		neighbours[a] = append(neighbours[a], b)
		neighbours[b] = append(neighbours[b], a)
	}

	degree := make(map[int64]int, len(neighbours))
	var leaves []int64
	for n, nb := range neighbours {
		degree[n] = len(nb)
		if len(nb) <= 1 {
			leaves = append(leaves, n)
		}
	}

	removed := make(map[int64]bool)
	for len(leaves) > 0 {
		n := leaves[len(leaves)-1]
		leaves = leaves[:len(leaves)-1]
		if removed[n] {
			continue
		}
		removed[n] = true

		for _, m := range neighbours[n] {
			if removed[m] {
				continue
			}
			degree[m]--
			if degree[m] <= 1 {
				leaves = append(leaves, m)
			}
		}
	}

	kept := make([]Edge, 0, len(edges))
	for _, e := range edges {
		if e.A != e.B && !removed[e.A] && !removed[e.B] {
			kept = append(kept, e)
		}
	}

	return kept
}
