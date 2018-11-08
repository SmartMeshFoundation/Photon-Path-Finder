package blockchainlistener

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/kataras/go-errors"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nkbai/dijkstra"
)

type channel struct {
	Participant1        common.Address
	Participant2        common.Address
	Participant1Balance *big.Int
	Participant2Balance *big.Int
	Participant1Fee     *model.Fee
	Participant2Fee     *model.Fee
}
type nodeStatus struct {
	isMobile bool
	isOnline bool
}

// TokenNetwork token network view
type TokenNetwork struct {
	TokenNetworkAddress common.Address
	channelViews        map[common.Address][]*channel //token to channels
	channels            map[common.Hash]*channel      //channel id to chann
	token2TokenNetwork  map[common.Address]common.Address
	tokenNetwork2Token  map[common.Address]common.Address
	decimals            map[common.Address]int
	viewlock            sync.RWMutex
	participantStatus   map[common.Address]nodeStatus
	nodeLock            sync.Mutex
}

// NewTokenNetwork token network initialization
func NewTokenNetwork(token2TokenNetwork map[common.Address]common.Address) (twork *TokenNetwork) {
	//read channel view from db
	twork = &TokenNetwork{
		channelViews:       make(map[common.Address][]*channel),
		channels:           make(map[common.Hash]*channel),
		token2TokenNetwork: make(map[common.Address]common.Address),
		tokenNetwork2Token: make(map[common.Address]common.Address),
		decimals:           make(map[common.Address]int),
		participantStatus:  make(map[common.Address]nodeStatus),
	}
	for t, tn := range token2TokenNetwork {
		twork.token2TokenNetwork[t] = tn
		twork.tokenNetwork2Token[tn] = t
	}
	for token := range twork.token2TokenNetwork {
		cs, err := model.GetAllTokenChannels(token)
		if err != nil {
			panic(err)
		}
		var cs2 []*channel
		for _, c := range cs {
			c2 := &channel{
				Participant1:        common.HexToAddress(c.Participants[0].Participant),
				Participant2:        common.HexToAddress(c.Participants[1].Participant),
				Participant1Balance: c.Participants[0].BalanceValue(),
				Participant2Balance: c.Participants[1].BalanceValue(),
				Participant1Fee:     c.Participants[0].Fee(),
				Participant2Fee:     c.Participants[1].Fee(),
			}
			cs2 = append(cs2, c2)
			twork.channels[common.HexToHash(c.ChannelID)] = c2
		}
		twork.channelViews[token] = cs2
	}
	nodes := model.GetAllNodes()
	for _, n := range nodes {
		twork.participantStatus[common.HexToAddress(n.Address)] = nodeStatus{
			isOnline: n.IsOnline,
			isMobile: n.DeviceType == "mobile",
		}
	}
	return
}

// handleChannelOpenedEvent Handle ChannelOpened Event
func (t *TokenNetwork) handleChannelOpenedEvent(tokenNetwork common.Address, channelID common.Hash, participant1, participant2 common.Address, blockNumber int64) (err error) {

	token := t.tokenNetwork2Token[tokenNetwork]
	if token == utils.EmptyAddress {
		return fmt.Errorf("tokennetwork %s is unknown", tokenNetwork.String())
	}
	if t.channels[channelID] != nil {
		return fmt.Errorf("channel open duplicate for %s", channelID.String())
	}
	c, err := model.AddChannel(token, participant1, participant2, channelID, blockNumber)
	if err != nil {
		return
	}
	c2 := &channel{
		Participant1:        common.HexToAddress(c.Participants[0].Participant),
		Participant2:        common.HexToAddress(c.Participants[1].Participant),
		Participant1Balance: big.NewInt(0),
		Participant2Balance: big.NewInt(0),
		Participant1Fee:     c.Participants[0].Fee(),
		Participant2Fee:     c.Participants[1].Fee(),
	}
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	cs := t.channelViews[token]
	cs = append(cs, c2)
	t.channelViews[token] = cs
	t.channels[channelID] = c2
	return
}
func (t *TokenNetwork) handleTokenNetworkAdded(token, tokenNetwork common.Address, blockNumber int64, decimal uint8) (err error) {
	t.tokenNetwork2Token[tokenNetwork] = token
	t.token2TokenNetwork[token] = tokenNetwork
	t.decimals[token] = int(decimal)
	err = model.AddTokeNetwork(token, tokenNetwork, blockNumber)
	return nil
}

