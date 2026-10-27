package routing

import (
	"container/heap"
	"math"
)

// Reconstruct Path returns the optimal path from current to the start node.
func reconstructPath(cameFrom map[Id]Id, current Id) (totalPath []Id) {

	totalPath = []Id{current}

	next, ok := cameFrom[current]

	for ok {
		totalPath = append(totalPath, next)
		next, ok = cameFrom[next]
	}

	reverse(totalPath)

	return
}

// openItem is an entry in the A* open set.
type openItem struct {
	id     Id
	fScore float64
}

// openSet is a min-heap of openItems ordered by fScore.
type openSet []openItem

func (s openSet) Len() int           { return len(s) }
func (s openSet) Less(i, j int) bool { return s[i].fScore < s[j].fScore }
func (s openSet) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *openSet) Push(x any)        { *s = append(*s, x.(openItem)) }
func (s *openSet) Pop() any {
	old := *s
	item := old[len(old)-1]
	*s = old[:len(old)-1]
	return item
}

// A* search is used to complete the routes into a cycle. Nodes whose score
// improves are pushed again rather than updated in place, and stale entries
// are skipped when popped.
func aStar(path []*Node, graph Graph) []Id {

	goal := path[0]
	start := path[len(path)-1]

	open := &openSet{{start.Id, Haversine(start, goal)}}
	closed := make(map[Id]bool)

	cameFrom := make(map[Id]Id)

	gScore := map[Id]float64{start.Id: 0}

	for open.Len() != 0 {
		current := heap.Pop(open).(openItem).id

		if current == goal.Id {
			return reconstructPath(cameFrom, current)
		}

		if closed[current] {
			continue
		}
		closed[current] = true

		for _, id := range graph[current].Adjacent {
			if closed[id] {
				continue
			}

			d := graph[current].Edges[id].Distance

			tentative_gScore := gScore[current] + d
			if tentative_gScore < getWithDefault(gScore, id, math.Inf(1)) {
				cameFrom[id] = current
				gScore[id] = tentative_gScore

				h := Haversine(graph[id], goal)
				heap.Push(open, openItem{id, tentative_gScore + h})
			}
		}

	}

	return []Id{}
}

// CompleteRoute takes an incomplete cycle and completes it using A* as
// well as adjusting values such as length.
func completeRoute(route Route, graph Graph) Route {
	lastStretch := aStar(route.Path, graph)

	i := 0
	node := lastStretch[i]

	for node == route.Path[len(route.Path)-1].Id {
		route.Length -= graph[node].Edges[route.Path[len(route.Path)-2].Id].Distance
		route.Path = route.Path[:len(route.Path)-1]
		i++
		if i == len(lastStretch) {
			break
		}
		node = lastStretch[i]
	}

	for i = i - 1; i < len(lastStretch); i++ {
		route.Length += route.Path[len(route.Path)-1].Edges[lastStretch[i]].Distance
		route.Path = append(route.Path, graph[lastStretch[i]])
	}

	return route

}

// GetWithDefault returns a value from a map if a key is present or it
// returns the supplied default value if the key is not present.
func getWithDefault(m map[Id]float64, target Id, def float64) float64 {
	res, ok := m[target]

	if !ok {
		return def
	} else {
		return res
	}
}

// Reverse a list of Ids
func reverse(l []Id) {
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}
