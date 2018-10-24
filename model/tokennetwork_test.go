package model

import (
	"testing"
	"github.com/nkbai/dijkstra"
	"math/big"
)

func TestTokenNetwork_GetPaths(t *testing.T) {

	db,err:=storage.NewDatabase("fps_xxx")

	g1:=*dijkstra.NewEmptyGraph()//必须做双向的
	g1.AddEdge(0, 1, 600)
	g1.AddEdge(1, 0, 600)
	g1.AddEdge(1, 2, 700)
	g1.AddEdge(2, 1, 700)
	g1.AddEdge(2, 3, 100)
	g1.AddEdge(3, 2, 100)
	paths1:=g1.AllShortestPath(0,3)
	t.Log("test 1:",paths1)

	g2:=*dijkstra.NewEmptyGraph()//必须做双向的
	g2.AddEdge(0, 1, 600)
	g2.AddEdge(1, 2, 700)
	g2.AddEdge(2, 3, 100)
	paths2:=g2.AllShortestPath(3,0)
	t.Log("test 1:",paths2)

	view := &TokenNetwork{}
	view.db=db
	paths3, err := view.GetPaths(utils.NewRandomAddress(), utils.NewRandomAddress(), big.NewInt(1), 1, "")
	if err != nil {
		t.Error(err)
	}
	t.Log("test 3:",paths3)
}
