package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
	"fmt"
	"math/big"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"encoding/binary"
	"github.com/Sirupsen/logrus"
)

// TokenNetwork token network view
type TokenNetwork struct {
	TokenNetworkAddress   common.Address
	ChannelID2Address     map[common.Hash][2]common.Address //cache key=channel_id value={0xparticipant1,0xparticipan2}
	PeerRelationshipGraph dijkstra.Graph
	MaxRelativeFee        int64
	db                    *storage.Database
	channelViews          map[common.Address]map[common.Address]*ChannelView
}

// InitTokenNetwork token network initialization
func InitTokenNetwork(tokenNetworkAddress common.Address,db *storage.Database) (twork *TokenNetwork) {
	//read channel view from db
	channelinfos, err := db.GetAllChannelHistoryStorage(nil)
	if err != nil {
		return
	}
	channelID2Address := make(map[common.Hash][2]common.Address)
	for _, channelinfo := range channelinfos {
		if channelinfo.Status != StateChannelClose {
			var participant= [2]common.Address{common.StringToAddress(channelinfo.Participant), common.StringToAddress(channelinfo.Partner)}
			channelID2Address[common.StringToHash(channelinfo.ChannelID)] = participant
		}
	}
	channelviews := make(map[common.Address]map[common.Address]*ChannelView)

	twork = &TokenNetwork{
		TokenNetworkAddress:   tokenNetworkAddress,
		ChannelID2Address:     channelID2Address, //make(map[common.Hash][2]common.Address),
		PeerRelationshipGraph: *dijkstra.NewEmptyGraph(),
		MaxRelativeFee:        0,
		db:                    db,
		channelViews:          channelviews,
	}
	return
}

// HandleChannelOpenedEvent Handle ChannelOpened Event
func (twork *TokenNetwork)HandleChannelOpenedEvent(channelID common.Hash,participant1,participant2 common.Address) (err error) {

	var participant = [2]common.Address{participant1, participant2}
	twork.ChannelID2Address[channelID] = participant

	//cview1:=twork.channelViews[participant1][participant2]

	cview1:=InitChannelView(channelID, participant1, participant2, big.NewInt(0),StateChannelOpen,twork.db)
	cview2:=InitChannelView(channelID, participant2, participant1, big.NewInt(0),StateChannelOpen,twork.db)

	twork.PeerRelationshipGraph.AddEdge(BytesToInt(participant1.Bytes()),BytesToInt(participant2.Bytes()),100)
	twork.PeerRelationshipGraph.AddEdge(BytesToInt(participant2.Bytes()),BytesToInt(participant1.Bytes()),100)

	cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return nil
}

// HandleChannelDeposit Handle Channel Deposit Event
func (twork *TokenNetwork)HandleChannelDepositEvent(channelID common.Hash,partner common.Address,totalDeposit *big.Int) (err error) {

	_, exist := twork.ChannelID2Address[channelID]
	if !exist{
		err=fmt.Errorf("Received ChannelClosed event for unknown channel %s",channelID.String())
		return
	}

	var participants = twork.ChannelID2Address[channelID]
	participant1:=participants[0]
	participant2:=participants[1]

	cview1:=InitChannelView(channelID, participant1, participant2, totalDeposit,StateChannelDeposit,twork.db)
	cview2:=InitChannelView(channelID, participant2, participant1, totalDeposit,StateChannelDeposit,twork.db)

	if partner==participant1{
		cview1.UpdateCapacity(0,totalDeposit,big.NewInt(0),big.NewInt(0),big.NewInt(0))
	}else if partner==participant2{
		cview2.UpdateCapacity(0,totalDeposit,big.NewInt(0),big.NewInt(0),big.NewInt(0))

	}else {
		err=fmt.Errorf("Partner in ChannelDeposit does not fit the internal channel",channelID.String())
	}
	return nil
}

// HandleChannelClosedEvent Handle Channel Closed Event
func (twork *TokenNetwork)HandleChannelClosedEvent(channelID common.Hash) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist{
		err=fmt.Errorf("Received ChannelClosed event for unknown channel %s",channelID.String())
		return
	}
	var participants = twork.ChannelID2Address[channelID]
	participant1:=participants[0]
	participant2:=participants[1]
	twork.PeerRelationshipGraph.RemoveEdge(BytesToInt(participant1.Bytes()),BytesToInt(participant2.Bytes()))
	twork.PeerRelationshipGraph.RemoveEdge(BytesToInt(participant2.Bytes()),BytesToInt(participant1.Bytes()))

	//标记通道禁用
	cview1:=InitChannelView(channelID, participant1, participant2, big.NewInt(0),StateChannelClose,twork.db)
	cview2:=InitChannelView(channelID, participant2, participant1, big.NewInt(0),StateChannelClose,twork.db)

	cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return
}

// HandleChannelWithdawEvent Handle Channel Withdaw Event
func (twork *TokenNetwork)HandleChannelWithdawEvent(channelID common.Hash,
	participant1,participant2 common.Address,participant1Balance,participant2Balance *big.Int,
	) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		err=fmt.Errorf("Received ChannelClosed event for unknown channel %s", channelID.String())
		return
	}

	var participants = twork.ChannelID2Address[channelID]
	participant01 := participants[0]
	participant02 := participants[1]
	//初始不知道哪一方取钱
	//totalBalance:=participant1Balance.Add(participant1Balance,participant2Balance)
	cview1:=InitChannelView(channelID, participant1, participant2, big.NewInt(0),StateChannelWithdraw,twork.db)
	cview2:=InitChannelView(channelID, participant2, participant1, big.NewInt(0),StateChannelWithdraw,twork.db)

	if participant1 == participant01 {
		cview1.UpdateCapacity(0,participant1Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		cview2.UpdateCapacity(0,participant2Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else if participant1 == participant02 {
		cview1.UpdateCapacity(0,participant2Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		cview2.UpdateCapacity(0,participant1Balance,big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else {
		err=fmt.Errorf("Partner in ChannelDeposit event does not fit the internal channel %s", channelID.String())
	}
	return nil
}

//UpdateBalance Update Balance
/*		err = ce.TokenNetwork.UpdateBalance(
			r.BalanceProof.ChannelID,
			partner,
			r.BalanceProof.Nonce,
			r.BalanceProof.TransferredAmount,
			r.LocksAmount)
*/
func (twork *TokenNetwork)UpdateBalance(
	channelID common.Hash,
	singer common.Address,
	nonce int64,
	transferredAmount *big.Int,
	lockedAmount *big.Int,
	) (err error){
	var partner common.Address
	participant1 := twork.ChannelID2Address[channelID][0]
	participant2 := twork.ChannelID2Address[channelID][1]

	if singer == participant1 {
		partner = participant2
	} else if singer == participant2 {
		partner = participant1
	} else {
		logrus.Error("Balance proof signature does not match any of the participants")
		return fmt.Errorf("Balance proof signature error")
	}

	cview1 := &ChannelView{
		SelfAddress: singer,
		PartnerAddress:partner,
	}
	cview2 := &ChannelView{
		SelfAddress: partner,
		PartnerAddress:singer,
	}
	//更新通道双方的Capacity
	err=cview1.UpdateCapacity(
		nonce,
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
		lockedAmount,
	)
	err=cview2.UpdateCapacity(
		0,
		big.NewInt(0),
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
	)
	return
}

func BytesToInt(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}

func IntToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
