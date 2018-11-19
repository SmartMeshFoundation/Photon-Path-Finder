package dijkstra

import (
	"math"
)

type CostGetter func(g *Graph, source, target int) (weight int)

//refer: http://www.linkedin.com/pulse/20140901041720-91330360-find-all-possible-shortest-paths-with-dijkstra-s-algorithm?trk=mp-reader-card
/**
* Computes all shortest paths between 2 Vertices using the
* Dijkstra's shortest path algorithm.
*
* @param source: starting vertex from which to find the shortest paths.
* @param target: end vertex
	[]int is one short path
*   return nil if there is no path
*/
func (g *Graph) AllShortestPath(source, target int, cg CostGetter) [][]int {
	//number of Vertices
	num := len(g.Vertices)
	//Distance to source vertex
	dist := make([]int, num)
	// Previous Vertices in shortest path from source to target.
	// Note: One vertex might have multiple previous Vertices
	prevs := make([][]int, num)
	// Initially all Vertices is unvisited
	// 1: Visited; 0: unvisited
	visited := make([]bool, num)
	for i := 0; i < num; i++ {
		dist[i] = math.MaxInt32
		visited[i] = false
	}

	// Distance from source to source
	dist[source] = 0
	//source is the current vertex
	var cur int = source
	//Mark source as Visited
	visited[cur] = true
	// main loop
	for !visited[target] {
		min := math.MaxInt32
		m := -1
		for i := 0; i < num; i++ {
			// tentative distance for the vertex i
			var d int
			if cg(g, cur, i) == math.MaxInt32 {
				d = math.MaxInt32
			} else {
				d = dist[cur] + cg(g, cur, i)
			}
			//Vertex i is not Visited yet
			if visited[i] == false {
				if d < dist[i] {
					//A shorter path to vertex i is found
					dist[i] = d
					//Clean up previous Vertices of i
					prevs[i] = nil
					//Add cur as a unique previous vertex of i
					prevs[i] = append(prevs[i], cur)
				} else if d == dist[i] {
					// An equivalent path to i is found
					// So add cur as a previous vertex of i
					prevs[i] = append(prevs[i], cur)
				}
				if min > dist[i] {
					// The vertex with min distance to source will be
					// the next current vertex
					min = dist[i]
					m = i
				}
			}
		}
		//All the unvisited Vertices are not reachable
		if min == math.MaxInt32 {
			break
		}
		cur = m
		visited[cur] = true
	}
	//Failed to find a path, the target might not be reachable
	if visited[target] == false {
		return nil
	}
	//fmt.Printf("prevs:%s\n,dist=%s", utils.StringInterface(prevs, 3), utils.StringInterface(dist, 3))
	_, paths := g.getAllPath(source, target, prevs, nil, num, nil)
	return paths
}

/**
* get all the paths by means of a backtracking algorithm
* @param source: starting vertex
* @param target: end vertex
* @param prevs: Previous Vertices in shortest path from
source to target, which is given by
allShortestPaths(...).
* @param path: current path
* @param num total number of vertex
* @param paths: all the path to return
*/
func (g *Graph) getAllPath(source, target int, prevs [][]int, path []int, num int, paths [][]int) ([]int, [][]int) {
	if len(path) > num {
		return path, paths
	}
	if source == target {
		path = append(path, source)
		// Print the path vector in the reverse order
		// in which Vertices push to the vector path

		//fmt.Println("Shortest Path xx:")
		//for i := len(path) - 1; i >= 0; i-- {
		//	fmt.Printf("%d  ", path[i])
		//}
		//fmt.Println("")
		newpath := make([]int, len(path))
		for i := 0; i < len(path); i++ {
			newpath[len(path)-i-1] = path[i]
		}
		paths = append(paths, newpath)
		return path, paths
	}
	for i := 0; i < len(prevs[target]); i++ {
		size := len(path)
		path = append(path, target)
		path, paths = g.getAllPath(source, prevs[target][i], prevs, path, num, paths)
		//rollback path
		path = path[:size]
	}
	return path, paths
}
func (g *Graph) buildCostMatrix() (cost [][]int) {
	cost = make([][]int, len(g.Vertices))
	for i := 0; i < len(cost); i++ {
		cost[i] = make([]int, len(cost))
	}
	for i := 0; i < len(cost); i++ {
		for j := 0; j < len(cost[i]); j++ {
			cost[i][j] = math.MaxInt32
		}
	}
	for k, v := range g.Vertices {
		for dst, weight := range v.Arcs {
			cost[k][dst] = weight
		}
	}
	return
}

func DefaultCostGetter(g *Graph, source, target int) int {
	w, ok := g.Vertices[source].Arcs[target]
	if !ok {
		return math.MaxInt32
	}
	return w
}