func (t *TokenNetwork) handleChannelSettled(tokenNetwork common.Address, channelID common.Hash) (err error) {
	_, err = model.SettleChannel(channelID)
	return //do nothing ,already removed when closed
}
func (t *TokenNetwork) doRemoveChannel(tokenNetwork common.Address, channelID common.Hash) (err error) {
	token := t.tokenNetwork2Token[tokenNetwork]
	if token == utils.EmptyAddress {
		return fmt.Errorf("unknown token network %s", tokenNetwork)
	}
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	c := t.channels[channelID]
	if c == nil {
		return fmt.Errorf("handleChannelCooperativeSettled but channel %s not found", channelID.String())
	}
	delete(t.channels, channelID)
	cs := t.channelViews[token]
	//c must not be nil
	for k, v := range cs {
		if v == c {
			cs = append(cs[:k], cs[k+1:]...)
		}
	}
	t.channelViews[token] = cs
	return
}
func (t *TokenNetwork) handleChannelCooperativeSettled(tokenNetwork common.Address, channelID common.Hash) (err error) {
	_, err = model.SettleChannel(channelID)
	if err != nil {
		return
	}
	return t.doRemoveChannel(tokenNetwork, channelID)
}

// handleChannelDepositEvent Handle Channel Deposit Event
func (t *TokenNetwork) handleChannelDepositEvent(tokenNetwork common.Address, channelID common.Hash, partner common.Address, totalDeposit *big.Int) (err error) {
	c, err := model.UpdateChannelDeposit(channelID, partner, totalDeposit)
	if err != nil {
		return
	}
	c2 := t.channels[channelID]
	if c2 == nil {
		log.Error(fmt.Sprintf("deposit ,but channel %s not found", channelID.String()))
		return errors.New("channel not found")
	}
	c2.Participant1Balance = c.Participants[0].BalanceValue()
	c2.Participant2Balance = c.Participants[1].BalanceValue()
	return
}

// handleChannelClosedEvent Handle Channel Closed Event
func (t *TokenNetwork) handleChannelClosedEvent(tokenNetwork common.Address, channelID common.Hash) (err error) {
	_, err = model.CloseChannel(channelID)
	if err != nil {
		return
	}
	return t.doRemoveChannel(tokenNetwork, channelID)
}

// handleChannelWithdrawEvent Handle Channel Withdaw Event
func (t *TokenNetwork) handleChannelWithdrawEvent(tokenNetwork common.Address, channelID common.Hash,
	participant1, participant2 common.Address, participant1Balance, participant2Balance *big.Int, blockNumber int64) (err error) {
	c, err := model.WithDrawChannel(channelID, participant1, participant2, participant1Balance, participant2Balance, blockNumber)
	if err != nil {
		return
	}
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	c2 := t.channels[channelID]
	if c2 == nil {
		return fmt.Errorf("withdraw event for channel %s unkown", channelID.String())
	}
	c2.Participant1Balance = c.Participants[0].BalanceValue()
	c2.Participant2Balance = c.Participants[1].BalanceValue()
	return
}

// UpdateBalance Update Balance
func (t *TokenNetwork) UpdateBalance(participant, partner common.Address, lockedAmount *big.Int, partnerBalanceProof *model.BalanceProof) (err error) {
	c, err := model.UpdateChannelBalanceProof(participant, partner, lockedAmount, partnerBalanceProof)
	if err != nil {
		return
	}
	c2 := t.channels[partnerBalanceProof.ChannelID]
	if c2 == nil {
		return fmt.Errorf("update balance proof,but channel %s unkown", partnerBalanceProof.ChannelID.String())
	}
	c2.Participant1Balance = c.Participants[0].BalanceValue()
	c2.Participant2Balance = c.Participants[1].BalanceValue()
	return
}

// PathResult is the json response for GetPaths
type PathResult struct {
	PathID  int              `json:"path_id"`  //从0开始
	PathHop int              `json:"path_hop"` //中间有多少跳,不计入源,目的节点
	Fee     *big.Int         `json:"fee"`
	Result  []common.Address `json:"result"`
}

