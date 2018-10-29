package model

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
)

var dataSource = "postgres://pfs:123456@192.168.124.15/pfs_xxx?sslmode=disable"

func TestTokenNetwork_GetPaths(t *testing.T) {
	g1 := *dijkstra.NewEmptyGraph()
	g1.AddEdge(1, 2, 100)
	fmt.Println(fmt.Sprintf("****1>%s", utils.StringInterface(g1, 5)))
	g1.RemoveEdge(2, 1)
	t.Log("=======================")
	fmt.Println(fmt.Sprintf("****2>%s", utils.StringInterface(g1, 5)))
	g1.AddEdge(2, 1, 100)
	fmt.Println(fmt.Sprintf("****2>%s", utils.StringInterface(g1, 5)))
	g1.AddEdge(2, 3, 100)
	g1.AddEdge(3, 2, 100)
	g1.AddEdge(3, 4, 100)
	g1.AddEdge(4, 3, 100)

	/*g1.AddEdge(1, 0, 100)
	g1.AddEdge(0, 1, 100)*/
	fmt.Println(fmt.Sprintf("****2%s", utils.StringInterface(g1, 5)))

	paths01 := g1.AllShortestPath(1, 3)
	t.Log("test 01:", paths01)

	g1.AddEdge(0, 1, 55)
	g1.AddEdge(1, 2, 55)
	g1.AddEdge(2, 3, 55)
	g1.AddEdge(3, 4, 55)
	fmt.Println(fmt.Sprintf("****4%s", utils.StringInterface(g1, 5)))
	paths02 := g1.AllShortestPath(0, 3)
	t.Log("test 02:", paths02)

	g1.AddEdge(1, 2, 700)
	g1.AddEdge(2, 3, 100)
	g1.AddEdge(3, 4, 100)
	paths03 := g1.AllShortestPath(1, 3)
	t.Log("test 03:", paths03)

	g1.AddEdge(0, 1, 600)
	g1.AddEdge(1, 0, 200)
	g1.AddEdge(1, 2, 700)
	g1.AddEdge(2, 3, 100)
	g1.AddEdge(3, 4, 100)
	g1.RemoveEdge(2, 1)
	paths04 := g1.AllShortestPath(0, 3)
	t.Log("test 04:", paths04)

	db, err := storage.NewDatabase(dataSource, "0.001")
	if err != nil {
		t.Error(err)
	}
	view := &TokenNetwork{}
	view.db = db
	source := "0x1111111111111111111111111111111111111111"
	target := "0x6666666666666666666666666666666666666666"
	paths1, err := view.GetPaths(common.HexToAddress(source), common.HexToAddress(target), utils.EmptyAddress, big.NewInt(1), 1, "")
	if err != nil {
		t.Error(err)
		fmt.Println(err)
	}
	t.Log("test 1:", paths1)

	paths2, err := view.GetPaths(common.HexToAddress(target), common.HexToAddress(source), utils.EmptyAddress, big.NewInt(1), 1, "")
	if err != nil {
		t.Error(err)
		fmt.Println(err)
	}
	t.Log("test 2:", paths2)
}

func TestTokenNetwork_UpdateBalance(t *testing.T) {
	dbx, err := storage.NewDatabase(dataSource, "0.001")
	if err != nil {
		t.Error(err)
	}
	channelID := utils.NewRandomHash()
	channelID = common.HexToHash("0x1212121212121212121212121212121212121212121212121212121212121212")
	signer := utils.NewRandomAddress()
	nonce := uint64(6)
	transferAmount := big.NewInt(22)
	lockAmount := big.NewInt(0)
	xtwork := &TokenNetwork{
		db: dbx,
	}
	err = xtwork.UpdateBalance(channelID, signer, nonce, transferAmount, lockAmount)
	if err != nil {
		t.Error(err)
	}
}

func TestTokenNetwork_HandleChannelWithdrawEvent(t *testing.T) {
	tokenNetwork := utils.NewRandomAddress()
	channelID := utils.NewRandomHash()
	channelID = common.HexToHash("0x1212121212121212121212121212121212121212121212121212121212121212")
	participant1 := utils.NewRandomAddress()
	participant2 := utils.NewRandomAddress()
	participant1Balance := big.NewInt(10)
	participant2Balance := big.NewInt(5)
	dbx, err := storage.NewDatabase(dataSource, "0.001")
	if err != nil {
		t.Error(err)
	}

	xtwork := &TokenNetwork{
		db: dbx,
	}
	err = xtwork.HandleChannelWithdrawEvent(tokenNetwork, channelID, participant1, participant2, participant1Balance, participant2Balance)
	if err != nil {
		t.Error(err)
	}
}

//func (twork *TokenNetwork) HandleChannelClosedEvent(tokenNetwork common.Address, channelID common.Hash) (err error) {
func TestTokenNetwork_HandleChannelClosedEvent(t *testing.T) {
	tokenNetwork := utils.NewRandomAddress()
	channelID := utils.NewRandomHash()
	channelID = common.HexToHash("0x1212121212121212121212121212121212121212121212121212121212121212")
	dbx, err := storage.NewDatabase(dataSource, "0.001")
	if err != nil {
		t.Error(err)
	}

	xtwork := &TokenNetwork{
		db: dbx,
	}
	err = xtwork.HandleChannelClosedEvent(tokenNetwork, channelID)
	if err != nil {
		t.Error(err)
	}
}
