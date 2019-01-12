package blockchainlistener

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/SmartMeshFoundation/Photon/network/gomatrix"

	"time"

	pparams "github.com/SmartMeshFoundation/Photon-Path-Finder/params"
	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/params"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ONLINE network state -online
	ONLINE = "online"
	// UNAVAILABLE state -unavailable
	UNAVAILABLE = "unavailable"
	// OFFLINE state -offline
	OFFLINE = "offline"
	// UNKNOWN or other state -unknown
	UNKNOWN = "unknown"
	// ROOMPREFIX room prefix
	ROOMPREFIX = "photon"
	// ROOMSEP with ',' to separate room name's part
	ROOMSEP = "_"
	// PATHPREFIX0 the lastest transport client api version
	PATHPREFIX0 = "/_matrix/client/r0"

	// LOGINTYPE login type we used
	LOGINTYPE = "m.login.password"
)

// MatrixObserver represents a transport transport Instantiation
type MatrixObserver struct {
	matrixcli      *gomatrix.MatrixClient //the instantiated transport
	key            *ecdsa.PrivateKey      //key
	nodeAddresses  common.Address
	serverURL      string
	serverName     string
	UserID         string //the current user's ID(@kitty:thisserver)
	nodeDeviceType string

	log      log.Logger
	running  bool
	listener NodePresenceListener
}

//NodePresenceListener for notification from transport
type NodePresenceListener interface {
	//an account is online
	Online(address common.Address, deviceType string)
	//an account is offline
	Offline(address common.Address)
}

var (
	// ValidUserIDRegex user ID 's format
	ValidUserIDRegex = regexp.MustCompile(`^@(0x[0-9a-f]{40})(?:\.[0-9a-f]{8})?(?::.+)?$`) //(`^[0-9a-z_\-./]+$`)
	//NETWORKNAME which network is used
	NETWORKNAME = params.NETWORKNAME
	//ALIASFRAGMENT the terminal part of alias
	ALIASFRAGMENT = params.AliasFragment
	//DISCOVERYROOMSERVER discovery room server name
	DISCOVERYROOMSERVER = params.DiscoveryServer
)

// NewMatrixObserver init transport
func NewMatrixObserver(listener NodePresenceListener) *MatrixObserver {
	key := pparams.ObserverKey()
	mtr := &MatrixObserver{
		nodeAddresses:  crypto.PubkeyToAddress(key.PublicKey),
		key:            key,
		nodeDeviceType: "other",
		log:            log.New("transport", "finder"),
		serverURL:      fmt.Sprintf("http://%s:8008", pparams.MatrixServer),
		serverName:     pparams.MatrixServer,
		running:        true,
		listener:       listener,
	}
	go mtr.start()
	return mtr
}

// Stop Does Stop need to destroy transport resource ?
func (m *MatrixObserver) Stop() {
	m.running = false
	if m.matrixcli != nil {
		err := m.matrixcli.SetPresenceState(&gomatrix.ReqPresenceUser{
			Presence: OFFLINE,
		})
		if err != nil {
			m.log.Error(fmt.Sprintf("[Matrix] SetPresenceState failed : %s", err.Error()))
		}
		m.matrixcli.StopSync()
		if _, err := m.matrixcli.Logout(); err != nil {
			m.log.Error("[Matrix] Logout failed")
		}
	}
}