// GetPaths get the lowest fee  path
func (t *TokenNetwork) GetPaths(source common.Address, target common.Address, tokenAddress common.Address,
	value *big.Int, limitPaths int, sortDemand string) (pathinfos []*PathResult, err error) {
	//todo 1\移除余额不够的边,2\移除节点不在线所处的通道,3\移除节点类型是手机的节点所处的通道matrix,4\移除节点不在线所处的所有通道matrix,5\移除节点网络状态为不在线的matrix
	t.viewlock.RLock()
	cs, ok := t.channelViews[tokenAddress]
	t.viewlock.RUnlock()
	if !ok {
		err = fmt.Errorf("unkown token %s", tokenAddress.String())
		return
	}
	//fmt.Println(fmt.Sprintf("-->s%",utils.StringInterface(latestJudgements,2)))
	djGraph := *dijkstra.NewEmptyGraph()
	gPeerToIndex := make(map[common.Address]int)
	gIndex := -1
	//作图，作图是把本次计算不符合上述条件的移除掉
	for _, c := range cs {
		p1Balance := c.Participant1Balance
		p2Balance := c.Participant2Balance
		t.nodeLock.Lock()
		//忽略所有不在线的节点
		if !t.participantStatus[c.Participant1].isOnline {
			t.nodeLock.Unlock()
			continue
		}
		if !t.participantStatus[c.Participant2].isOnline {
			t.nodeLock.Unlock()
			continue
		}
		//手机节点不能作为路由中间结点
		if t.participantStatus[c.Participant1].isMobile && c.Participant1 != source && c.Participant1 != target {
			t.nodeLock.Unlock()
			continue
		}
		t.nodeLock.Unlock()
		//只要有一个节点余额够,那么至少应该加入一条边
		if p1Balance.Cmp(value) < 0 && p2Balance.Cmp(value) < 0 {
			continue
		} else {
			if _, exist := gPeerToIndex[c.Participant1]; !exist {
				gIndex++
				gPeerToIndex[c.Participant1] = gIndex
			}
			if _, exist := gPeerToIndex[c.Participant2]; !exist {
				gIndex++
				gPeerToIndex[c.Participant2] = gIndex
			}
		}
		//有可能是双向的,有可能是单向的,根据金额来决定
		if p1Balance.Cmp(value) >= 0 {
			weight := t.getWeight(tokenAddress, c.Participant1Fee, value)
			if c.Participant1 == source {
				weight = 0
			}
			djGraph.AddEdge(gPeerToIndex[c.Participant1], gPeerToIndex[c.Participant2], weight) //int(peerBalance0)
		}
		if p2Balance.Cmp(value) >= 0 {
			weight := t.getWeight(tokenAddress, c.Participant2Fee, value)
			if c.Participant2 == source {
				weight = 0
			}
			djGraph.AddEdge(gPeerToIndex[c.Participant2], gPeerToIndex[c.Participant1], weight)
		}
	}
	if _, exist := gPeerToIndex[source]; !exist {
		return nil, errors.New("There is no suitable path")
	}
	if _, exist := gPeerToIndex[target]; !exist {
		return nil, errors.New("There is no suitable path")
	}
	xsource := gPeerToIndex[source]
	xtarget := gPeerToIndex[target]
	djResult := djGraph.AllShortestPath(xsource, xtarget)
	if djResult == nil {
		return nil, errors.New("There is no suitable path")
	}
	//log.Trace(fmt.Sprintf("result=%s,index=%s", utils.StringInterface(djResult, 5), utils.StringInterface(gPeerToIndex, 3)))
	//将所有可能的最短路径转换为Address结果,同时计算费用
	for k, pathSlice := range djResult {
		sinPathInfo := &PathResult{}
		sinPathInfo.PathID = k
		sinPathInfo.PathHop = len(pathSlice) - 2
		var xaddr []common.Address
		var totalfeerates = new(big.Int)
		var lastAddress common.Address

		//跳过源节点,他是不会收费的
		for i := 1; i < len(pathSlice)-1; i++ {
			var p1, p2 common.Address
			foundNumber := 0
			for addr, index := range gPeerToIndex {
				if index == pathSlice[i] {
					p1 = addr
					foundNumber++
					if foundNumber >= 2 {
						break
					}
				} else if index == pathSlice[i+1] {
					p2 = addr
					lastAddress = p2
					foundNumber++
					if foundNumber >= 2 {
						break
					}
				}
			}
			//我们确定能找到
			xaddr = append(xaddr, p1)
			xfee := t.calcFeeByParticipantPartner(tokenAddress, p1, p2, value)
			totalfeerates = totalfeerates.Add(totalfeerates, xfee)
		}
		//把Target添加到路由列表中
		xaddr = append(xaddr, lastAddress)
		sinPathInfo.Fee = totalfeerates
		sinPathInfo.Result = xaddr
		pathinfos = append(pathinfos, sinPathInfo)
	}
	return
}
func calcFee(value *big.Int, fee *model.Fee) (w *big.Int) {
	w = new(big.Int)
	if fee.FeePercent > 0 {
		w.Set(value)
		w = w.Div(w, big.NewInt(fee.FeePercent))
	}
	if fee.FeeConstant.Cmp(utils.BigInt0) > 0 {
		w = w.Add(w, fee.FeeConstant)
	}
	return w
}

