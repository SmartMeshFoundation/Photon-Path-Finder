package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
	"fmt"
	"math/big"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"encoding/binary"
	"github.com/sirupsen/logrus"
)

// TokenNetwork token network view
type TokenNetwork struct {
	TokenNetworkAddress   common.Address
	ChannelID2Address     map[common.Hash][2]common.Address //cache key=channel_id value={0xparticipant1,0xparticipan2}
	PeerRelationshipGraph dijkstra.Graph
	MaxRelativeFee        int64
	db                    *storage.Database
	channelViews          map[common.Address]map[common.Address]*ChannelView
	GPeerAddress2Index    map[common.Address]int
	PeerNonceMap          map[common.Address][]*PeerNonce
}

// InitTokenNetwork token network initialization
func InitTokenNetwork(tokenNetworkAddress common.Address,db *storage.Database) (twork *TokenNetwork) {
	//read channel view from db
	channelinfos, err := db.GetAllChannelHistoryStorage(nil)
	if err != nil {
		return
	}
	channelID2Address := make(map[common.Hash][2]common.Address)
	gPeerAddress2Index := make(map[common.Address]int)
	peerNonceMap := make(map[common.Address][]*PeerNonce)
	var addrIndex = -1
	for _, channelinfo := range channelinfos {
		//if channelinfo.Status != StateChannelClose {
		peerAddr1 := common.HexToAddress(channelinfo.Partipant1)
		peerAddr2 := common.HexToAddress(channelinfo.Partipant2)
		var participant= [2]common.Address{peerAddr1, peerAddr2}
		var channelID= common.HexToHash(channelinfo.ChannelID)
		channelID2Address[channelID] = participant

		// Initialization dijkstra graph data
		addrIndex++
		gPeerAddress2Index[peerAddr1] = addrIndex
		addrIndex++
		gPeerAddress2Index[peerAddr1] = addrIndex

		peerNonceMap[peerAddr1] = append(peerNonceMap[peerAddr1], &PeerNonce{peerAddr1, channelID, channelinfo.P1Nonce})
		peerNonceMap[peerAddr1] = append(peerNonceMap[peerAddr2], &PeerNonce{peerAddr2, channelID, channelinfo.P2Nonce})
	}
	channelviews := make(map[common.Address]map[common.Address]*ChannelView)

	twork = &TokenNetwork{
		TokenNetworkAddress:   tokenNetworkAddress,
		ChannelID2Address:     channelID2Address,
		PeerRelationshipGraph: *dijkstra.NewEmptyGraph(),
		MaxRelativeFee:        0,
		db:                    db,
		channelViews:          channelviews,
		GPeerAddress2Index:    gPeerAddress2Index,
		PeerNonceMap:          peerNonceMap,
	}
	return
}

// HandleChannelOpenedEvent Handle ChannelOpened Event
func (twork *TokenNetwork)HandleChannelOpenedEvent(tokenNetwork common.Address, channelID common.Hash,participant1,participant2 common.Address) (err error) {

	var participant = [2]common.Address{participant1, participant2}
	twork.ChannelID2Address[channelID] = participant
	if _,exist:=storage.TokenNetwork2TokenMap[tokenNetwork];!exist{
		err=fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s",tokenNetwork)
		return
	}
	token:=storage.TokenNetwork2TokenMap[tokenNetwork]
	cview1:=InitChannelView(token,channelID, participant1, participant2, big.NewInt(0),StateChannelOpen,nil,twork.db)
	cview2:=InitChannelView(token,channelID, participant2, participant1, big.NewInt(0),StateChannelOpen,nil,twork.db)

	twork.PeerRelationshipGraph.AddEdge(twork.GPeerAddress2Index[participant1],twork.GPeerAddress2Index[participant2],100)
	twork.PeerRelationshipGraph.AddEdge(twork.GPeerAddress2Index[participant2],twork.GPeerAddress2Index[participant1],100)

	err=cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	err=cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return nil
}

