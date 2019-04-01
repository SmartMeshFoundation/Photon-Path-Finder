package blockchainlistener

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"sync"

	"fmt"

	"strings"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener/xmpppass"
	pparams "github.com/SmartMeshFoundation/Photon-Path-Finder/params"
	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/network/netshare"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mattn/go-xmpp"
)

var (
	errTimeout            = errors.New("timed out")
	errInvalidMessage     = errors.New("invalid message")
	errDuplicateWaiter    = errors.New("waiter with uid already exists")
	errWaiterClosed       = errors.New("waiter closed")
	errClientDisconnected = errors.New("client disconnected")
)

const (
	/*
			after offer send sdp,answer do the following jobs:
			1. receive sdp
			2. createicesteamtransport, contact with stun server,tunserver
			3. get it's own sdp
			4. send it's own sdp back to offer
		so how long should be better?
	*/
	defaultTimeout   = 15 * time.Second
	defaultReconnect = true
	nameSuffix       = "@mobileraiden"
	//TypeMobile photon run on a mobile device
	TypeMobile = "mobile"
	//TypeMeshBox photon run on a meshbox
	TypeMeshBox = "meshbox"
	//TypeOtherDevice photon run on a other device
	TypeOtherDevice = "other"
)

type testPasswordGeter struct {
	key *ecdsa.PrivateKey
}

func (t *testPasswordGeter) GetPassWord() string {
	pass, err := xmpppass.CreatePassword(t.key)
	if err != nil {
		panic(err)
	}
	return pass
}

// Config contains various client options.
type Config struct {
	Timeout time.Duration
}

// DefaultConfig with standard private channel prefix and 1 second timeout.
var DefaultConfig = &Config{
	Timeout: defaultTimeout,
}

/*
PasswordGetter generate login password
*/
type PasswordGetter interface {
	//get current password
	GetPassWord() string
}

//NodeStatus is status of a photon node
type NodeStatus struct {
	IsOnline   bool
	DeviceType string
}

// XMPPConnection describes client connection to xmpp server.
type XMPPConnection struct {
	lock           sync.RWMutex
	config         *Config
	options        xmpp.Options
	client         *xmpp.Client
	waitersMutex   sync.RWMutex
	waiters        map[string]chan interface{} //message waiting for response
	closed         chan struct{}
	reconnect      bool
	status         netshare.Status
	NextPasswordFn PasswordGetter
	name           string
	db             XMPPDb
	addrMap        map[common.Address]int //addr neighbor count
	listener       NodePresenceListener
	log            log.Logger
}

