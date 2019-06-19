package blockchainlistener

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"

	"errors"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/params"
	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/dijkstra"
	pparams "github.com/SmartMeshFoundation/Photon-Path-Finder/params"
	"github.com/ethereum/go-ethereum/common"
)

type channel struct {
	Participant1        common.Address
	Participant2        common.Address
	Participant1Balance *big.Int
	Participant2Balance *big.Int
	Participant1Fee     *model.Fee
	Participant2Fee     *model.Fee
	Token               common.Address
}
type nodeStatus struct {
	isMobile bool
	isOnline bool
}

// TokenNetwork token network view
type TokenNetwork struct {
	TokensNetworkAddress common.Address
	channelViews         map[common.Address][]*channel //token to channels
	channels             map[common.Hash]*channel      //channel id to chann
	token2TokenNetwork   map[common.Address]common.Address
	decimals             map[common.Address]int
	viewlock             sync.RWMutex
	participantStatus    map[common.Address]nodeStatus
	nodeLock             sync.Mutex
	transport            Transporter
}

// NewTokenNetwork token network initialization
func NewTokenNetwork(token2TokenNetwork map[common.Address]common.Address, tokensNetworkAddress common.Address, useMatrix bool, decimals map[common.Address]int) (twork *TokenNetwork) {
	//read channel view from db
	twork = &TokenNetwork{
		TokensNetworkAddress: tokensNetworkAddress,
		channelViews:         make(map[common.Address][]*channel),
		channels:             make(map[common.Hash]*channel),
		token2TokenNetwork:   make(map[common.Address]common.Address),
		decimals:             make(map[common.Address]int),
		participantStatus:    make(map[common.Address]nodeStatus),
	}
	if decimals != nil {
		twork.decimals = decimals
	}
	for t, tn := range token2TokenNetwork {
		twork.token2TokenNetwork[t] = tn
	}
	if useMatrix {
		twork.transport = NewMatrixObserver(twork)
	} else {
		var err error
		twork.transport, err = NewXMPPConnection(params.DefaultXMPPServer, dbXMPPWrapper{}, twork)
		if err != nil {
			log.Crit(fmt.Sprintf("NewXMPPConnection err %s", err))
		}
	}
	for token := range twork.token2TokenNetwork {
		cs, err := model.GetAllTokenChannels(token)
		if err != nil {
			panic(err)
		}
		var cs2 []*channel
		for _, c := range cs {
			//token := common.HexToAddress(c.Token)
			c2 := &channel{
				Participant1:        common.HexToAddress(c.Participants[0].Participant),
				Participant2:        common.HexToAddress(c.Participants[1].Participant),
				Participant1Balance: c.Participants[0].BalanceValue(),
				Participant2Balance: c.Participants[1].BalanceValue(),
				Participant1Fee:     c.Participants[0].Fee(token),
				Participant2Fee:     c.Participants[1].Fee(token),
				Token:               token,
			}
			cs2 = append(cs2, c2)
			twork.channels[common.HexToHash(c.ChannelID)] = c2
			err = twork.transport.SubscribeNeighbors([]common.Address{c2.Participant1, c2.Participant2})
			if err != nil {
				log.Error(fmt.Sprintf("SubscribeNeighbors err %s", err))
			}
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
func (t *TokenNetwork) handleChannelOpenedEvent(tokenAddress common.Address, channelID common.Hash, participant1, participant2 common.Address, blockNumber int64) (err error) {
	if t.channels[channelID] != nil {
		return fmt.Errorf("channel open duplicate for %s", channelID.String())
	}
	c, err := model.AddChannel(tokenAddress, participant1, participant2, channelID, blockNumber)
	if err != nil {
		return
	}
	c2 := &channel{
		Participant1:        common.HexToAddress(c.Participants[0].Participant),
		Participant2:        common.HexToAddress(c.Participants[1].Participant),
		Participant1Balance: big.NewInt(0),
		Participant2Balance: big.NewInt(0),
		Participant1Fee:     c.Participants[0].Fee(tokenAddress),
		Participant2Fee:     c.Participants[1].Fee(tokenAddress),
		Token:               tokenAddress,
	}
	err = t.transport.SubscribeNeighbors([]common.Address{c2.Participant1, c2.Participant2})
	if err != nil {
		log.Error(fmt.Sprintf("SubscribeNeighbors err %s", err))
	}
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	_, ok := t.decimals[tokenAddress]
	if !ok {
		panic(fmt.Sprintf("uknown token %s", tokenAddress.String()))
	}
	cs := t.channelViews[tokenAddress]
	cs = append(cs, c2)
	t.channelViews[tokenAddress] = cs
	t.channels[channelID] = c2
	//log.Trace(fmt.Sprintf("handleChannelOpenedEvent token=%s, channelViews=%s", utils.APex2(tokenAddress), utils.StringInterface(cs, 5)))
	return
}
func (t *TokenNetwork) handleTokenNetworkAdded(token common.Address, blockNumber int64, decimal uint8) (err error) {

	t.token2TokenNetwork[token] = utils.EmptyAddress
	t.decimals[token] = int(decimal)
	err = model.AddTokeNetwork(token, utils.EmptyAddress, blockNumber)
	return
}

func (t *TokenNetwork) handleChannelSettled(channelID common.Hash) (err error) {
	_, err = model.SettleChannel(channelID)
	return //do nothing ,already removed when closed
}
func (t *TokenNetwork) doRemoveChannel(token common.Address, channelID common.Hash) (err error) {
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	c := t.channels[channelID]
	if c == nil {
		return fmt.Errorf("handleChannelCooperativeSettled but channel %s not found", channelID.String())
	}
	err = t.transport.Unsubscribe(c.Participant1)
	if err != nil {
		log.Error(fmt.Sprintf("Unsubscribe %s err %s", c.Participant1.String(), err))
	}
	err = t.transport.Unsubscribe(c.Participant2)
	if err != nil {
		log.Error(fmt.Sprintf("Unsubscribe %s err %s", c.Participant1.String(), err))
	}
	delete(t.channels, channelID)
	cs := t.channelViews[token]
	found := false
	//c must not be nil
	for k, v := range cs {
		if v == c {
			cs = append(cs[:k], cs[k+1:]...)
			found = true
			break
		}
	}
	if !found {
		log.Error(fmt.Sprintf("doRemoveChannel channle not found  t,oken=%s %s,", utils.APex2(token), utils.StringInterface(c, 3)))
		log.Error(fmt.Sprintf("channelViews=%s", utils.StringInterface(cs, 5)))
	}
	t.channelViews[token] = cs
	return
}
func (t *TokenNetwork) handleChannelCooperativeSettled(channelID common.Hash) (err error) {
	c, err := model.SettleChannel(channelID)
	if err != nil {
		return
	}
	return t.doRemoveChannel(common.HexToAddress(c.Token), channelID)
}

// handleChannelDepositEvent Handle Channel Deposit Event
func (t *TokenNetwork) handleChannelDepositEvent(channelID common.Hash, participant common.Address, totalDeposit *big.Int) (err error) {
	c, err := model.UpdateChannelDeposit(channelID, participant, totalDeposit)
	if err != nil {
		return
	}
	c2 := t.channels[channelID]
	if c2 == nil {
		log.Error(fmt.Sprintf("deposit ,but channel %s not found in memory,maybe closed, status=%d",
			channelID.String(), c.Status))
		return errors.New("channel not found")
	}
	c2.Participant1Balance = c.Participants[0].BalanceValue()
	c2.Participant2Balance = c.Participants[1].BalanceValue()
	return
}

// handleChannelClosedEvent Handle Channel Closed Event
func (t *TokenNetwork) handleChannelClosedEvent(channelID common.Hash) (err error) {
	c, err := model.CloseChannel(channelID)
	if err != nil {
		return
	}
	return t.doRemoveChannel(common.HexToAddress(c.Token), channelID)
}

// handleChannelWithdrawEvent Handle Channel Withdaw Event
func (t *TokenNetwork) handleChannelWithdrawEvent(channelID common.Hash,
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

func (t *TokenNetwork) checkorder(cs []*channel) {
	if pparams.DebugMode {
		for _, c := range cs {
			if c.Participant1.String() > c.Participant2.String() {
				panic(fmt.Sprintf("channel order error c=%s", utils.StringInterface(c, 3)))
			}
		}
	}
}

// GetPaths get the lowest fee  path
func (t *TokenNetwork) GetPaths(source common.Address, target common.Address, tokenAddress common.Address,
	value *big.Int, limitPaths int, sortDemand string, sourceChargeFee bool) (pathinfos []*PathResult, err error) {
	//todo 1\移除余额不够的边,2\移除节点不在线所处的通道,3\移除节点类型是手机的节点所处的通道matrix,4\移除节点不在线所处的所有通道matrix,5\移除节点网络状态为不在线的matrix
	t.viewlock.RLock()
	cs, ok := t.channelViews[tokenAddress]
	t.viewlock.RUnlock()
	t.checkorder(cs)
	if !ok {
		err = fmt.Errorf("unkown token %s", tokenAddress.String())
		return
	}
	log.Trace(fmt.Sprintf("GetPaths requests source=%s,target=%s token=%s value=%s",
		utils.APex2(source), utils.APex2(target), utils.APex2(tokenAddress), value,
	))
	log.Trace(fmt.Sprintf("channels=%s", utils.StringInterface(cs, 7)))
	log.Trace(fmt.Sprintf("nodestatus=%s", utils.StringInterface(t.participantStatus, 5)))
	start := time.Now()
	//fmt.Println(fmt.Sprintf("-->s%",utils.StringInterface(latestJudgements,2)))
	djGraph := *dijkstra.NewEmptyGraph()
	gPeerToIndex := make(map[common.Address]int)
	gIndex := -1
	//作图，作图是把本次计算不符合上述条件的移除掉
	t.nodeLock.Lock()
	for _, c := range cs {
		p1Balance := c.Participant1Balance
		p2Balance := c.Participant2Balance

		//忽略所有不在线的节点
		if !t.participantStatus[c.Participant1].isOnline {
			continue
		}
		if !t.participantStatus[c.Participant2].isOnline {
			continue
		}
		//手机节点不能作为路由中间结点
		if t.participantStatus[c.Participant1].isMobile && c.Participant1 != source && c.Participant1 != target {
			continue
		}
		//通道双方只要有一个是手机并且既不是发起方也不是接收方,都应该 跳过
		if t.participantStatus[c.Participant2].isMobile && c.Participant2 != source && c.Participant2 != target {
			continue
		}
		//只要有一个节点余额够,那么至少应该加入一条边
		if p1Balance.Cmp(value) < 0 && p2Balance.Cmp(value) < 0 {
			continue
		} else {
			if _, exist := gPeerToIndex[c.Participant1]; !exist {
				djGraph.AddVertex()
				gIndex++
				gPeerToIndex[c.Participant1] = gIndex
			}
			if _, exist := gPeerToIndex[c.Participant2]; !exist {
				djGraph.AddVertex()
				gIndex++
				gPeerToIndex[c.Participant2] = gIndex
			}
		}
		//有可能是双向的,有可能是单向的,根据金额来决定
		if p1Balance.Cmp(value) >= 0 {
			weight := t.getWeight(tokenAddress, c.Participant1Fee, value)
			if c.Participant1 == source && !sourceChargeFee {
				weight = 0
			}
			djGraph.AddEdge(gPeerToIndex[c.Participant1], gPeerToIndex[c.Participant2], weight) //int(peerBalance0)
		}
		if p2Balance.Cmp(value) >= 0 {
			weight := t.getWeight(tokenAddress, c.Participant2Fee, value)
			if c.Participant2 == source && !sourceChargeFee {
				weight = 0
			}
			djGraph.AddEdge(gPeerToIndex[c.Participant2], gPeerToIndex[c.Participant1], weight)
		}
	}
	t.nodeLock.Unlock()
	if _, exist := gPeerToIndex[source]; !exist {
		return nil, errors.New("There is no suitable path")
	}
	if _, exist := gPeerToIndex[target]; !exist {
		return nil, errors.New("There is no suitable path")
	}
	//for addr, id := range gPeerToIndex {
	//	fmt.Printf("addr=%s,id=%d\n", addr.String(), id)
	//}
	//djGraph.PrintGraph()
	xsource := gPeerToIndex[source]
	xtarget := gPeerToIndex[target]
	buildtime := time.Now()
	djResult := djGraph.AllShortestPath(xsource, xtarget, dijkstra.DefaultCostGetter)
	if djResult == nil {
		return nil, errors.New("There is no suitable path")
	}
	//log.Trace(fmt.Sprintf("result=%v", djResult))
	calcpathtime := time.Now()
	//log.Trace(fmt.Sprintf("result=%s,index=%s", utils.StringInterface(djResult, 5), utils.StringInterface(gPeerToIndex, 3)))
	//将所有可能的最短路径转换为Address结果,同时计算费用
	for k, pathSlice := range djResult {
		//限制返回结果的数量,不要太多
		if limitPaths > 0 && k >= limitPaths {
			break
		}
		sinPathInfo := &PathResult{}
		sinPathInfo.PathID = k
		sinPathInfo.PathHop = len(pathSlice) - 2
		var xaddr []common.Address
		var totalfeerates = new(big.Int)

		//要区分源节点是否收费
		var i = 1
		if sourceChargeFee {
			i = 0
		}
		for ; i < len(pathSlice)-1; i++ {
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
		var lastAddress common.Address
		for addr, index := range gPeerToIndex {
			if index == pathSlice[len(pathSlice)-1] {
				lastAddress = addr
				break
			}
		}
		if lastAddress == utils.EmptyAddress {
			err = fmt.Errorf("impossible error, no target in path")
			return
		}
		xaddr = append(xaddr, lastAddress)
		//无论源节点是否收费,都不能把源节点放到路径中去
		if sourceChargeFee && len(xaddr) > 0 {
			xaddr = xaddr[1:]
		}
		sinPathInfo.Fee = totalfeerates
		sinPathInfo.Result = xaddr
		pathinfos = append(pathinfos, sinPathInfo)
	}
	/*
		由于计算精度问题,有可能导致计算出来的fee并不一样,最好按照费用大小排序
	*/
	sort.Slice(pathinfos, func(i, j int) bool {
		return pathinfos[i].Fee.Cmp(pathinfos[j].Fee) < 0
	})
	log.Info(fmt.Sprintf("buildgraph=%s,path=%s", buildtime.Sub(start), calcpathtime.Sub(buildtime)))
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
		log.Error(fmt.Sprintf("weight overflow token=%s,value=%s,feeconstant=%s,feepercent=%d,decimal=%d",
			token.String(), value, fee.FeeConstant, fee.FeePercent, decimal))
		return math.MaxInt32
	}
	weight = int(w.Int64())
	return
}

//注意与合约上计算方式保持完全一致.
func calcChannelID(token, tokensNetwork, p1, p2 common.Address) common.Hash {
	var channelID common.Hash
	//log.Trace(fmt.Sprintf("p1=%s,p2=%s,tokennetwork=%s", p1.String(), p2.String(), tokenNetwork.String()))
	if bytes.Compare(p1[:], p2[:]) < 0 {
		channelID = utils.Sha3(p1[:], p2[:], token[:], tokensNetwork[:])
	} else {
		channelID = utils.Sha3(p2[:], p1[:], token[:], tokensNetwork[:])
	}
	return channelID
}

// calcFeeByParticipantPartner get fee_rate when the peer in some channel
func (t *TokenNetwork) calcFeeByParticipantPartner(token, p1, p2 common.Address, value *big.Int) (xfee *big.Int) {
	channelID := calcChannelID(token, t.TokensNetworkAddress, p1, p2)
	c := t.channels[channelID]
	if c == nil {
		log.Trace(fmt.Sprintf("channels=%s", utils.StringInterface(t.channels, 5)))
		log.Trace(fmt.Sprintf("token=%s,p1=%s,p2=%s,tokennetwork=%s", utils.APex2(token),
			utils.APex2(p1), utils.APex2(p2), utils.APex2(t.TokensNetworkAddress)))
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

//Online implements NodePresenceListener
func (t *TokenNetwork) Online(address common.Address, deviceType string) {
	t.nodeLock.Lock()
	defer t.nodeLock.Unlock()
	t.participantStatus[address] = nodeStatus{
		isMobile: deviceType == "mobile",
		isOnline: true,
	}
	log.Trace(fmt.Sprintf("%s online ,type=%s", address.String(), deviceType))
	model.NewOrUpdateNodeStatus(address, true, deviceType)
}

//Offline implements NodePresenceListener
func (t *TokenNetwork) Offline(address common.Address) {
	t.nodeLock.Lock()
	defer t.nodeLock.Unlock()
	t.participantStatus[address] = nodeStatus{
		isOnline: false,
	}
	log.Trace(fmt.Sprintf("%s offliine", address.String()))
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
	return model.UpdateChannelFeeRate(channelID, peerAddress, c.Token, fee)
}
//UpdateAccountFee  update acount's all channel feerate,保持内存与数据库中收费信息的一致
func (t *TokenNetwork) UpdateAccountFee(peerAddress common.Address, req *model.SetAllFeeRateRequest)   {
	getFee:= func(channelID common.Hash,token common.Address) *model.Fee {
		f,ok:=req.ChannelsFee[channelID]
		if ok{
			return f.Fee
		}
		f,ok=req.TokensFee[token]
		if ok{
			return f.Fee
		}
		return req.AccountFee.Fee
	}
	t.viewlock.Lock()
	defer t.viewlock.Unlock()
	for cid,c:=range t.channels{
		if c.Participant1==peerAddress{
			c.Participant1Fee=getFee(cid,c.Token)
			continue
		}
		if c.Participant2==peerAddress{
			c.Participant2Fee=getFee(cid,c.Token)
			continue
		}
	}
	return
}
//Stop stop TokenNetwork service
func (t *TokenNetwork) Stop() {
	t.transport.Stop()
}
