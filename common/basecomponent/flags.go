package basecomponent

import (
	"flag"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/sirupsen/logrus"
)

// configPath the path to the config file
var configPath = flag.String("config", "pathfinder.yaml", "The path to the config file. For more information, see the config file in this repository.")

// ParseFlags parses the commandline flags and uses them to create a config.
func ParseFlags() *config.PathFinder {
	flag.Parse()

	if *configPath==""{
		logrus.Fatal("--config must be supplied")
	}

	cfg,err:=config.Load(*configPath)

	if err!=nil{
		logrus.Fatalf("Invalid config file: %s", err)
	}
	return cfg
}

func ParseMonolithFlags() *config.PathFinder {
	flag.Parse()
	if *configPath == "" {
		logrus.Fatal("--config must be supplied")
	}
	cfg, err := config.LoadMonolithic(*configPath)
	if err != nil {
		logrus.Fatalf("Invalid config file: %s", err)
	}
	return cfg
}

