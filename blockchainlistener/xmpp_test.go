package blockchainlistener

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"time"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"
	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stderr)))
}

type testdb struct {
	m map[common.Address]bool
}

func (t *testdb) XMPPIsAddrSubed(addr common.Address) bool {
	return t.m[addr]
}
func (t *testdb) XMPPMarkAddrSubed(addr common.Address) {
	t.m[addr] = true
}
func (t *testdb) XMPPUnMarkAddr(addr common.Address) {
	t.m[addr] = false
}

type testlistener struct {
	m    map[common.Address]*nodeStatus
	name string
}

func (t *testlistener) Online(address common.Address, deviceType string) {
	log.Trace(fmt.Sprintf("%s online,deviceType=%s observer=%s", address.String(), deviceType, t.name))
	t.m[address] = &nodeStatus{
		isOnline: true,
		isMobile: deviceType == "mobile",
	}
}

//an account is offline
func (t *testlistener) Offline(address common.Address) {
	log.Trace(fmt.Sprintf("%s offline,observer=%s", address.String(), t.name))
	delete(t.m, address)
}
func TestSubscribe(t *testing.T) {
	if testing.Short() {
		return
	}

	key1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)
	key2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)
	log.Trace(fmt.Sprintf("addr1=%s,addr2=%s\n", addr1.String(), addr2.String()))

	params.ObserverKey = func() *ecdsa.PrivateKey { return key1 }
	t1listener := &testlistener{
		m:    make(map[common.Address]*nodeStatus),
		name: "a1",
	}
	x1, err := NewXMPPConnection(params.DefaultXMPPServer, &testdb{
		m: make(map[common.Address]bool),
	}, t1listener)

	if err != nil {
		t.Error(err)
		return
	}
	err = x1.SubscribeNeighbour(addr2)
	if err != nil {
		t.Error(err)
		return
	}
	log.Trace(fmt.Sprintf("subscribe %s", addr2.String()))

	defer x1.Stop()
	if t1listener.m[addr2] != nil {
		t.Error("should not online")
		return
	}
	log.Trace("client2 will login")
	params.ObserverKey = func() *ecdsa.PrivateKey { return key2 }
	x2, err := NewXMPPConnection(params.DefaultXMPPServer, &testdb{
		m: make(map[common.Address]bool),
	}, &testlistener{
		m:    make(map[common.Address]*nodeStatus),
		name: "a2",
	})
	if err != nil {
		t.Error(err)
		return
	}
	//wait notification from server
	time.Sleep(time.Millisecond * 1000)
	if !t1listener.m[addr2].isOnline || t1listener.m[addr2].isMobile {
		t.Errorf("should  online,err=%v,isonline=%v,ismobile=%v", err, t1listener.m[addr2].isOnline, t1listener.m[addr2].isMobile)
		return
	}
	log.Trace("client2 will logout")
	x2.Stop()
	time.Sleep(time.Millisecond * 1000)

	if t1listener.m[addr2] != nil {
		t.Error("should  offline")
		return
	}

	err = x1.Unsubscribe(addr2)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Millisecond * 100)
	log.Trace("client2 will relogin")
	params.ObserverKey = func() *ecdsa.PrivateKey { return key2 }
	x2, err = NewXMPPConnection(params.DefaultXMPPServer, &testdb{
		m: make(map[common.Address]bool),
	}, &testlistener{
		m:    make(map[common.Address]*nodeStatus),
		name: "a2",
	})
	if err != nil {
		t.Error(err)
		return
	}
	log.Trace("client2 will logout")
	x2.Stop()
}
