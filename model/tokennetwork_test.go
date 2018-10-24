package model

import (
	"testing"
	"github.com/nkbai/dijkstra"
	"math/big"
	"github.com/ethersphere/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
)

func TestTokenNetwork_GetPaths(t *testing.T) {

	source := common.HexToAddress("0xc67f23CE04ca5E8DD9f2E1B5eD4FaD877f79267A")
	target := common.HexToAddress("0x9Bae616edfF19A9C38F211d568baF24c35beECB3")
	channelid := "0x0398beea63f098e2d3bb59884be79eda00cf042e39ad65e5c43a0a280f969f93"
	cid2addr := make(map[common.Hash][2]common.Address)
	var participant= [2]common.Address{source, target}
	cid2addr[common.HexToHash(channelid)] = participant
	view := &TokenNetwork{}
	dijstraGraph := *dijkstra.NewEmptyGraph()
	dijstraGraph.AddEdge(0, 1, 100)
	dijstraGraph.AddEdge(1, 0, 100)
	dijstraGraph.AddEdge(1, 2, 50)
	dijstraGraph.AddEdge(2, 1, 50)
	dijstraGraph.AddEdge(2, 3, 10)
	dijstraGraph.AddEdge(3, 2, 10)
	dijstraGraph.AddEdge(1, 3, 10)
	dijstraGraph.AddEdge(3, 1, 10)
	dijstraGraph.AddEdge(4, 5, 10)
	dijstraGraph.AddEdge(5, 4, 10)
	dijstraGraph.RemoveEdge(1, 0)
	dijstraGraph.RemoveEdge(0, 1)
	//last del channel
	view.PeerRelationshipGraph = dijstraGraph
	add2index := make(map[common.Address]int)
	add2index[source] = 1
	add2index[target] = 2
	//view.GPeerAddress2Index=
	paths, err := view.GetPaths(utils.NewRandomAddress(), utils.NewRandomAddress(), big.NewInt(100), 1, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(paths)
}

/*//test case:
	gMapToIndex := make(map[common.Address]int)
	index1 := len(gMapToIndex)
	fmt.Println(index1)
	gMapToIndex[common.HexToAddress("0xc67f23CE04ca5E8DD9f2E1B5eD4FaD877f79267A")] = index1
	index2 := len(gMapToIndex)
	fmt.Println(index2)
	gMapToIndex[common.HexToAddress("0xd4bd8fAcD16704C2B6Ed4B06775467d44f216174")] = index2
	index3 := len(gMapToIndex)
	fmt.Println(index3)
	gMapToIndex[common.HexToAddress("0xd4bd8fAcD16704C2B6Ed4B06775467d44f216188")] = index3

	xsource = 3
	xtarget = 2
	twork.PeerRelationshipGraph=*dijkstra.NewEmptyGraph()
	twork.PeerRelationshipGraph.AddEdge(0, 1, 100)
	twork.PeerRelationshipGraph.AddEdge(1, 0, 100)
	twork.PeerRelationshipGraph.AddEdge(1, 2, 50)
	twork.PeerRelationshipGraph.AddEdge(2, 1, 50)
	twork.PeerRelationshipGraph.AddEdge(2, 3, 10)
	twork.PeerRelationshipGraph.AddEdge(3, 2, 10)
	twork.PeerRelationshipGraph.AddEdge(1, 3, 10)
	twork.PeerRelationshipGraph.AddEdge(3, 1, 10)
	twork.PeerRelationshipGraph.AddEdge(4, 5, 10)
	twork.PeerRelationshipGraph.AddEdge(5, 4, 10)
	twork.PeerRelationshipGraph.RemoveEdge(1, 0) //删除本次计算余额不够的边（一次计算用）
	twork.PeerRelationshipGraph.RemoveEdge(0, 1)
	*/