package dijkstra

type Vertex struct {
	ID   int
	Arcs map[int]int // Arcs[vertex ID] = weight
}

/*
Graph graph's vertex should be from 0 to n-1 when there are n Vertices
*/
type Graph struct {
	Vertices []*Vertex
}

func NewEmptyGraph() *Graph {
	return &Graph{}
}
func NewGraph(vs []*Vertex) *Graph {
	g := new(Graph)
	g.Vertices = make([]*Vertex, len(vs))
	copy(g.Vertices, vs)
	return g
}

func (g *Graph) GetAllVertices() []*Vertex {
	return g.Vertices
}
func (g *Graph) Len() int { return len(g.Vertices) }

//source target conntect directly
func (g *Graph) HasEdge(source, target int) bool {
	if len(g.Vertices) <= source {
		return false
	}
	_, ok := g.Vertices[source].Arcs[target]
	return ok
}
func (g *Graph) AddVertex() int {
	id := len(g.Vertices)
	g.Vertices = append(g.Vertices, &Vertex{
		ID:   id,
		Arcs: make(map[int]int),
	})
	return id
}
func (g *Graph) AddEdge(src, dst, w int) bool {
	if src >= len(g.Vertices) || dst >= len(g.Vertices) {
		return false
	}
	g.Vertices[src].Arcs[dst] = w
	return true
}
func (g *Graph) RemoveEdge(src, dst int) bool {
	if src >= len(g.Vertices) || dst >= len(g.Vertices) {
		return false
	}
	_, ok := g.Vertices[src].Arcs[dst]
	if ok {
		delete(g.Vertices[src].Arcs, dst)
	}
	return ok
}
func (g *Graph) GetAllNeighbours(source int) []int {
	var t []int
	if len(g.Vertices) <= source {
		return nil
	}
	v := g.Vertices[source]

	for target := range v.Arcs {
		t = append(t, target)
	}
	return t
}
