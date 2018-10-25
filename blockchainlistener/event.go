package blockchainlistener

import (
	"crypto/ecdsa"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer"
	"fmt"
	"sync/atomic"
	"github.com/SmartMeshFoundation/SmartRaiden/blockchain"
	"github.com/SmartMeshFoundation/SmartRaiden/network/helper"
	"github.com/SmartMeshFoundation/SmartRaiden/network/rpc"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer/mediatedtransfer"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/sirupsen/logrus"
)

//ChainEvents block chain operations
type ChainEvents struct {
	client          *helper.SafeEthClient
	be              *blockchain.Events
	bcs             *rpc.BlockChainService
	key             *ecdsa.PrivateKey
	quitChan        chan struct{}
	alarm           *blockchain.AlarmTask
	stopped         bool
	BlockNumberChan chan int64
	blockNumber     *atomic.Value
	TokenNetwork     model.TokenNetwork
	db              *storage.Database
}

// NewChainEvents create chain events
func NewChainEvents(key *ecdsa.PrivateKey, client *helper.SafeEthClient, tokenNetworkRegistryAddress common.Address,db *storage.Database) *ChainEvents { //, db *models.ModelDB
	logrus.Info("Token Network registry address=",tokenNetworkRegistryAddress.String())
	bcs, err := rpc.NewBlockChainService(key, tokenNetworkRegistryAddress, client)
	if err != nil {
		logrus.Panic(err)
	}
	registry := bcs.Registry(tokenNetworkRegistryAddress, true)
	if registry == nil {
		logrus.Panic("Register token network error : cannot get registry")
	}

	token2TokenNetwork,err:=db.GetAllTokensStorage(nil)
	//fmt.Printf("token2TokenNetwork startup=%s",utils.StringInterface(token2TokenNetwork,2))
	logrus.Infoln("token2TokenNetwork startup=%s",utils.StringInterface(token2TokenNetwork,2))
	//logrus.in
	return &ChainEvents{
		client:          client,
		be:              blockchain.NewBlockChainEvents(client, bcs, token2TokenNetwork),
		bcs:             bcs,
		key:             key,
		db:              db,
		quitChan:        make(chan struct{}),
		alarm:           blockchain.NewAlarmTask(client),
		BlockNumberChan: make(chan int64, 1),
		blockNumber:     new(atomic.Value),
		TokenNetwork:     *model.InitTokenNetwork(tokenNetworkRegistryAddress,db),
	}
}

// Start moniter blockchain
func (chainevent *ChainEvents) Start() error {
	err := chainevent.alarm.Start()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case blockNumber := <-chainevent.alarm.LastBlockNumberChan:
				chainevent.SaveLatestBlockNumber(blockNumber)
				chainevent.setBlockNumber(blockNumber)
			}
		}
	}()
	err = chainevent.be.Start(chainevent.GetLatestBlockNumber())
	if err != nil {
		logrus.Errorf("Block chain events start err %s",err)
	}
	go chainevent.loop()
	return nil
}

// Stop service
func (chainevent *ChainEvents) Stop() {
	chainevent.alarm.Stop()
	chainevent.be.Stop()
	close(chainevent.quitChan)
}

//setBlockNumber set block number
func (chainevent *ChainEvents) setBlockNumber(blocknumber int64) {
	if chainevent.stopped {
		logrus.Infof("New block number arrived %d,but has stopped", blocknumber)
		return
	}
	chainevent.BlockNumberChan <- blocknumber
}

// loop loop
func (chainevent *ChainEvents) loop() {
	for {
		select {
		case st, ok := <-chainevent.be.StateChangeChannel:
			if !ok {
				logrus.Info("StateChangeChannel closed")
				return
			}
			chainevent.handleStateChange(st)
		case n,ok:=<-chainevent.BlockNumberChan:
			if !ok{
				logrus.Info("BlockNumberChan closed")
				return
			}
			chainevent.handleBlockNumber(n)
		case <-chainevent.quitChan:
			return
		}
	}
}
// handleStateChange 通道打开、通道关闭、通道存钱、通道取钱
func (chainevent *ChainEvents) handleStateChange(st transfer.StateChange) {
	switch st2 := st.(type) {
	case *mediatedtransfer.ContractNewChannelStateChange: //open channel event
		chainevent.handleChainChannelOpend(st2)
	case *mediatedtransfer.ContractClosedStateChange: //close channel event
		chainevent.handleChainChannelClosed(st2)
	case *mediatedtransfer.ContractBalanceStateChange: //deposit event
		chainevent.handleChainChannelDeposit(st2)
	case *mediatedtransfer.ContractChannelWithdrawStateChange: //withdaw event
		chainevent.handleWithdrawStateChange(st2)
	case *mediatedtransfer.ContractTokenAddedStateChange:
		//chainevent.be.TokenNetworks[st2.TokenNetworkAddress] = true
		chainevent.handleTokenAddedStateChange(st2)
	}
}

