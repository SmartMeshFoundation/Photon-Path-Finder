package main

import (
	"flag"
	"fmt"
	"net/http"
	debug2 "runtime/debug"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/clientapi"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common/basecomponent"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/Photon/accounts"
	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/network/helper"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	// httpBindAddr http listening port is 9001
	httpBindAddr = flag.String("http-bind-address", ":9001", "The HTTP listening port for the server")
)

func init() {
	debug2.SetTraceback("crash")
	log.LocationTrims = append(log.LocationTrims,
		"github.com/SmartMeshFoundation/Photon-Path-Finder/vendor/github.com/SmartMeshFoundation/Photon/",
		"github.com/SmartMeshFoundation/Photon-Monitoring",
	)

}

// main main
func main() {
	StartMain()
}

// StartMain Start path finding service
func StartMain() {
	cfg := basecomponent.ParseMonolithFlags()
	base := basecomponent.NewBasePathFinder(cfg, "PathFinder")
	defer base.Close()
	logrus.Info("Welcome to Photon-path-finder,version ", base.Cfg.Version)
	PfsDB := base.CreatePfsDB()

	httpHandler := common.WrapHandlerInCORS(base.APIMux)
	http.Handle("/pathfinder", promhttp.Handler())
	http.Handle("/", httpHandler)

	// Connect to geth and listening block chain events
	ethEndpoint := cfg.EthRPCEndpoint
	client, err := helper.NewSafeClient(ethEndpoint)
	if err != nil {
		logrus.Fatalf("Cannot connect to geth :%s err=%s", ethEndpoint, err)
	}
	address := ethcommon.HexToAddress(base.Cfg.Address)
	address, privkeyBin, err := accounts.PromptAccount(address, base.Cfg.KeystorePath, base.Cfg.PasswordFile)

	if err != nil {
		logrus.Fatal("error :", err)
	}
	config.Address = address
	config.PrivKey, err = crypto.ToECDSA(privkeyBin)
	if err != nil {
		logrus.Fatal("privkey error :", err)
	}
	ce := blockchainlistener.NewChainEvents(config.PrivKey, client, ethcommon.HexToAddress(base.Cfg.RegistryAddress), PfsDB)
	err = ce.Start()
	if err != nil {
		log.Crit(fmt.Sprintf("ce start err %s", err))
	}

	// Setup PFS service interface
	clientapi.SetupClientAPIComponent(
		base,
		PfsDB,
		ce,
	)

	// Expose the PFS APIs directly,Handle http
	go func() {
		logrus.Info("PFS listening on ", *httpBindAddr)
		logrus.Fatal(http.ListenAndServe(*httpBindAddr, nil))

	}()
	// block forever to let the HTTP handler serve the APIs
	select {}
}
