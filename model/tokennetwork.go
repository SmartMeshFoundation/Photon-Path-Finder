package model

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/nkbai/dijkstra"
)

// TokenNetwork token network view
type TokenNetwork struct {
	TokenNetworkAddress   common.Address
	ChannelID2Address     map[common.Hash][2]common.Address //cache key=channel_id value={0xparticipant1,0xparticipan2}
	//PeerRelationshipGraph dijkstra.Graph
	MaxRelativeFee        int64
	db                    *storage.Database
	channelViews          map[common.Address]map[common.Address]*ChannelView
	//GPeerAddress2Index    map[common.Address]int
	//CacheRemoveEdge       map[int]int //Remember those who were removed for calc.
}

// InitTokenNetwork token network initialization
func InitTokenNetwork(tokenNetworkAddress common.Address, db *storage.Database) (twork *TokenNetwork) {
	//read channel view from db
	channelinfos, err := db.GetAllChannelHistoryStorage(nil)
	if err != nil {
		return
	}
	channelID2Address := make(map[common.Hash][2]common.Address)
	//gPeerAddress2Index := make(map[common.Address]int)
	//dijstraGraph := *dijkstra.NewEmptyGraph()
	//var addrIndex = -1
	for _, channelinfo := range channelinfos {
		//if channelinfo.Status != StateChannelClose {
		peerAddr1 := common.HexToAddress(channelinfo.Partipant1)
		peerAddr2 := common.HexToAddress(channelinfo.Partipant2)
		var participant = [2]common.Address{peerAddr1, peerAddr2}
		var channelID = common.HexToHash(channelinfo.ChannelID)
		channelID2Address[channelID] = participant

		// Initialization dijkstra graph data
		/*if _, exist := gPeerAddress2Index[peerAddr1]; !exist {
			addrIndex++
			gPeerAddress2Index[peerAddr1] = addrIndex
		}
		if _, exist := gPeerAddress2Index[peerAddr2]; !exist {
			addrIndex++
			gPeerAddress2Index[peerAddr2] = addrIndex
		}

		dijstraGraph.AddEdge(gPeerAddress2Index[peerAddr1], gPeerAddress2Index[peerAddr2], 100)
		dijstraGraph.AddEdge(gPeerAddress2Index[peerAddr2], gPeerAddress2Index[peerAddr1], 100)*/
	}
	channelviews := make(map[common.Address]map[common.Address]*ChannelView)

	twork = &TokenNetwork{
		TokenNetworkAddress:   tokenNetworkAddress,
		ChannelID2Address:     channelID2Address,
		//PeerRelationshipGraph: dijstraGraph,
		MaxRelativeFee:        0,
		db:                    db,
		channelViews:          channelviews,
		//GPeerAddress2Index:    gPeerAddress2Index,
		//CacheRemoveEdge:       make(map[int]int),
	}
	return
}

// HandleChannelOpenedEvent Handle ChannelOpened Event
func (twork *TokenNetwork) HandleChannelOpenedEvent(tokenNetwork common.Address, channelID common.Hash, participant1, participant2 common.Address) (err error) {

	var participant = [2]common.Address{participant1, participant2}

	if _, exist := storage.TokenNetwork2TokenMap[tokenNetwork]; !exist {
		err = fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s", tokenNetwork)
		return
	}
	token := storage.TokenNetwork2TokenMap[tokenNetwork]
	cview1 := InitChannelView(token, channelID, participant1, participant2, big.NewInt(0), StateChannelOpen, nil, twork.db)
	cview2 := InitChannelView(token, channelID, participant2, participant1, big.NewInt(0), StateChannelOpen, nil, twork.db)

	/*if _, exist := twork.ChannelID2Address[channelID]; !exist {
		addrIndex := 0
		addrIndex = len(twork.GPeerAddress2Index)
		if _, exist := twork.GPeerAddress2Index[participant1]; !exist {
			addrIndex++
			twork.GPeerAddress2Index[participant1] = addrIndex
		}
		if _, exist := twork.GPeerAddress2Index[participant2]; !exist {
			addrIndex++
			twork.GPeerAddress2Index[participant2] = addrIndex
		}
		twork.PeerRelationshipGraph.AddEdge(twork.GPeerAddress2Index[participant1], twork.GPeerAddress2Index[participant2], 100)
		twork.PeerRelationshipGraph.AddEdge(twork.GPeerAddress2Index[participant2], twork.GPeerAddress2Index[participant1], 100)
	}*/
	//cache channel->participants
	twork.ChannelID2Address[channelID] = participant

	err = cview1.UpdateCapacity(0, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0))
	err = cview2.UpdateCapacity(0, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0))

	return
}

