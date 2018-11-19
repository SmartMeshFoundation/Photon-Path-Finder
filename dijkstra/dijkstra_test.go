package dijkstra

import (
	"testing"
)

//func TestGraph_AllShortestPath(t *testing.T) {
//	v := map[int]Vertex{
//		0: {
//			ID: 0,
//			Arcs: map[int]int{
//				1: 1,
//				2: 1,
//			},
//		},
//		1: {
//			ID: 1,
//			Arcs: map[int]int{
//				0: 1,
//				3: 1,
//			},
//		},
//		2: {
//			ID: 2,
//			Arcs: map[int]int{
//				0: 1,
//				3: 1,
//			},
//		},
//		3: {
//			ID: 3,
//			Arcs: map[int]int{
//				1: 1,
//				2: 1,
//			},
//		},
//	}
//	g := NewGraph(v)
//	result := g.AllShortestPath(0, 3, DefaultCostGetter)
//	/*
//		result:=[[0,1,3],[0,2,3]]
//	*/
//	if len(result) != 2 {
//		t.Error("shoude be two shortest path")
//	}
//}
//
//func TestGraph_AllShortestPath2(t *testing.T) {
//	v := map[int]Vertex{
//		0: {
//			ID: 0,
//			Arcs: map[int]int{
//				1: 1,
//				2: 1,
//			},
//		},
//		1: {
//			ID: 1,
//			Arcs: map[int]int{
//				0: 1,
//				3: 1,
//			},
//		},
//		2: {
//			ID: 2,
//			Arcs: map[int]int{
//				0: 1,
//				3: 1,
//				4: 1,
//			},
//		},
//		3: {
//			ID: 3,
//			Arcs: map[int]int{
//				1: 1,
//				2: 1,
//				4: 1,
//			},
//		},
//		4: {
//			ID: 4,
//			Arcs: map[int]int{
//				2: 1,
//				3: 1,
//			},
//		},
//	}
//	g := NewGraph(v)
//	result := g.AllShortestPath(0, 3, DefaultCostGetter)
//	/*
//		result:=[[0,1,3],[0,2,3]]
//	*/
//	if len(result) != 2 {
//		t.Error("shoude be two shortest path")
//	}
//}

func Benchmark_AllShortestPathMassVertices(b *testing.B) {
	numNodes := 10000
	v := make([]*Vertex, numNodes)
	for i := 0; i < numNodes-1; i++ {
		v[i] = &Vertex{
			ID: i,
			Arcs: map[int]int{
				i + 1: i + 1,
			},
		}
	}
	//最后一个节点,什么都不指向
	v[numNodes-1] = &Vertex{
		ID: numNodes - 1,
	}
	g := NewGraph(v)
	b.N = 20
	for i := 0; i < b.N; i++ {
		result := g.AllShortestPath(0, numNodes-1, DefaultCostGetter)
		if len(result) != 1 && len(result[0]) != numNodes-1 {
			b.Error("shoude be one shortest path")
			return
		}
	}

}

func TestGraph_AllShortestPath3(b *testing.T) {
	numNodes := 10000
	v := make([]*Vertex, numNodes)
	for i := 0; i < numNodes-1; i++ {
		v[i] = &Vertex{
			ID: i,
			Arcs: map[int]int{
				i + 1: 1,
			},
		}
	}
	//最后一个节点,什么都不指向
	v[numNodes-1] = &Vertex{
		ID: numNodes - 1,
	}
	g := NewGraph(v)

	//for i := 0; i < b.N; i++ {
	result := g.AllShortestPath(0, numNodes-1, DefaultCostGetter)
	if len(result) != 1 && len(result[0]) != numNodes-1 {
		b.Error("shoude be one shortest path")
		return
	}
}