var weightPrecision int64 = 100000
var weightLog = int(math.Log10(float64(weightPrecision)))

/*
将收费信息转换为图中边的权重问题,
基本思路是:
收费既不会太高也不会太低
比如十万分之一的收费应该是下线了,
同样,raiden是一个小额支付系统,金额也不会太高,太高的交易多半会失败.
因此在这些限制之下, 是有可能把收费金额转换为一个int来表示的.
*/
func (t *TokenNetwork) getWeight(token common.Address, fee *model.Fee, value *big.Int) (weight int) {
	w := calcFee(value, fee)
	decimal := t.decimals[token]
	if decimal > weightLog {
		//decimal是一般的18,或者22之类的
		decimal -= weightLog
		v := big.NewInt(int64(math.Pow10(int(decimal))))
		w = w.Div(w, v)
	}
	maxWeight := big.NewInt(int64(math.MaxInt32))
	if w.Cmp(maxWeight) > 0 {
		log.Error(fmt.Sprintf("weight overflow token=%s,value=%s,feeconstant=%s,feepercent=%d",
			token.String(), value, fee.FeeConstant, fee.FeePercent))
		return math.MaxInt32
	}
	weight = int(w.Int64())
	return
}

//注意与合约上计算方式保持完全一致.
func calcChannelID(tokenNetwork, p1, p2 common.Address) common.Hash {
	var channelID common.Hash
	//log.Trace(fmt.Sprintf("p1=%s,p2=%s,tokennetwork=%s", p1.String(), p2.String(), tokenNetwork.String()))
	if bytes.Compare(p1[:], p2[:]) < 0 {
		channelID = utils.Sha3(p1[:], p2[:], tokenNetwork[:])
	} else {
		channelID = utils.Sha3(p2[:], p1[:], tokenNetwork[:])
	}
	return channelID
}

// calcFeeByParticipantPartner get fee_rate when the peer in some channel
func (t *TokenNetwork) calcFeeByParticipantPartner(token, p1, p2 common.Address, value *big.Int) (xfee *big.Int) {
	channelID := calcChannelID(t.token2TokenNetwork[token], p1, p2)
	c := t.channels[channelID]
	if c == nil {
		//todo fixme 在发布的时候应该替换为返回0,并记录错误
		panic(fmt.Sprintf("channel not found,p1=%s,p2=%s,token=%s", p1.String(), p2.String(), token.String()))
	}
	var fee *model.Fee
	if p1 == c.Participant1 {
		fee = c.Participant1Fee
	} else if p1 == c.Participant2 {
		fee = c.Participant2Fee
	} else {
		panic(fmt.Sprintf("channel error channleid=%s,p1=%s,p2=%s", channelID.String(), p1.String(), p2.String()))
	}
	return calcFee(value, fee)
}

//Online implements MatrixPresenceListener
func (t *TokenNetwork) Online(address common.Address, deviceType string) {
	t.nodeLock.Lock()
	defer t.nodeLock.Unlock()
	t.participantStatus[address] = nodeStatus{
		isMobile: deviceType == "mobile",
		isOnline: true,
	}
	model.NewOrUpdateNodeStatus(address, true, deviceType)
}

//Offline implements MatrixPresenceListener
func (t *TokenNetwork) Offline(address common.Address) {
	t.nodeLock.Lock()
	defer t.nodeLock.Unlock()
	t.participantStatus[address] = nodeStatus{
		isOnline: false,
	}
	model.NewOrUpdateNodeOnline(address, false)
}

//UpdateChannelFeeRate set channel fee rate
func (t *TokenNetwork) UpdateChannelFeeRate(channelID common.Hash, peerAddress common.Address, fee *model.Fee) error {
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	c, ok := t.channels[channelID]
	if !ok {
		return fmt.Errorf("channel %s not found", channelID.String())
	}
	if c.Participant1 == peerAddress {
		c.Participant1Fee = fee
	} else if c.Participant2 == peerAddress {
		c.Participant2Fee = fee
	} else {
		return fmt.Errorf("peer %s not match channel %s", peerAddress.String(), channelID.String())
	}
	_, err := model.UpdateChannelFeeRate(channelID, peerAddress, fee)
	return err
}