// Start transport
func (m *MatrixObserver) start() {

	for {
		var err error
		var syncer *gomatrix.DefaultSyncer
		if !m.running {
			return
		}
		m.matrixcli, err = gomatrix.NewClient(fmt.Sprintf("http://%s:8008", pparams.MatrixServer), "", "", PATHPREFIX0, m.log)
		if err != nil {
			log.Error(fmt.Sprintf("transport connection error %s", err))
			time.Sleep(time.Second * 5)
			continue
		}
		_, err = m.matrixcli.Versions()
		if err != nil {
			m.log.Error(fmt.Sprintf("Could not connect to requested server %s,and retrying,err %s", pparams.MatrixServer, err))
			continue
		}

		// log in
		if err = m.loginOrRegister(); err != nil {
			m.log.Error(fmt.Sprintf("loginOrRegister err %s", err))
			time.Sleep(time.Second * 5)
			continue
		}
		//initialize Filters/NextBatch/Rooms
		m.matrixcli.Store = gomatrix.NewInMemoryStore()

		//handle the issue of discoveryroom,FOR TEST,temporarily retain this room
		if err = m.joinDiscoveryRoom(); err != nil {
			m.log.Error(fmt.Sprintf("joinDiscoveryRoom err %s", err))
			goto tryNext
		}
		//notify to server i am online（include the other participating servers）
		if err = m.matrixcli.SetPresenceState(&gomatrix.ReqPresenceUser{
			Presence:  ONLINE,
			StatusMsg: m.nodeDeviceType, //register device type to server
		}); err != nil {
			m.log.Error(fmt.Sprintf("SetPresenceState err %s", err))
			goto tryNext
		}
		//register receive-datahandle or other message received

		m.matrixcli.Syncer = gomatrix.NewDefaultSyncer(m.UserID, m.matrixcli.Store)
		syncer = m.matrixcli.Syncer.(*gomatrix.DefaultSyncer)

		syncer.OnEventType("m.presence", m.onHandlePresenceChange)

		for {
			err2 := m.matrixcli.Sync()

			if !m.running {
				return
			}
			if err2 != nil {
				m.log.Error(fmt.Sprintf("Matrix Sync return,err=%s ,will try agin..", err))
				time.Sleep(time.Second * 5)
			}
		}
	tryNext:
		time.Sleep(time.Second * 5)
	}
}

/*
onHandlePresenceChange handle events in this message, about changes of nodes and update AddressToPresence

{
	"content": {
		"status_msg": "other",
		"currently_active": true,
		"last_active_ago": 13,
		"presence": "online"
	},
	"type": "m.presence",
	"sender": "@0xf156aba37a64767769a96a0083f02f540e7856ab:transport01.smartmesh.cn"
}
*/
func (m *MatrixObserver) onHandlePresenceChange(event *gomatrix.Event) {
	if !m.running {
		return
	}
	if event.Type != "m.presence" {
		m.log.Error(fmt.Sprintf("onHandlePresenceChange receive unkonw event %s", utils.StringInterface(event, 5)))
		return
	}
	// parse address of message sender
	presence, exists := event.ViewContent("presence") //newest network status
	if !exists {
		return
	}
	deviceType, _ := event.ViewContent("status_msg") //newest network status
	address := m.userIDToAddress(event.Sender)

	if presence == ONLINE {
		m.listener.Online(address, deviceType)

	} else {
		m.listener.Offline(address)
	}
	m.log.Trace(fmt.Sprintf("peer %s status=%s,deviceType=%s", utils.APex2(address), presence, deviceType))
}

//register new user on homeserver using application service
func (m *MatrixObserver) register(username, password string) (userID string, err error) {
	type reg struct {
		LocalPart   string `json:"localpart"`   //@someone:transport.org someone is localpoart,transport.org is domain
		DisplayName string `json:"displayname"` // displayname of this user
		Password    string `json:"password,omitempty"`
	}
	type regResp struct {
		AccessToken string `json:"access_token"`
		HomeServer  string `json:"home_server"`
		UserID      string `json:"user_id"`
	}
	regurl := fmt.Sprintf("%s/regapp/1/register", m.serverURL)
	userID = fmt.Sprintf("@%s:%s", username, m.serverName)
	log.Trace(fmt.Sprintf("register user userid=%s", userID))
	req := &reg{
		LocalPart:   username,
		Password:    password,
		DisplayName: m.getUserDisplayName(userID),
	}
	resp := &regResp{}
	_, err = m.matrixcli.MakeRequest(http.MethodPost, regurl, req, resp)
	if err != nil {
		return
	}
	if resp.UserID != userID {
		err = fmt.Errorf("expect userid=%s,got=%s", userID, resp.UserID)
	}
	return
}