// handleTokenAddedStateChange Token added
func (chainevent *ChainEvents) handleTokenAddedStateChange(st2 *mediatedtransfer.ContractTokenAddedStateChange) {
	logrus.Info("Received TokenAddedStateChange event for token ", st2.TokenAddress.String())
	fmt.Println(fmt.Sprintf("Received TokenAddedStateChange data: %s", utils.StringInterface(st2, 2)))
	err := chainevent.db.SaveTokensStorage(nil, st2.TokenAddress.String(), st2.TokenNetworkAddress.String())
	if err != nil {
		logrus.Error("Handle token added state change event error,err=", err)
	}
}

func (chainevent *ChainEvents) handleBlockNumber(n int64) {
}

// existTokenNetwork
func (chainevent *ChainEvents)existTokenNetwork(channelID common.Hash) bool{
	if _,exist:=chainevent.TokenNetwork.ChannelID2Address[channelID];!exist{
		return false
	}
	return true
}

// handleNewChannelStateChange Open channel
func (chainevent *ChainEvents)handleChainChannelOpend(st2 *mediatedtransfer.ContractNewChannelStateChange)  {
	tokenNetwork:=st2.TokenNetworkAddress

	logrus.Info("Received ChannelOpened event for token network ",tokenNetwork.String())

	channelID:=st2.ChannelIdentifier.ChannelIdentifier
	participant1:=st2.Participant1
	participant2:=st2.Participant2
	fmt.Println(fmt.Sprintf("Received ChannelOpened data: %s",utils.StringInterface(st2,2)))
	err:=chainevent.TokenNetwork.HandleChannelOpenedEvent(tokenNetwork,channelID,participant1,participant2)
	if err!=nil{
		logrus.Warn("Handle channel open event error,err=",err)
	}

}

// handleDepositStateChange deposit
func (chainevent *ChainEvents) handleChainChannelDeposit(st2 *mediatedtransfer.ContractBalanceStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress

	logrus.Info("Received ChannelDeposit event for token network ",tokenNetwork.String())

	channelID:=st2.ChannelIdentifier
	participantAddress:=st2.ParticipantAddress
	totalDeposit:=st2.Balance
	fmt.Println(fmt.Sprintf("Received ChannelDeposit data: %s",utils.StringInterface(st2,2)))
	err:=chainevent.TokenNetwork.HandleChannelDepositEvent(tokenNetwork,channelID,participantAddress,totalDeposit)
	if err!=nil{
		logrus.Warn("Handle channel deposit event error,err=",err)
	}
}

// handleChainChannelClosed Close Channel
func (chainevent *ChainEvents) handleChainChannelClosed(st2 *mediatedtransfer.ContractClosedStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress

	logrus.Info("Received ChannelClosed event for token network ",tokenNetwork.String())

	channelID:=st2.ChannelIdentifier
	fmt.Println(fmt.Sprintf("Received ChannelClosed data: %s",utils.StringInterface(st2,2)))
	err:=chainevent.TokenNetwork.HandleChannelClosedEvent(tokenNetwork,channelID)
	if err!=nil{
		logrus.Warn("Handle channel close event error,err=",err)
	}
}

// handleWithdrawStateChange Withdraw
func (chainevent *ChainEvents) handleWithdrawStateChange(st2 *mediatedtransfer.ContractChannelWithdrawStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress

	logrus.Info("Received ChannelWithdraw event for token network ",tokenNetwork.String())

	channelID:=st2.ChannelIdentifier.ChannelIdentifier
	participant1:=st2.Participant1
	participant2:=st2.Participant2
	participant1Balance:=st2.Participant1Balance
	participant2Balance:=st2.Participant2Balance
	fmt.Println(fmt.Sprintf("Received ChannelWithdraw data: %s",utils.StringInterface(st2,2)))
	err:=chainevent.TokenNetwork.HandleChannelWithdrawEvent(tokenNetwork,channelID,participant1,participant2,participant1Balance,participant2Balance)
	if err!=nil{
		logrus.Warn("Handle channel withdaw event error,err=",err)
	}
}

// SaveLatestBlockNumber
func (chainevent *ChainEvents)SaveLatestBlockNumber(blockNumber int64){
	err:=chainevent.db.SaveLatestBlockNumberStorage(nil,blockNumber)
	if err!=nil{
		logrus.Error("Models (SaveLatestBlockNumber) err=",err)
	}
}

// GetLatestBlockNumber
func (chainevent *ChainEvents)GetLatestBlockNumber() int64 {
	number,err:=chainevent.db.GetLatestBlockNumberStorage(nil)
	if err != nil {
		logrus.Error("Models (GetLatestBlockNumber) err=",err)
	}
	fmt.Println(number)
	//return number
	return 0//just test
}