// HandleChannelDepositEvent Handle Channel Deposit Event
func (twork *TokenNetwork) HandleChannelDepositEvent(tokenNetwork common.Address, channelID common.Hash, partner common.Address, totalDeposit *big.Int) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err = fmt.Errorf("Received ChannelDeposit event for unknown channel %s", channelID.String())
		return
	}

	if _, exist := storage.TokenNetwork2TokenMap[tokenNetwork]; !exist {
		err = fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s", tokenNetwork)
		return
	}
	token := storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants = twork.ChannelID2Address[channelID]
	participant1 := participants[0]
	participant2 := participants[1]

	cview1 := InitChannelView(token, channelID, participant1, participant2, totalDeposit, StateChannelDeposit, nil, twork.db)
	cview2 := InitChannelView(token, channelID, participant2, participant1, totalDeposit, StateChannelDeposit, nil, twork.db)

	if partner == participant1 {
		err = cview1.UpdateCapacity(0, totalDeposit, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else if partner == participant2 {
		err = cview2.UpdateCapacity(0, totalDeposit, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else {
		err = fmt.Errorf("Partner in ChannelDeposit does not fit the internal channel", channelID.String())
	}
	return
}

// HandleChannelClosedEvent Handle Channel Closed Event
func (twork *TokenNetwork) HandleChannelClosedEvent(tokenNetwork common.Address, channelID common.Hash) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err = fmt.Errorf("Received ChannelClosed event for unknown channel %s", channelID.String())
		return
	}

	if _, exist := storage.TokenNetwork2TokenMap[tokenNetwork]; !exist {
		err = fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s", tokenNetwork)
		return
	}
	token := storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants = twork.ChannelID2Address[channelID]
	participant1 := participants[0]
	participant2 := participants[1]

	/*twork.PeerRelationshipGraph.RemoveEdge(twork.GPeerAddress2Index[participant1], twork.GPeerAddress2Index[participant2])
	twork.PeerRelationshipGraph.RemoveEdge(twork.GPeerAddress2Index[participant2], twork.GPeerAddress2Index[participant1])*/
	//标记通道禁用
	cview1 := InitChannelView(token, channelID, participant1, participant2, big.NewInt(0), StateChannelClose, nil, twork.db)
	cview2 := InitChannelView(token, channelID, participant2, participant1, big.NewInt(0), StateChannelClose, nil, twork.db)

	err = cview1.UpdateCapacity(0, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0))
	err = cview2.UpdateCapacity(0, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0))

	return
}

