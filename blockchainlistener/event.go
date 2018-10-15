package blockchainlistener

import (
	"crypto/ecdsa"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer"
	"fmt"
	"sync/atomic"
	//"github.com/SmartMeshFoundation/SmartRaiden-Monitoring/models"
	"github.com/SmartMeshFoundation/SmartRaiden/blockchain"
	"github.com/SmartMeshFoundation/SmartRaiden/log"
	"github.com/SmartMeshFoundation/SmartRaiden/network/helper"
	"github.com/SmartMeshFoundation/SmartRaiden/network/rpc"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer/mediatedtransfer"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/model"
)

//ChainEvents block chain operations
type ChainEvents struct {
	client          *helper.SafeEthClient
	be              *blockchain.Events
	bcs             *rpc.BlockChainService
	key             *ecdsa.PrivateKey
	//db              *models.ModelDB
	quitChan        chan struct{}
	alarm           *blockchain.AlarmTask
	stopped         bool
	BlockNumberChan chan int64
	blockNumber     *atomic.Value
	tokennetork     model.TokenNetwork
}

//NewChainEvents create chain events
func NewChainEvents(key *ecdsa.PrivateKey, client *helper.SafeEthClient, tokenNetworkRegistryAddress common.Address) *ChainEvents {//, db *models.ModelDB
	log.Trace(fmt.Sprintf("tokenNetworkRegistryAddress %s", tokenNetworkRegistryAddress.String()))
	bcs := rpc.NewBlockChainService(key, tokenNetworkRegistryAddress, client)
	registry := bcs.Registry(tokenNetworkRegistryAddress)
	if registry == nil {
		panic("startup error : cannot get registry")
	}
	secretRegistryAddress, err := registry.GetContract().SecretRegistryAddress(nil)
	if err != nil {
		panic(err)
	}
	return &ChainEvents{
		client:          client,
		be:              blockchain.NewBlockChainEvents(client, tokenNetworkRegistryAddress, secretRegistryAddress, nil),
		bcs:             bcs,
		key:             key,
		//db:              db,
		quitChan:        make(chan struct{}),
		alarm:           blockchain.NewAlarmTask(client),
		BlockNumberChan: make(chan int64, 1),
		blockNumber:     new(atomic.Value),
	}
}

//Start moniter blockchain
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
		log.Error(fmt.Sprintf("blockchain events start err %s", err))
	}
	go chainevent.loop()
	return nil
}

//Stop service
func (chainevent *ChainEvents) Stop() {
	chainevent.alarm.Stop()
	chainevent.be.Stop()
	close(chainevent.quitChan)
}
func (chainevent *ChainEvents) setBlockNumber(blocknumber int64) {
	if chainevent.stopped {
		log.Info(fmt.Sprintf("new block number arrived %d,but has stopped", blocknumber))
		return
	}
	chainevent.BlockNumberChan <- blocknumber
}

func (chainevent *ChainEvents) loop() {
	for {
		select {
		case st, ok := <-chainevent.be.StateChangeChannel:
			if !ok {
				log.Info("StateChangeChannel closed")
				return
			}
			chainevent.handleStateChange(st)
		case n,ok:=<-chainevent.BlockNumberChan:
			if !ok{
				log.Info("BlockNumberChan closed")
				return
			}
			chainevent.handleBlockNumber(n)
		case <-chainevent.quitChan:
			return
		}
	}
}
//handleStateChange 通道打开、通道关闭、通道存钱、通道取钱
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
	}
}

func (chainevent *ChainEvents) handleBlockNumber(n int64) {

}
// handleNewChannelStateChange Open channel
func (chainevent *ChainEvents)handleChainChannelOpend(st2 *mediatedtransfer.ContractNewChannelStateChange)  {
	tokenNetwork:=st2.TokenNetworkAddress

	if tokenNetwork==utils.EmptyAddress{
		return
	}
	if !checkValidity(){
		return
	}

	log.Debug(fmt.Sprintf("Received ChannelOpened event for token network %s", tokenNetwork.String()))

	channelID:=st2.ChannelIdentifier.ChannelIdentifier
	participant1:=st2.Participant1
	participant2:=st2.Participant2

	chainevent.tokennetork.HandleChannelOpenedEvent(channelID,participant1,participant2)
}

// handleDepositStateChange deposit
func (chainevent *ChainEvents) handleChainChannelDeposit(st2 *mediatedtransfer.ContractBalanceStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress
	if tokenNetwork==utils.EmptyAddress{
		return
	}
	if !checkValidity(){
		return
	}
	log.Debug(fmt.Sprintf("Received ChannelDeposit event for token network %s", tokenNetwork.String()))
	channelID:=st2.ChannelIdentifier
	participantAddress:=st2.ParticipantAddress
	totalDeposit:=st2.Balance
	chainevent.tokennetork.HandleChannelDeposit(channelID,participantAddress,totalDeposit)
}

// handleChainChannelClosed Close Channel
func (chainevent *ChainEvents) handleChainChannelClosed(st2 *mediatedtransfer.ContractClosedStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress

	if tokenNetwork==utils.EmptyAddress{
		return
	}
	if !checkValidity(){
		return
	}
	log.Debug(fmt.Sprintf("Received ChannelClosed event for token network %s", tokenNetwork.String()))

	channelID:=st2.ChannelIdentifier

	chainevent.tokennetork.HandleChannelClosedEvent(channelID)
}

// handleWithdrawStateChange Withdraw
func (chainevent *ChainEvents) handleWithdrawStateChange(st2 *mediatedtransfer.ContractChannelWithdrawStateChange) {
	tokenNetwork:=st2.TokenNetworkAddress
	if tokenNetwork==utils.EmptyAddress{
		return
	}
	if !checkValidity(){
		return
	}
	log.Debug(fmt.Sprintf("Received ChannelWithdraw event for token network %s", tokenNetwork.String()))

	channelID:=st2.ChannelIdentifier.ChannelIdentifier
	participant1:=st2.Participant1
	participant2:=st2.Participant2
	participant1Balance:=st2.Participant1Balance
	participant2Balance:=st2.Participant2Balance
	chainevent.tokennetork.HandleChannelWithdawEvent(channelID,participant1,participant2,participant1Balance,participant2Balance)
}

// SaveLatestBlockNumber
func (chainevent *ChainEvents)SaveLatestBlockNumber(blockNumber int64){

}

// GetLatestBlockNumber
func (chainevent *ChainEvents)GetLatestBlockNumber() int64 {
	return 6524830
}

func checkValidity()bool  {
	//...
	return true
}




