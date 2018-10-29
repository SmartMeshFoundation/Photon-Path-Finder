package clientapi

import (
	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/clientapi/routing"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common/basecomponent"
)

//SetupClientAPIComponent set up client api component,reserved for s-s api
func SetupClientAPIComponent(
	base *basecomponent.BasePathFinder,
	pfsdb *storage.Database, ce *blockchainlistener.ChainEvents,
) {
	routing.Setup(
		base.APIMux,
		*base.Cfg,
		pfsdb,
		*ce,
	)
}