/*
NewXMPPConnection create Xmpp connection to signal sever
*/
func NewXMPPConnection(ServerURL string, db XMPPDb, listener NodePresenceListener) (x2 *XMPPConnection, err error) {
	key := pparams.ObserverKey()
	User := crypto.PubkeyToAddress(key.PublicKey)
	name := utils.APex2(User)
	deviceType := "other"
	passwordFn := &testPasswordGeter{key}
	x := &XMPPConnection{
		lock:   sync.RWMutex{},
		config: DefaultConfig,
		options: xmpp.Options{
			Host:                         ServerURL,
			User:                         fmt.Sprintf("%s%s", strings.ToLower(User.String()), nameSuffix),
			Password:                     passwordFn.GetPassWord(),
			NoTLS:                        true,
			InsecureAllowUnencryptedAuth: true,
			Debug:                        false,
			Session:                      false,
			Status:                       "xa",
			StatusMessage:                name,
			Resource:                     deviceType,
		},
		client:         nil,
		waitersMutex:   sync.RWMutex{},
		waiters:        make(map[string]chan interface{}),
		closed:         make(chan struct{}),
		addrMap:        make(map[common.Address]int),
		reconnect:      true,
		status:         netshare.Disconnected,
		NextPasswordFn: passwordFn,
		name:           name,
		listener:       listener,
		db:             db,
		log:            log.New("xmpp", name),
	}
	x.log.Trace(fmt.Sprintf("%s new xmpp user %s password %s", name, User.String(), x.options.Password))
	x.client, err = x.options.NewClient()
	if err != nil {
		err = fmt.Errorf("%s new xmpp client err %s", name, err)
		return
	}
	x.changeStatus(netshare.Connected)
	go x.loop()
	x2 = x
	return
}
func (x *XMPPConnection) loop() {
	for {
		chat, err := x.client.Recv()
		if x.status == netshare.Closed {
			return
		}
		if err != nil {
			//todo how to detect network error ,disconnect
			x.log.Error(fmt.Sprintf("%s receive error %s ,try to reconnect ", x.name, err))
			err = x.client.Close()
			if err != nil {
				x.log.Error(fmt.Sprintf("xmpp close err %s", err))
			}
			x.reConnect()
			continue
		}
		switch v := chat.(type) {

		case xmpp.Presence:
			if len(v.ID) > 0 {
				//subscribe or unsubscribe
				uid := v.ID
				x.waitersMutex.Lock()
				ch, ok := x.waiters[uid]
				x.waitersMutex.Unlock()
				if ok {
					x.log.Trace(fmt.Sprintf("%s %s received response", x.name, uid))
					ch <- &v
				} else {
					x.log.Info(fmt.Sprintf("receive unkonwn iq message %s", utils.StringInterface(v, 3)))
				}
			} else {
				var id, device string
				ss := strings.Split(v.From, "/")
				if len(ss) >= 2 {
					device = ss[1]
				}
				id = ss[0]
				bs := &NodeStatus{
					DeviceType: device,
					IsOnline:   len(v.Type) == 0,
				}
				if bs.IsOnline && len(bs.DeviceType) == 0 {
					x.log.Error(fmt.Sprintf("receive unexpected presence %s", utils.StringInterface(v, 3)))
				}
				ids := strings.Split(id, "@")
				addr := common.HexToAddress(ids[0])
				if bs.IsOnline {

					x.listener.Online(addr, bs.DeviceType)
				} else {
					x.listener.Offline(addr)
				}
				//x.nodesStatus[id] = bs
				x.log.Trace(fmt.Sprintf("node status change %s, deviceType=%s,isonline=%v", id, bs.DeviceType, bs.IsOnline))
			}
		default:
			//x.log.Trace(fmt.Sprintf("recv %s", utils.StringInterface(v, 3)))
		}
	}
}
func (x *XMPPConnection) changeStatus(newStatus netshare.Status) {
	x.log.Info(fmt.Sprintf("changeStatus from %d to %d", x.status, newStatus))
	x.status = newStatus
}

//Reconnect :
func (x *XMPPConnection) Reconnect() {
	err := x.client.Close()
	if err != nil {
		x.log.Warn(fmt.Sprintf("xmpp client close err %s", err))
	}
	return
}

func (x *XMPPConnection) reConnect() {
	x.changeStatus(netshare.Reconnecting)
	o := x.options
	for {
		if x.status == netshare.Closed {
			return
		}
		o.Password = x.NextPasswordFn.GetPassWord()
		client, err := o.NewClient()
		if err != nil {
			x.log.Error(fmt.Sprintf("%s xmpp reconnect error %s", x.name, err))
			time.Sleep(time.Second)
			continue
		}
		x.client = client
		break
	}
	x.changeStatus(netshare.Connected)
}

func (x *XMPPConnection) send(msg *xmpp.Chat) error {
	select {
	case <-x.closed:
		return errClientDisconnected
	default:
		cli := x.client
		x.log.Trace(fmt.Sprintf("%s send msg %s:%s %s", x.name, msg.Remote, msg.Subject, msg.Text))
		_, err := cli.Send(*msg)
		if err != nil {
			return err
		}
	}
	return nil
}

//Stop close this connection
func (x *XMPPConnection) Stop() {
	x.changeStatus(netshare.Closed)
	close(x.closed)
	err := x.client.Close()
	if err != nil {
		x.log.Error(fmt.Sprintf("close err %s", err))
	}
}

//Connected returns true when this connection is ready for sent
func (x *XMPPConnection) Connected() bool {
	return x.status == netshare.Connected
}

