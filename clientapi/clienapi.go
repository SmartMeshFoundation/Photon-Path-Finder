package clientapi

import (
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/basecomponent"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/routing"
)

//SetupClientAPIComponent set up client api component,reserved for s-s api
func SetupClientAPIComponent(
	base *basecomponent.BasePathFinder,
	balanceDB *storage.Database,
) {
	routing.Setup(
		base.APIMux,
		*base.Cfg,
		balanceDB,
	)
}
