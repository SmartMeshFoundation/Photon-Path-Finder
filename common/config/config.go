package config

import (
	"io/ioutil"
	"path/filepath"
	yaml "gopkg.in/yaml.v2"
	"github.com/sirupsen/logrus"
	"io"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
	"fmt"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
)

const MaxPathRedundancyFactorNum  = 20

const Version = 0

type configErrors []string

var PrivKey *ecdsa.PrivateKey

//The ethereum address you would like smartraiden monitoring to use sign transaction on ethereum
var Address common.Address

// DataSource for opening a postgresql database using lib/pq.
type DataSource string

type PathFinder struct {
	Version int `yaml:"version"`

	Address string `yaml:"address"`

	RegistryAddress string `yaml:"registory_address"`

	KeystorePath string `yaml:"keystore_path"`

	EthRPCEndpoint string `yaml:"eth_rpc_endpoint"`

	PasswordFile string `yaml:"password-file"`

	ChainID int `yaml:"chain_id"`

	MatrixServerAddress string `yaml:"matrix_server_address"`

	Pfs     struct {
		ServerName string `yaml:"server_name"`
		//APIPort             int
		//APIPath             string
	} `yaml:"pfs"`

	RateLimited struct {
		MaxPathPerRequest        int    `yaml:"max_path_per_request"`
		MinPathRedundancy        int    `yaml:"min_path_redundancy"`
		PathRedundancyFactor     int    `yaml:"path_redundancy_factor"`
		DiversityPenDefault      int    `yaml:"diversity_pen_default"`
		StationaryFeeRateDefault string `yaml:"stationary_feerate_default"`
	} `yaml:"ratelimited"`

	// The config for logging informations. Each hook will be added to logrus.
	Logging []LogrusHook `yaml:"logging"`

	// The config for tracing the dendrite servers.
	Tracing struct {
		// The config for the jaeger opentracing reporter.
		Jaeger jaegerconfig.Configuration `yaml:"jaeger"`
	} `yaml:"tracing"`

	httpSync struct {
		MaxSyncPerRequest int `yaml:"max_sync_per_request"`
	} `yaml:"sync"`

	// Postgres config
	Database struct {
		NodeInfos DataSource `yaml:"nodeinfos"`
		//Fee     DataSource `yaml:"fee_rate"`
	} `yaml:"database"`
}

// LogrusHook represents a single logrus hook.
type LogrusHook struct {
	// The type of hook, currently only "file" is supported.
	Type string `yaml:"type"`
	// The level of the logs to produce. Will output only this level and above.
	Level string `yaml:"level"`
	// The parameters for this hook.
	Params map[string]interface{} `yaml:"params"`
}

// Path a path on the filesystem
type Path string

func Load(configPath string) (*PathFinder, error) {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	basePath, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	// Pass the current working directory and ioutil.ReadFile so that they can
	// be mocked in the tests
	monolithic := false
	return loadConfig(basePath, configData, ioutil.ReadFile, monolithic)
}

func LoadMonolithic(configPath string) (*PathFinder, error) {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	basePath, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	// Pass the current working directory and ioutil.ReadFile so that they can
	// be mocked in the tests
	monolithic := true
	return loadConfig(basePath, configData, ioutil.ReadFile, monolithic)
}

func loadConfig(
	basePath string,
	configData []byte,
	readFile func(string) ([]byte, error),
	monolithic bool,
) (*PathFinder, error) {
	var config PathFinder
	var err error
	if err = yaml.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	config.setDefaults()

	if err = config.check(monolithic); err != nil {
		return nil, err
	}


	return &config, nil
}

// SetupTracing configures the opentracing using the supplied configuration.
func (config *PathFinder) SetupTracing(serviceName string) (closer io.Closer, err error) {
	return config.Tracing.Jaeger.InitGlobalTracer(
		serviceName,
		jaegerconfig.Logger(logrusLogger{logrus.StandardLogger()}),
		jaegerconfig.Metrics(jaegermetrics.NullFactory),
	)
}

// logrusLogger is a small wrapper that implements jaeger.Logger using logrus.
type logrusLogger struct {
	l *logrus.Logger
}
// Error error info of logrus logger
func (l logrusLogger) Error(msg string) {
	l.l.Error(msg)
}

// Infof interface of print infomation
func (l logrusLogger) Infof(msg string, args ...interface{}) {
	l.l.Infof(msg, args...)
}

func (config *PathFinder) setDefaults() {
	if config.Pfs.ServerName == "" {
		config.Pfs.ServerName = "localhost"
	}

	if config.RegistryAddress == "" {
		config.RegistryAddress = "0xd66d3719E89358e0790636b8586b539467EDa596"
	}
	if config.KeystorePath == "" {
		//config.KeystorePath = DefaultKeyStoreDir()
	}
	if config.EthRPCEndpoint == "" {
		//config.EthRPCEndpoint = node.DefaultIPCEndpoint("geth")
	}
	if config.RateLimited.MaxPathPerRequest == 0 {
		config.RateLimited.MaxPathPerRequest = 25
	}
	if config.RateLimited.MinPathRedundancy == 0 {
		config.RateLimited.MinPathRedundancy = 20
	}
	if config.RateLimited.PathRedundancyFactor == 0 {
		config.RateLimited.PathRedundancyFactor = 4
	}
	if config.RateLimited.DiversityPenDefault == 0 {
		config.RateLimited.DiversityPenDefault = 1000
	}
	if config.RateLimited.StationaryFeeRateDefault == "" {
		config.RateLimited.StationaryFeeRateDefault = "0.0001"
	}
	if config.ChainID==0{
		config.ChainID=8888
	}
	if config.MatrixServerAddress==""{
		config.MatrixServerAddress="transport01.smartmesh.cn"
	}
}

/*// DefaultKeyStoreDir keystore path of ethereum
func DefaultKeyStoreDir() string {
	return filepath.Join(node.DefaultDataDir(), "keystore")
}*/

// Add appends an error to the list of errors in this configErrors.
func (errs *configErrors) Add(str string) {
	*errs = append(*errs, str)
}

// Error returns a string detailing how many errors were contained within a configErrors type.
func (errs configErrors) Error() string {
	if len(errs) == 1 {
		return errs[0]
	}
	return fmt.Sprintf(
		"%s (and %d other problems)", errs[0], len(errs)-1,
	)
}
//check
func (config *PathFinder) check(monolithic bool) error {
	var configErrs configErrors

	if config.Version != Version {
		configErrs.Add(fmt.Sprintf(
			"unknown config version %q, expected %q", config.Version, Version,
		))
		return configErrs
	}
	config.checkPfs(&configErrs)

	if !monolithic{
		config.checkListen(&configErrs)
	}
	if configErrs!=nil{
		return configErrs
	}
	return nil
}
func (config *PathFinder) checkListen(configErrs *configErrors) {
	//todo s-s later
}
func (config *PathFinder) checkPfs(configErrs *configErrors) {
	checkNotEmpty(configErrs, "pfs.server_name", string(config.Pfs.ServerName))
}
// checkNotEmpty verifies the given value is not empty in the configuration.
func checkNotEmpty(configErrs *configErrors, key, value string) {
	if value == "" {
		configErrs.Add(fmt.Sprintf("missing config key %q", key))
	}
}
// absPath returns the absolute path for a given relative or absolute path.
func absPath(dir string, path Path) string {
	if filepath.IsAbs(string(path)) {
		return filepath.Clean(string(path))
	}
	return filepath.Join(dir, string(path))
}
