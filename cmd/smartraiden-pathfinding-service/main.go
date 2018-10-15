package main

import (
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/basecomponent"
	"flag"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"net/http"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/SmartMeshFoundation/SmartRaiden/network/helper"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/blockchainlistener"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/SmartMeshFoundation/SmartRaiden/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	httpBindAddr  = flag.String("http-bind-address", ":9001", "The HTTP listening port for the server")
)

func main()  {
	StartMain()
}

//StartMain start path finding service
func StartMain() {
	cfg := basecomponent.ParseMonolithFlags()
	base := basecomponent.NewBasePathFinder(cfg, "PathFinder")
	defer base.Close()
	logrus.Info("Welcome to smartraiden-path-finder,version ",base.Cfg.Version)
	PfsDB := base.CreatePfsDB()
	// Setup PFS service interface
	clientapi.SetupClientAPIComponent(
		base,
		PfsDB,
	)
	httpHandler := common.WrapHandlerInCORS(base.APIMux)
	http.Handle("/pathfinder", promhttp.Handler())
	http.Handle("/", httpHandler)

	// Connect to geth and listening block chain
	ethEndpoint:=cfg.EthRPCEndpoint
	client,err:=helper.NewSafeClient(ethEndpoint)
	if err!=nil{
		logrus.Fatalf("Cannot connect to geth :%s err=%s",ethEndpoint, err)
	}
	address:=ethcommon.HexToAddress(base.Cfg.Address)
	address,privkeyBin,err:=accounts.PromptAccount(address,base.Cfg.KeystorePath,base.Cfg.PasswordFile)

	if err!=nil{
		logrus.Fatalf("error :%s", err)
	}
	config.Address=address
	config.PrivKey,err=crypto.ToECDSA(privkeyBin)
	if err!=nil{
		logrus.Fatalf("privkey error :%s", err)
	}
	ce:=blockchainlistener.NewChainEvents(config.PrivKey,client,ethcommon.HexToAddress(base.Cfg.RegistryAddress))
	ce.Start()
	//==============================================

	// Expose the PFS APIs directly,Handle http
	go func() {
		logrus.Info("PFS listening on ", *httpBindAddr)
		logrus.Fatal(http.ListenAndServe(*httpBindAddr, nil))

	}()
	// block forever to let the HTTP and HTTPS handler serve the APIs
	select {}
}