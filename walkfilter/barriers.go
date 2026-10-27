package walkfilter

// BlocksFoot reports whether a barrier node, such as a gate, stops people on
// foot, and a short reason. Nodes that are not barriers never block.
func BlocksFoot(tags Tags) (bool, string) {

	if tags["barrier"] == "" {
		return false, ""
	}

	// a locked vehicle gate often has a side gate for people
	if footAllowed[tags["foot"]] {
		return false, ""
	}

	switch {
	case oneOf(tags["foot"], "no", "private", "customers", "discouraged"):
		return true, "foot not allowed"
	case tags["locked"] == "yes" && tags["opening_hours"] == "":
		// gates with opening hours, like many park gates, are only locked
		// at times
		return true, "locked"
	case oneOf(tags["access"], "no", "private", "customers"):
		return true, "private access"
	}

	return false, ""
}

// RemoveBarriers returns the edges that do not touch a blocked node, cutting
// ways where they pass through barriers that stop people on foot.
func RemoveBarriers(edges []Edge, blocked map[int64]bool) []Edge {

	kept := make([]Edge, 0, len(edges))
	for _, e := range edges {
		if !blocked[e.A] && !blocked[e.B] {
			kept = append(kept, e)
		}
	}

	return kept
}
