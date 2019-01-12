package blockchainlistener

import "github.com/ethereum/go-ethereum/common"

//Transporter 用于节点上线下线发现
type Transporter interface {
	Stop() //停止服务
	SubscribeNeighbors(addrs []common.Address) error
	Unsubscribe(addr common.Address) error
}
