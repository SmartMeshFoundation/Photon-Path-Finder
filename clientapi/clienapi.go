package clientapi

import (
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/basecomponent"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/routing"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
)

//SetupClientAPIComponent set up client api component,reserved for s-s api
func SetupClientAPIComponent(
	base *basecomponent.BasePathFinder,
	pfsdb *storage.Database,ce *blockchainlistener.ChainEvents,
) {
	routing.Setup(
		base.APIMux,
		*base.Cfg,
		pfsdb,
		*ce,
	)
}