// HandleChannelWithdrawEvent Handle Channel Withdaw Event
func (twork *TokenNetwork) HandleChannelWithdrawEvent(tokenNetwork common.Address, channelID common.Hash,
	participant1, participant2 common.Address, participant1Balance, participant2Balance *big.Int,
) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err = fmt.Errorf("Received ChannelClosed event for unknown channel %s", channelID.String())
		return
	}

	if _, exist := storage.TokenNetwork2TokenMap[tokenNetwork]; !exist {
		err = fmt.Errorf("Unknown TokenNetwork,TokenNetwork=%s", tokenNetwork)
		return
	}
	token := storage.TokenNetwork2TokenMap[tokenNetwork]

	var participants = twork.ChannelID2Address[channelID]
	participant01 := participants[0]
	participant02 := participants[1]
	//初始不知道哪一方取钱
	cview1 := InitChannelView(token, channelID, participant1, participant2, big.NewInt(0), StateChannelWithdraw, nil, twork.db)
	cview2 := InitChannelView(token, channelID, participant2, participant1, big.NewInt(0), StateChannelWithdraw, nil, twork.db)

	if participant1 == participant01 {
		err = cview1.UpdateCapacity(0, participant1Balance, big.NewInt(0), big.NewInt(0), big.NewInt(0))
		err = cview2.UpdateCapacity(0, participant2Balance, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else if participant1 == participant02 {
		err = cview1.UpdateCapacity(0, participant2Balance, big.NewInt(0), big.NewInt(0), big.NewInt(0))
		err = cview2.UpdateCapacity(0, participant1Balance, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	} else {
		err = fmt.Errorf("Partner in ChannelDeposit event does not fit the internal channel %s", channelID.String())
	}
	return
}

// UpdateBalance Update Balance
func (twork *TokenNetwork) UpdateBalance(
	channelID common.Hash,
	signer common.Address, //signer=谁的balance proof
	nonce uint64,
	transferredAmount *big.Int,
	lockedAmount *big.Int,
) (err error) {
	var partner common.Address
	participant1 := twork.ChannelID2Address[channelID][0]
	participant2 := twork.ChannelID2Address[channelID][1]

	token, err := twork.db.GetTokenByChannelID(nil, channelID.String())
	if err != nil {
		err = fmt.Errorf("An error occurred while obtaining token by channelID,err=%s", err)
		return
	}
	if token == "" {
		err = fmt.Errorf("No token can be querying by channelID,channelID=%s", channelID)
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

	var oldNonce uint64
	oldNonce, err = twork.db.GetLastNonceByChannel(nil, channelID.String(), signer.String(), partner.String())
	if err != nil {
		return fmt.Errorf("Can not validate nonce(internal error),nonce=%s", nonce)
	}
	//token和通道是一一对应的
	cview := InitChannelView(common.HexToAddress(token), channelID, signer, partner, big.NewInt(0), StateUpdateBalance, nil, twork.db)

	if nonce <= oldNonce {
		return fmt.Errorf("Outdated balance proof")
	}

	err = cview.UpdateCapacity(
		nonce,
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
		lockedAmount,
	)
	if err != nil {
		logrus.Error("Update balance proof error,err=", err.Error())
		return fmt.Errorf("Update balance proof error,err=%s", err.Error())
	}
	return
}

// pathResult is the json response for GetPaths
type pathResult struct {
	PathID  int              `json:"path_id"`
	PathHop int              `json:"path_hop"`
	Fee     float64          `json:"fee"`
	Result  []common.Address `json:"result"`
}

// GetPaths get shortest path
func (twork *TokenNetwork) GetPaths(source common.Address,target common.Address,
	value *big.Int,limitPaths int,sortDemand string,
) (pathinfos []interface{}, err error) {
	//todo 1\移除余额不够的边,2\移除节点不在线所处的通道,3\移除节点类型是手机的节点所处的通道matrix,4\移除节点不在线所处的所有通道matrix
	//检索出所有的节点的数据
	latestJudgements, err := twork.db.GetLatestFeeJudge(nil)
	if err != nil {
		return nil, fmt.Errorf("Can not get peer graph's data,err=%s", err)
	}

	djGraph := *dijkstra.NewEmptyGraph()
	gPeerToIndex := make(map[common.Address]int)
	gIndex:=-1
	//作图，作图是把本次计算不符合上述条件的移除掉
	for _, peerData := range latestJudgements {
		//该节点所处通道是关闭状态的
		if peerData.ChannelStatus==StateChannelClose{
			continue
		}
		//===========================================
		var peerBalance0 int64
		var peerBalance1 int64
		if peerData.PeerAddr == peerData.Participant1 {
			peerBalance0=peerData.P1Balance
			peerBalance1=peerData.P2Balance
		} else {
			peerBalance0=peerData.P2Balance
			peerBalance1=peerData.P1Balance
		}
		//该节点所处通道的余额不够
		if peerBalance0<value.Int64(){
			continue
		}
		peerHex:=common.HexToAddress(peerData.PeerAddr)
		//===========================================
		if peerData.PeerAddr == peerData.Participant1 {
			if _,exist:=gPeerToIndex[peerHex];!exist {
				gIndex++
				gPeerToIndex[peerHex] = gIndex
			}
			if _,exist:=gPeerToIndex[common.HexToAddress(peerData.Participant2)];!exist {
				gIndex++
				gPeerToIndex[common.HexToAddress(peerData.Participant2)] = gIndex
			}
			djGraph.AddEdge(gPeerToIndex[peerHex],gPeerToIndex[common.HexToAddress(peerData.Participant2)],int(peerBalance0))
			djGraph.AddEdge(gPeerToIndex[common.HexToAddress(peerData.Participant2)],gPeerToIndex[peerHex],int(peerBalance1))
		}else {
			if _,exist:=gPeerToIndex[peerHex];!exist {
				gIndex++
				gPeerToIndex[peerHex] = gIndex
			}
			if _,exist:=gPeerToIndex[common.HexToAddress(peerData.Participant1)];!exist {
				gIndex++
				gPeerToIndex[common.HexToAddress(peerData.Participant1)] = gIndex
			}
			djGraph.AddEdge(gPeerToIndex[peerHex],gPeerToIndex[common.HexToAddress(peerData.Participant1)],int(peerBalance0))
			djGraph.AddEdge(gPeerToIndex[common.HexToAddress(peerData.Participant1)],gPeerToIndex[peerHex],int(peerBalance0))
		}
	}
	if _,exist:=gPeerToIndex[source];!exist{
		return nil, fmt.Errorf("There is no suitable path")
	}
	if _,exist:=gPeerToIndex[target];!exist{
		return nil, fmt.Errorf("There is no suitable path")
	}
	xsource := gPeerToIndex[source]
	xtarget := gPeerToIndex[target]
	djResult := djGraph.AllShortestPath(xsource, xtarget)
	if djResult==nil{
		return nil, fmt.Errorf("There is no suitable path")
	}
	for k, pathSlice := range djResult {
		sinPathInfo := &pathResult{}
		sinPathInfo.PathID = k
		sinPathInfo.PathHop = len(pathSlice) - 2
		var xaddr []common.Address
		var totalfeerates float64
		// ignore peer_from and peer_to from result
		pathSlice = removePeer(pathSlice, 0)
		pathSlice = removePeer(pathSlice, len(pathSlice)-1)
		for _, peerIndex := range pathSlice {
			for addr, index := range gPeerToIndex {
				if index == peerIndex {
					xaddr = append(xaddr, addr)
					//calc fee_rate per peer
					for _, v := range latestJudgements {
						if v.PeerAddr == addr.String() {
							xfee, err := strconv.ParseFloat(v.FeeRate, 32)
							if err != nil {
								return nil, fmt.Errorf("Formatting error(fee_rate per peer in path)")
							}
							totalfeerates += xfee
							break
						}
					}
					break
				}
			}
		}
		sinPathInfo.Result = xaddr
		valuef, err := strconv.ParseFloat(value.String(), 32)
		if err != nil {
			return nil, fmt.Errorf("Formatting error(value per peer in path)")
		}
		sinPathInfo.Fee = valuef * totalfeerates
		pathinfos = append(pathinfos, sinPathInfo)
	}
	return
}

// removePeer remove source and target peer from best-path result
func removePeer(s []int, i int) []int {
	return append(s[0:i], s[i+1:]...)
}
