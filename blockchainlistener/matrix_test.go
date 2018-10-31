package blockchainlistener

import (
	"os"
	"testing"
	"time"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stderr)))
}

type mockListener struct {
	t *testing.T
}

func (m *mockListener) Online(address common.Address, deviceType string) {
	m.t.Logf("online %s %s ", address.String(), deviceType)
}
func (m *mockListener) Offline(address common.Address) {
	m.t.Logf("offline %s", address.String())
}
func TestNewMatrixObserver(t *testing.T) {
	m := NewMatrixObserver(&mockListener{t})
	time.Sleep(time.Second * 15)
	m.Stop()
}
