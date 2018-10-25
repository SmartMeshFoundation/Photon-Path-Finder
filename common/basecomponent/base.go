package basecomponent

import (
	"io"
	"github.com/gorilla/mux"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common/config"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/common"
	"github.com/sirupsen/logrus"
	"github.com/SmartMeshFoundation/SmartRaiden-Path-Finder/clientapi/storage"
)

type BasePathFinder struct {
	componentName string
	tracerCloser  io.Closer
	APIMux        *mux.Router
	Cfg           *config.PathFinder
}

func NewBasePathFinder(cfg *config.PathFinder,componentName string) *BasePathFinder{
	common.SetupStdLogging()
	common.SetupHookLogging(cfg.Logging,componentName)

	closer,err:=cfg.SetupTracing("SmartRaiden"+componentName)
	if err!=nil{
		logrus.WithError(err).Panicf("failed to start opentracing")
	}

	return &BasePathFinder{
		componentName:componentName,
		tracerCloser:  closer,
		APIMux:        mux.NewRouter(),
		Cfg:           cfg,
	}
}

// Close implements io.Closer
func (bpf *BasePathFinder) Close() error {
	return bpf.tracerCloser.Close()
}


// CreateDeviceDB creates a new instance of the balance database. Should only be called once per component.
func (bpf *BasePathFinder) CreatePfsDB() *storage.Database {
	db, err := storage.NewDatabase(string(bpf.Cfg.Database.NodeInfos),string(bpf.Cfg.RateLimited.StationaryFeeRateDefault))
	if err != nil {
		logrus.WithError(err).Panicf("failed to connect to progresSql(db)")
	}
	return db
}

