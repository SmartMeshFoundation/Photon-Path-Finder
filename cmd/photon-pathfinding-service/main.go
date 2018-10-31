package main

import (
	"fmt"
	"os/signal"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/rest"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/blockchainlistener"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/internal/debug"
	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"os"
	debug2 "runtime/debug"

	"github.com/SmartMeshFoundation/Photon/log"
	"github.com/SmartMeshFoundation/Photon/network/helper"
	"github.com/SmartMeshFoundation/Photon/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/node"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	StartMain()
}

func init() {
	debug2.SetTraceback("crash")
	log.LocationTrims = append(log.LocationTrims,
		"github.com/SmartMeshFoundation/Photon-Path-Finder/vendor/github.com/SmartMeshFoundation/Photon/",
		"github.com/SmartMeshFoundation/Photon-Path-Finder",
	)

}

//StartMain entry point of Photon app
func StartMain() {
	fmt.Printf("os.args=%q\n", os.Args)
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "eth-rpc-endpoint",
			Usage: `"host:port" address of ethereum JSON-RPC server.\n'
	           'Also accepts a protocol prefix (ws:// or ipc channel) with optional port',`,
			Value: node.DefaultIPCEndpoint("geth"),
		},
		cli.StringFlag{
			Name:  "registry-contract-address",
			Usage: `hex encoded address of the registry contract.`,
			Value: params.RegistryAddress.String(),
		},
		cli.IntFlag{
			Name:  "port",
			Usage: ` port  for the RPC server to listen on.`,
			Value: 7000,
		},
		cli.StringFlag{
			Name:  "dbtype",
			Usage: "database type sqlite3/mysql/postgres",
			Value: "sqlite3",
		},
		cli.StringFlag{
			Name:  "dbconnection",
			Usage: "database connection string",
			Value: "./photon.db",
		},
	}
	app.Flags = append(app.Flags, debug.Flags...)
	app.Action = mainCtx
	app.Name = "PhotonPathFinder"
	app.Version = "0.1"
	app.Before = func(ctx *cli.Context) error {
		if err := debug.Setup(ctx); err != nil {
			return err
		}
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Error(fmt.Sprintf("run err %s", err))
	}
}

func mainCtx(ctx *cli.Context) error {

	var err error
	fmt.Printf("Welcom to Photon Path Finder,version %s\n", ctx.App.Version)
	config(ctx)
	//log.Debug(fmt.Sprintf("Config:%s", utils.StringInterface(cfg, 2)))
	ethEndpoint := ctx.String("eth-rpc-endpoint")
	client, err := helper.NewSafeClient(ethEndpoint)
	if err != nil {
		log.Error(fmt.Sprintf("cannot connect to geth :%s err=%s", ethEndpoint, err))
		utils.SystemExit(1)
	}
	params.DBType = ctx.String("dbtype")
	params.DBPath = ctx.String("dbconnection")
	model.SetUpDB(params.DBType, params.DBPath)
	key, _ := utils.MakePrivateKeyAddress()
	ce := blockchainlistener.NewChainEvents(key, client, params.RegistryAddress)
	err = ce.Start()
	if err != nil {
		log.Error(fmt.Sprintf("ce start err =%s ", err))
		utils.SystemExit(3)
	}
	/*
		quit handler
	*/
	go func() {
		quitSignal := make(chan os.Signal, 1)
		signal.Notify(quitSignal, os.Interrupt, os.Kill)
		<-quitSignal
		signal.Stop(quitSignal)
		ce.Stop()
		model.CloseDB()
		utils.SystemExit(0)
	}()
	rest.Start(ce, ce.TokenNetwork)
	return nil
}
func config(ctx *cli.Context) {

	params.Port = ctx.Int("port")
	registAddrStr := ctx.String("registry-contract-address")
	if len(registAddrStr) > 0 {
		params.RegistryAddress = common.HexToAddress(registAddrStr)
	}

}