func (x *XMPPConnection) sendPresence(msg *xmpp.Presence) error {
	select {
	case <-x.closed:
		return errClientDisconnected
	default:
		cli := x.client
		x.log.Trace(fmt.Sprintf("%s send msg %s:%s %s", x.name, msg.From, msg.To, msg.ID))
		_, err := cli.SendPresence(*msg)
		if err != nil {
			return err
		}
	}
	return nil
}
func (x *XMPPConnection) sendSyncPresence(msg *xmpp.Presence) (response *xmpp.Presence, err error) {
	uid := msg.ID
	wait := make(chan interface{})
	err = x.addWaiter(uid, wait)
	if err != nil {
		return nil, err
	}
	defer x.removeWaiter(uid)
	err = x.sendPresence(msg)
	if err != nil {
		return nil, err
	}
	r, err := x.wait(wait)
	if err != nil {
		return
	}
	response, ok := r.(*xmpp.Presence)
	if !ok {
		x.log.Error(fmt.Sprintf("recevie response %s,but type error ", utils.StringInterface(r, 3)))
		err = errors.New("type error")
	}
	return
}
func (x *XMPPConnection) addWaiter(uid string, ch chan interface{}) error {
	x.waitersMutex.Lock()
	defer x.waitersMutex.Unlock()
	if _, ok := x.waiters[uid]; ok {
		return errDuplicateWaiter
	}
	x.waiters[uid] = ch
	return nil
}

func (x *XMPPConnection) removeWaiter(uid string) error {
	x.waitersMutex.Lock()
	defer x.waitersMutex.Unlock()
	delete(x.waiters, uid)
	return nil
}

func (x *XMPPConnection) wait(ch chan interface{}) (response interface{}, err error) {
	select {
	case data, ok := <-ch:
		if !ok {
			return nil, errWaiterClosed
		}
		return data, nil
	case <-time.After(x.config.Timeout):
		return nil, errTimeout
	case <-x.closed:
		return nil, errClientDisconnected
	}
}

//SubscribeNeighbour the status change of `addr`
func (x *XMPPConnection) SubscribeNeighbour(addr common.Address) error {
	x.lock.Lock()
	defer x.lock.Unlock()
	cnt := x.addrMap[addr]
	x.addrMap[addr] = cnt + 1
	//数据库中记录已经查询过了
	if x.db.XMPPIsAddrSubed(addr) {
		return nil
	}
	addrName := fmt.Sprintf("%s%s", strings.ToLower(addr.String()), nameSuffix)
	p := xmpp.Presence{
		From: x.options.User,
		To:   addrName,
		Type: "subscribe",
		ID:   utils.RandomString(10),
	}
	err := x.sendPresence(&p)
	if err == nil {
		x.db.XMPPMarkAddrSubed(addr)
	}
	return err
}

//Unsubscribe the status change of `addr`
/*
```xml
<presence id='xk3h1v69' to='leon@mobilephoton' type='unsubscribe'/>
```
*/
func (x *XMPPConnection) Unsubscribe(addr common.Address) error {
	x.lock.Lock()
	defer x.lock.Unlock()
	cnt := x.addrMap[addr]
	cnt--
	if cnt <= 0 {
		x.addrMap[addr] = 0
		addrName := fmt.Sprintf("%s%s", strings.ToLower(addr.String()), nameSuffix)
		p := xmpp.Presence{
			From: x.options.User,
			To:   addrName,
			Type: "unsubscribe",
			ID:   utils.RandomString(10),
		}
		_, err := x.sendSyncPresence(&p)
		x.db.XMPPUnMarkAddr(addr)
		return err
	}
	x.addrMap[addr] = cnt
	return nil
}

//SubscribeNeighbors I want to know these `addrs` status change
func (x *XMPPConnection) SubscribeNeighbors(addrs []common.Address) error {
	for _, addr := range addrs {
		err := x.SubscribeNeighbour(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
XMPPDb 解耦 db 依赖
*/
type XMPPDb interface {
	XMPPIsAddrSubed(addr common.Address) bool
	XMPPMarkAddrSubed(addr common.Address)
	XMPPUnMarkAddr(addr common.Address)
}
