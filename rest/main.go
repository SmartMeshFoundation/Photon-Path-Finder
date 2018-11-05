package rest

import (
	"fmt"
	"net/http"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"
	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/ant0ine/go-json-rest/rest"
)

var ce *blockchainlistener.ChainEvents
var tn *blockchainlistener.TokenNetwork

/*
Start the restful server
*/
func Start(e *blockchainlistener.ChainEvents, t *blockchainlistener.TokenNetwork) {
	ce = e
	tn = t
	api := rest.NewApi()
	if params.DebugMode {
		api.Use(rest.DefaultDevStack...)
	} else {
		api.Use(rest.DefaultProdStack...)
	}

	router, err := rest.MakeRouter(
		//peer 提交Partner的BalanceProof,更新Partner的余额
		rest.Put("/pfs/1/:peer/balance", UpdateBalanceProof),
		rest.Put("/pfs/1/channel_rate/:channel/:peer", setChannelRate),
		rest.Get("/pfs/1/channel_rate/:channel/:peer", getChannelRate),
		rest.Put("/pfs/1/token_rate/:token/:peer", setTokenRate),
		rest.Get("/pfs/1/token_rate/:token/:peer", getTokenRate),
		rest.Put("/pfs/1/account_rate/:peer", setAccountRate),
		rest.Get("/pfs/1/account_rate/:peer", getAccountRate),
		rest.Post("/pfs/1/paths", GetPaths),
	)
	if err != nil {
		log.Crit(fmt.Sprintf("maker router :%s", err))
	}
	api.SetApp(router)
	listen := fmt.Sprintf("0.0.0.0:%d", params.Port)
	log.Crit(fmt.Sprintf("http listen and serve :%s", http.ListenAndServe(listen, api.MakeHandler())))
}
