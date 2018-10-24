package model

import (
	"testing"
	"github.com/nkbai/dijkstra"
	"math/big"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
)

func TestTokenNetwork_GetPaths(t *testing.T) {

	db,err:=storage.NewDatabase("fps_xxx")
	//source := common.HexToAddress("0xc67f23CE04ca5E8DD9f2E1B5eD4FaD877f79267A")
	//target := common.HexToAddress("0x9Bae616edfF19A9C38F211d568baF24c35beECB3")
	g:=*dijkstra.NewEmptyGraph()//必须做双向的
	g.AddEdge(0, 1, 600)
	g.AddEdge(1, 0, 600)
	g.AddEdge(1, 2, 700)
	g.AddEdge(2, 1, 700)
	g.AddEdge(2, 3, 100)
	g.AddEdge(3, 2, 100)
	paths:=g.AllShortestPath(3,1)
	t.Log(paths)
	view := &TokenNetwork{}
	view.db=db
	pathss, err := view.GetPaths(utils.NewRandomAddress(), utils.NewRandomAddress(), big.NewInt(100), 1, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(pathss)
}
