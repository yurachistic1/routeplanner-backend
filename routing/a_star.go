package routing

import (
	"math"
)

func reconstruct_path(cameFrom map[Id]Id, current Id) (totalPath []Id) {

	totalPath = []Id{current}

	next, ok := cameFrom[current]

	for ok {
		totalPath = append(totalPath, next)
		next, ok = cameFrom[next]
	}

	reverse(totalPath)

	return
}

func a_star(path []*Node, graph Graph) []Id {

	goal := path[0]
	start := path[len(path)-1]
	openSet := map[Id]struct{}{start.Id: {}}

	cameFrom := make(map[Id]Id)

	gScore := map[Id]float64{start.Id: 0}

	fScore := map[Id]float64{start.Id: Haversine(start, goal)}

	for len(openSet) != 0 {
		min := math.Inf(1)

		var minNode Id

		for node, score := range fScore {
			_, ok := openSet[node]
			if score < min && ok {
				minNode = node
				min = score
			}
		}

		current := minNode
		if current == goal.Id {
			return reconstruct_path(cameFrom, current)
		}

		delete(openSet, current)

		neighbours := graph[current].Adjacent

		for _, id := range neighbours {
			node := graph[id]
			d := graph[current].Edges[id].Distance

			tentative_gScore := getWithDefault(gScore, current, math.Inf(1)) + d
			if tentative_gScore < getWithDefault(gScore, id, math.Inf(1)) {
				cameFrom[id] = current
				gScore[id] = tentative_gScore

				h := Haversine(node, goal)
				fScore[id] = getWithDefault(gScore, id, math.Inf(1)) + h

				openSet[id] = struct{}{}
			}
		}

	}

	return []Id{}
}

func completeRoute(route Route, graph Graph) Route {
	lastStretch := a_star(route.Path, graph)

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

func getWithDefault(m map[Id]float64, target Id, def float64) float64 {
	res, ok := m[target]

	if !ok {
		return def
	} else {
		return res
	}
}

func reverse(l []Id) {
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}
