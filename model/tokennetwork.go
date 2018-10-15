package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
	"encoding/binary"
	"fmt"
	"math/big"
)

//TokenNetwork token network view
type TokenNetwork struct {
	TokenNetworkAddress common.Address
	ChannelID2Address map[common.Hash][2]common.Address
	PeerRelationshipGraph dijkstra.Graph
	MaxRelativeFee int64
}

//InitTokenNetwork token network initialization
func InitTokenNetwork(tokenNetworkAddress common.Address) (*TokenNetwork) {
	twork := &TokenNetwork{
		TokenNetworkAddress:   tokenNetworkAddress,
		ChannelID2Address:     make(map[common.Hash][2]common.Address),
		PeerRelationshipGraph: *dijkstra.NewEmptyGraph(),
		MaxRelativeFee:        0,
	}
	return twork
}
//var ChannelID2Address map[common.Hash][2]common.Address

//HandleChannelOpenedEvent 处理通道打开事件
func (twork TokenNetwork)HandleChannelOpenedEvent(channelID common.Hash,participant1,participant2 common.Address) (err error) {
	if common.IsHexAddress(participant1.String()) {
		return
	}
	if common.IsHexAddress(participant2.String()) {
		return
	}

	var participant = [2]common.Address{participant1, participant2}
	twork.ChannelID2Address[channelID] = participant


	cview1:=InitChannelView(channelID, participant1, participant2, big.NewInt(0),StateChannelOpen)
	cview2:=InitChannelView(channelID, participant2, participant1, big.NewInt(0),StateChannelOpen)

	twork.PeerRelationshipGraph.AddEdge(BytesToInt(participant1.Bytes()),BytesToInt(participant2.Bytes()),100)
	twork.PeerRelationshipGraph.AddEdge(BytesToInt(participant2.Bytes()),BytesToInt(participant1.Bytes()),100)

	cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return nil
}

//HandleChannelDeposit 处理通道存钱事件
func (twork TokenNetwork)HandleChannelDeposit(channelID common.Hash,partner common.Address,totalDeposit *big.Int) (err error) {
	if common.IsHexAddress(partner.String()) {
		return
	}
	_, exist := twork.ChannelID2Address[channelID]
	if !exist{
		fmt.Errorf("Received ChannelClosed event for unknown channel %s",channelID.String())
		return
	}

	var participants = twork.ChannelID2Address[channelID]
	participant1:=participants[0]
	participant2:=participants[1]

	cview1:=InitChannelView(channelID, participant1, participant2, totalDeposit,StateChannelDeposit)
	cview2:=InitChannelView(channelID, participant2, participant1, totalDeposit,StateChannelDeposit)

	if partner==participant1{
		cview1.UpdateCapacity(0,totalDeposit,big.NewInt(0),big.NewInt(0),big.NewInt(0))
	}else if partner==participant2{
		cview2.UpdateCapacity(0,totalDeposit,big.NewInt(0),big.NewInt(0),big.NewInt(0))

	}else {
		fmt.Errorf("Partner in ChannelDeposit does not fit the internal channel",channelID.String())
	}
	return nil
}

//HandleChannelClosedEvent 处理通道关闭事件
func (twork TokenNetwork)HandleChannelClosedEvent(channelID common.Hash) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist{
		fmt.Errorf("Received ChannelClosed event for unknown channel %s",channelID.String())
		return
	}
	var participants = twork.ChannelID2Address[channelID]
	participant1:=participants[0]
	participant2:=participants[1]
	twork.PeerRelationshipGraph.RemoveEdge(BytesToInt(participant1.Bytes()),BytesToInt(participant2.Bytes()))
	twork.PeerRelationshipGraph.RemoveEdge(BytesToInt(participant2.Bytes()),BytesToInt(participant1.Bytes()))

	//标记通道禁用
	cview1:=InitChannelView(channelID, participant1, participant2, big.NewInt(0),StateChannelClose)
	cview2:=InitChannelView(channelID, participant2, participant1, big.NewInt(0),StateChannelClose)

	cview1.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))
	cview2.UpdateCapacity(0,big.NewInt(0),big.NewInt(0),big.NewInt(0),big.NewInt(0))

	return
}

//HandleChannelWithdawEvent
func (twork TokenNetwork)HandleChannelWithdawEvent(channelID common.Hash,
	participant1,participant2 common.Address,participant1Balance,participant2Balance *big.Int,
	) (err error) {
	_, exist := twork.ChannelID2Address[channelID]
	if !exist {
		fmt.Errorf("Received ChannelClosed event for unknown channel %s", channelID.String())
		return
	}

	var participants = twork.ChannelID2Address[channelID]
	participant01 := participants[0]
	participant02 := participants[1]
	//初始不知道哪一方取钱
	totalBalance:=participant1Balance.Add(participant1Balance,participant2Balance)
	cview1:=InitChannelView(channelID, participant1, participant2, totalBalance,StateChannelDeposit)
	cview2:=InitChannelView(channelID, participant2, participant1, totalBalance,StateChannelDeposit)

	if participant1 == participant01 {
		cview1.UpdateCapacity(0,totalBalance.Sub(totalBalance,participant2Balance),big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		cview2.UpdateCapacity(0,totalBalance.Sub(totalBalance,participant1Balance),big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else if participant1 == participant02 {
		cview1.UpdateCapacity(0,totalBalance.Sub(totalBalance,participant1Balance),big.NewInt(0),big.NewInt(0),big.NewInt(0),)
		cview2.UpdateCapacity(0,totalBalance.Sub(totalBalance,participant2Balance),big.NewInt(0),big.NewInt(0),big.NewInt(0),)
	} else {
		fmt.Errorf("Partner in ChannelDeposit does not fit the internal channel", channelID.String())
	}
	return nil
}

//UpdateBalance 更新余额
func (twork TokenNetwork)UpdateBalance(
	channelID common.Hash,
	singer common.Address,
	nonce int64,
	transferredAmount *big.Int,
	lockedAmount *big.Int,
	) {
	var partner common.Address
	participant1 := twork.ChannelID2Address[channelID][0]
	participant2 := twork.ChannelID2Address[channelID][1]

	if singer == participant1 {
		partner = participant2
	} else if singer == participant2 {
		partner = participant1
	} else {
		//dayin
	}
	//获取图
	cview1 := &ChannelView{
		SelfAddress: partner,
	}
	cview2 := &ChannelView{
		SelfAddress: singer,
	}
	//更新通道双方的Capacity
	cview1.UpdateCapacity(
		nonce,
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
		lockedAmount,
	)
	cview2.UpdateCapacity(
		0,
		big.NewInt(0),
		big.NewInt(0),
		transferredAmount,
		big.NewInt(0),
	)
}

func BytesToInt(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}

func IntToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
