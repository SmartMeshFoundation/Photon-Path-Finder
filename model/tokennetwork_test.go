package model

import (
	"testing"
	"github.com/nkbai/dijkstra"
	"math/big"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
)

func TestTokenNetwork_GetPaths(t *testing.T) {
	g1:=*dijkstra.NewEmptyGraph()//必须做双向的
	g1.AddEdge(0, 1, 600)
	g1.AddEdge(1, 0, 200)
	g1.AddEdge(1, 2, 700)
	g1.AddEdge(2, 1, 700)
	g1.AddEdge(2, 3, 100)
	g1.AddEdge(3, 2, 100)
	g1.RemoveEdge(1,3)
	paths0:=g1.AllShortestPath(0,3)
	t.Log("test 0:",paths0)

	db,err:=storage.NewDatabase("postgres://pfs:123456@localhost/pfs_xxx?sslmode=disable")
	if err!=nil{
		t.Error(err)
	}
	view := &TokenNetwork{}
	view.db=db
	source:="0x1111111111111111111111111111111111111111"
	target:="0x6666666666666666666666666666666666666666"
	paths1, err := view.GetPaths(common.HexToAddress(source), common.HexToAddress(target), big.NewInt(1), 1, "")
	if err != nil {
		t.Error(err)
	}
	t.Log("test 1:",paths1)

	//反之
	paths2, err := view.GetPaths(common.HexToAddress(target), common.HexToAddress(source), big.NewInt(1), 1, "")
	if err != nil {
		t.Error(err)
	}
	t.Log("test 2:",paths2)
}

func TestTokenNetwork_UpdateBalance(t *testing.T) {
	dbx,err:=storage.NewDatabase("fps_xxx")
	if err!=nil{
		t.Error(err)
	}
	channelID:=utils.NewRandomHash()
	signer:=utils.NewRandomAddress()
	nonce:=uint64(6)
	transferAmount:=big.NewInt(22)
	lockAmount:=big.NewInt(0)
	xtwork:=&TokenNetwork{
		db:dbx,
	}
	err=xtwork.UpdateBalance(channelID,signer,nonce,transferAmount,lockAmount)
	if err!=nil{
		t.Error(err)
	}
}



func TestTokenNetwork_HandleChannelWithdrawEvent(t *testing.T) {
	tokenNetwork:=utils.NewRandomAddress()
	channelID:=utils.NewRandomHash()
	participant1:=utils.NewRandomAddress()
	participant2:=utils.NewRandomAddress()
	participant1Balance:=big.NewInt(10)
	participant2Balance:=big.NewInt(5)
	dbx,err:=storage.NewDatabase("fps_xxx")
	if err!=nil{
		t.Error(err)
	}

	xtwork:=&TokenNetwork{
		db:dbx,
	}
	err=xtwork.HandleChannelWithdrawEvent(tokenNetwork,channelID,participant1,participant2,participant1Balance,participant2Balance)
	if err!=nil{
		t.Error(err)
	}
}


