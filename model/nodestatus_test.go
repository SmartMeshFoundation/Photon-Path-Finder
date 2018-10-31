package model

import (
	"testing"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/stretchr/testify/assert"
)

func TestGetAllNodes(t *testing.T) {
	SetupTestDB()
	//key, _ := utils.MakePrivateKeyAddress()
	//t.Logf(hex.EncodeToString(crypto.FromECDSA(key)))
	nodes := GetAllNodes()
	assert.EqualValues(t, len(nodes), 0)
	NewOrUpdateNodeStatus(utils.NewRandomAddress(), true, "mobile")
	nodes = GetAllNodes()
	assert.EqualValues(t, len(nodes), 1)
	addr := utils.NewRandomAddress()
	NewOrUpdateNodeStatus(addr, true, "mobile")
	assert.EqualValues(t, len(GetAllNodes()), 2)
	NewOrUpdateNodeOnline(addr, false)
	nodes = GetAllNodes()
	assert.EqualValues(t, len(nodes), 2)
	t.Logf("nodes=%s", utils.StringInterface(nodes, 3))
}
