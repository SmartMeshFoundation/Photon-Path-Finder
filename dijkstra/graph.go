package dijkstra

import (
	"fmt"
	"math"
)

type Vertex struct {
	ID   int
	Arcs map[int]int // Arcs[vertex ID] = weight
}

/*
Graph graph's vertex should be from 0 to n-1 when there are n vertices
*/
type Graph struct {
	vertices []*Vertex
}

func NewEmptyGraph() *Graph {
	return &Graph{}
}
func NewGraph(vs []*Vertex) *Graph {
	g := new(Graph)
	g.vertices = make([]*Vertex, len(vs))
	copy(g.vertices, vs)
	for _, v := range vs {
		for id, w := range v.Arcs {
			if w <= 0 {
				panic(fmt.Sprintf("%d-%d=%d weight must not be 0", v.ID, id, w))
			}
		}
	}
	return g
}
func (g *Graph) PrintGraph() {
	l := len(g.vertices)
	for i := 0; i < l; i++ {
		for j := 0; j < l; j++ {
			fmt.Printf("%20d", DefaultCostGetter(g, i, j))
		}
		fmt.Println("")
	}
}
func (g *Graph) GetAllVertices() []*Vertex {
	return g.vertices
}
func (g *Graph) Len() int { return len(g.vertices) }

//source target conntect directly
func (g *Graph) HasEdge(source, target int) bool {
	if len(g.vertices) <= source {
		return false
	}
	_, ok := g.vertices[source].Arcs[target]
	return ok
}
func (g *Graph) AddVertex() int {
	id := len(g.vertices)
	g.vertices = append(g.vertices, &Vertex{
		ID:   id,
		Arcs: make(map[int]int),
	})
	return id
}
func (g *Graph) AddEdge(src, dst, w int) bool {
	if w < 0 {
		panic(fmt.Sprintf("w must great or equal than zero"))
	}
	//todo 统一加1,避免w为0这种情况,这种情况会导致路径计算错误
	if w != math.MaxInt32 {
		w++
	}
	if src >= len(g.vertices) || dst >= len(g.vertices) {
		return false
	}
	g.vertices[src].Arcs[dst] = w
	return true
}
func (g *Graph) RemoveEdge(src, dst int) bool {
	if src >= len(g.vertices) || dst >= len(g.vertices) {
		return false
	}
	_, ok := g.vertices[src].Arcs[dst]
	if ok {
		delete(g.vertices[src].Arcs, dst)
	}
	return ok
}
func (g *Graph) GetAllNeighbours(source int) []int {
	var t []int
	if len(g.vertices) <= source {
		return nil
	}
	v := g.vertices[source]

	for target := range v.Arcs {
		t = append(t, target)
	}
	return t
}