// HandleChannelDeposit Handle Channel Deposit Event
func (twork *TokenNetwork)HandleChannelDepositEvent(tokenNetwork common.Address,channelID common.Hash,partner common.Address,totalDeposit *big.Int) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err = fmt.Errorf("Received ChannelDeposit event for unknown channel %s", channelID.String())
		return
	}

	if _,exist:=storage.TokenNetwork2TokenMap[tokenNetwork];!exist{
		err=fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s",tokenNetwork)
		return
	}
	token:=storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants= twork.ChannelID2Address[channelID]
	participant1 := participants[0]
	participant2 := participants[1]

	cview1 := InitChannelView(token,channelID, participant1, participant2, totalDeposit, StateChannelDeposit, nil, twork.db)
	cview2 := InitChannelView(token,channelID, participant2, participant1, totalDeposit, StateChannelDeposit, nil, twork.db)

	if partner == participant1 {
		err = cview1.UpdateCapacity(0, totalDeposit, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else if partner == participant2 {
		err = cview2.UpdateCapacity(0, totalDeposit, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else {
		err = fmt.Errorf("Partner in ChannelDeposit does not fit the internal channel", channelID.String())
	}
	return nil
}

// HandleChannelClosedEvent Handle Channel Closed Event
func (twork *TokenNetwork)HandleChannelClosedEvent(tokenNetwork common.Address,channelID common.Hash) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist{
		err=fmt.Errorf("Received ChannelClosed event for unknown channel %s",channelID.String())
		return
	}

	if _,exist:=storage.TokenNetwork2TokenMap[tokenNetwork];!exist{
		err=fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s",tokenNetwork)
		return
	}
	token:=storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants = twork.ChannelID2Address[channelID]
	participant1:=participants[0]
	participant2:=participants[1]

	twork.PeerRelationshipGraph.RemoveEdge(twork.GPeerAddress2Index[participant1],twork.GPeerAddress2Index[participant2])
	twork.PeerRelationshipGraph.RemoveEdge(twork.GPeerAddress2Index[participant2],twork.GPeerAddress2Index[participant1])
	//标记通道禁用
	cview1:=InitChannelView(token,channelID, participant1, participant2, big.NewInt(0),StateChannelClose,nil,twork.db)
	cview2:=InitChannelView(token,channelID, participant2, participant1, big.NewInt(0),StateChannelClose,nil,twork.db)

	err=cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	err=cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return
}

// HandleChannelWithdawEvent Handle Channel Withdaw Event
func (twork *TokenNetwork)HandleChannelWithdrawEvent(tokenNetwork common.Address,channelID common.Hash,
	participant1,participant2 common.Address,participant1Balance,participant2Balance *big.Int,
	) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err=fmt.Errorf("Received ChannelClosed event for unknown channel %s", channelID.String())
		return
	}

	if _,exist:=storage.TokenNetwork2TokenMap[tokenNetwork];!exist{
		err=fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s",tokenNetwork)
		return
	}
	token:=storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants = twork.ChannelID2Address[channelID]
	participant01 := participants[0]
	participant02 := participants[1]
	//初始不知道哪一方取钱
	cview1:=InitChannelView(token,channelID, participant1, participant2, big.NewInt(0),StateChannelWithdraw,nil,twork.db)
	cview2:=InitChannelView(token,channelID, participant2, participant1, big.NewInt(0),StateChannelWithdraw,nil,twork.db)

	if participant1 == participant01 {
		err=cview1.UpdateCapacity(0,participant1Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		err=cview2.UpdateCapacity(0,participant2Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else if participant1 == participant02 {
		cview1.UpdateCapacity(0,participant2Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		cview2.UpdateCapacity(0,participant1Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else {
		err=fmt.Errorf("Partner in ChannelDeposit event does not fit the internal channel %s", channelID.String())
	}
	return nil
}

//UpdateBalance Update Balance
func (twork *TokenNetwork)UpdateBalance(
	channelID common.Hash,
	signer common.Address,//signer=谁的balance proof
	nonce int,
	transferredAmount *big.Int,
	lockedAmount *big.Int,
	) (err error) {
	var partner common.Address
	participant1 := twork.ChannelID2Address[channelID][0]
	participant2 := twork.ChannelID2Address[channelID][1]

	token,err:=twork.db.GetTokenByChannelID(nil,channelID.String())
	if err!=nil{
		err=fmt.Errorf("An error occurred while obtaining token by channelID,err=%s",err)
		return
	}
	if token==""{
		err=fmt.Errorf("No token can be querying by channelID,channelID=%s",channelID)
		return
	}

	if signer == participant1 {
		partner = participant2
	} else if signer == participant2 {
		partner = participant1
	} else {
		logrus.Error("Balance proof signature does not match any of the participants")
		return fmt.Errorf("Balance proof signature error")
	}

	var oldNonce int
	for _,v:=range twork.PeerNonceMap[signer]{
		if v.ChannelID==channelID{
			oldNonce=v.Nonce
			break
		}
	}
	//token和通道是一一对应的
	cview := InitChannelView(common.HexToAddress(token),channelID, signer, partner, big.NewInt(0), StateUpdateBalance, nil, twork.db)
	//cview2 := InitChannelView(common.HexToAddress(token),channelID, partner, signer, big.NewInt(0), StateUpdateBalance, twork.PeerNonceMap[partner], twork.db)

	if nonce <= oldNonce {
		logrus.Error("Outdated balance proof.")
		return
	}
	//更新通道双方的Capacity
	err = cview.UpdateCapacity(
		nonce,
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
		lockedAmount,
	)

	if err != nil {
		logrus.Error("Update balance proof error,err=", err.Error())
	}
	return
}

type pathStru struct {

}

// pathResult is the json response for GetPaths
type pathResult struct {
	PathID  int      `json:"path_id"`
	PathHop int      `json:"path_hop"`
	fee     int64    `json:"fee"`
	Result  []int `json:"result"`
}

func (twork *TokenNetwork)GetPahts(
	source common.Address,
	target common.Address,
	value *big.Int,
	limitPaths int,
	sortDemand string,
	) (pathinfos []interface{}){
	xsource := twork.GPeerAddress2Index[source]
	xtarget := twork.GPeerAddress2Index[target]

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

	xsource = 0
	xtarget = 5
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
	//twork.PeerRelationshipGraph.RemoveEdge(1, 0) //删除本次计算余额不够的边（一次计算用）
	//twork.PeerRelationshipGraph.RemoveEdge(0, 1)
	*/

	//删除余额不够的边

	result := twork.PeerRelationshipGraph.AllShortestPath(xsource, xtarget)
	var pathInfos []interface{}
	for k,pathSlice:=range result {
		sinPathInfo := &pathResult{}
		sinPathInfo.PathID = k
		sinPathInfo.PathHop = len(pathSlice) - 2
		sinPathInfo.fee = 20
		sinPathInfo.Result = pathSlice
		pathInfos=append(pathInfos, sinPathInfo)
	}
	//fmt.Println(result)
	return pathInfos
}

func BytesToInt(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}

func IntToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