// loginOrRegister node login, if failed, register again then try login,
// displayname of nodes as the signature of userID
func (m *MatrixObserver) loginOrRegister() (err error) {
	//TODO:Consider the risk of being registered maliciously
	loginok := false
	baseAddress := crypto.PubkeyToAddress(m.key.PublicKey)
	baseUsername := strings.ToLower(baseAddress.String())

	username := baseUsername
	password := hex.EncodeToString(m.dataSign([]byte(m.serverName)))
	//password := "12345678"
	for i := 0; i < 5; i++ {
		var resplogin *gomatrix.RespLogin
		m.matrixcli.AccessToken = ""
		resplogin, err = m.matrixcli.Login(&gomatrix.ReqLogin{
			Type:     LOGINTYPE,
			User:     username,
			Password: password,
			DeviceID: "",
		})
		if err != nil {
			httpErr, ok := err.(gomatrix.HTTPError)
			if !ok { // network error,try again
				continue
			}
			if httpErr.Code == 403 { //Invalid username or password
				if i > 0 {
					m.log.Trace(fmt.Sprintf("couldn't sign in for transport,trying register %s", username))
				}
				userID, rerr := m.register(username, password)
				if rerr != nil {
					return rerr
				}
				m.matrixcli.UserID = userID
				continue
			}
		} else {
			//cache the node's and report the UserID and AccessToken to transport
			m.matrixcli.SetCredentials(resplogin.UserID, resplogin.AccessToken)
			m.UserID = resplogin.UserID
			m.nodeAddresses = baseAddress
			loginok = true
			break
		}
	}
	if !loginok {
		err = fmt.Errorf("could not register or login")
		return
	}
	//set displayname as publicly visible
	dispname := m.getUserDisplayName(m.matrixcli.UserID)
	if err = m.matrixcli.SetDisplayName(dispname); err != nil {
		err = fmt.Errorf("could set the node's displayname and quit as well")
		m.matrixcli.ClearCredentials()
		return
	}
	m.log.Trace(fmt.Sprintf("userdisplayname=%s", dispname))
	return err
}

// makeRoomAlias name room's alias
func (m *MatrixObserver) makeRoomAlias(thepart string) string {
	return ROOMPREFIX + ROOMSEP + NETWORKNAME + ROOMSEP + thepart
}

func (m *MatrixObserver) getUserDisplayName(userID string) string {
	sig := m.dataSign([]byte(userID))
	return fmt.Sprintf("%s-%s", utils.APex2(m.nodeAddresses), hex.EncodeToString(sig))
}

// dataSign 签名数据
// dataSign signature data
func (m *MatrixObserver) dataSign(data []byte) (signature []byte) {
	signature, err := utils.SignData(m.key, data)
	if err != nil {
		m.log.Error(fmt.Sprintf("SignData err %s", err))
		return nil
	}
	return
}

// joinDiscoveryRoom : check discoveryroom if not exist, then create a new one.
// client caches all memebers of this room, and invite nodes checked from this room again.
func (m *MatrixObserver) joinDiscoveryRoom() (err error) {
	//read discovery room'name and fragment from "params-settings"
	// combine discovery room's alias
	discoveryRoomAlias := m.makeRoomAlias(ALIASFRAGMENT)
	discoveryRoomAliasFull := "#" + discoveryRoomAlias + ":" + DISCOVERYROOMSERVER

	_, err = m.matrixcli.JoinRoom(discoveryRoomAliasFull, m.serverName, nil)
	if err != nil {
		m.log.Error(fmt.Sprintf("joinDiscoveryRoom %s ,err %s", discoveryRoomAliasFull, err))
		return err
	}

	return nil
}

func (m *MatrixObserver) userIDToAddress(userID string) common.Address {
	//check grammar of user ID
	_match := ValidUserIDRegex.MatchString(userID)
	if _match == false {
		m.log.Warn(fmt.Sprintf("UserID %s, format error", userID))
		return utils.EmptyAddress
	}
	addressHex, err := extractUserLocalpart(userID) //"@myname:photon.org:cy"->"myname"
	if err != nil {
		m.log.Error(fmt.Sprintf("extractUserLocalpart err %s", err))
		return utils.EmptyAddress
	}
	var addrmuti = regexp.MustCompile(`^(0x[0-9a-f]{40})`)
	addrlocal := addrmuti.FindString(addressHex)
	if addrlocal == "" {
		err = fmt.Errorf("%s not match our userid rule", userID)
		return utils.EmptyAddress
	}
	address := common.HexToAddress(addrlocal)
	return address
}

// ExtractUserLocalpart Extract user name from user ID
func extractUserLocalpart(userID string) (string, error) {
	if len(userID) == 0 || userID[0] != '@' {
		return "", fmt.Errorf("%s is not a valid user id", userID)
	}
	return strings.SplitN(userID[1:], ":", 2)[0], nil
}

//SubscribeNeighbors for Transporter interface only
func (m *MatrixObserver) SubscribeNeighbors(addrs []common.Address) error {
	return nil
}

//Unsubscribe for Transporter interface only
func (m *MatrixObserver) Unsubscribe(addr common.Address) error {
	return nil
}
