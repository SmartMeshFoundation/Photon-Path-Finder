package params

import (
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
)

//ChainID for the block chain
var ChainID = big.NewInt(8888)

//DefaultFeePolicy 缺省按比例收费
var DefaultFeePolicy = 1 //model3.FeePolicyPercent

//DefaultFeeConstantPart 收费固定部分为0
var DefaultFeeConstantPart = big.NewInt(0)

//DefaultFeePercentPart 比例缺省万分之一
var DefaultFeePercentPart int64 = 10000

//RegistryAddress contract works on
var RegistryAddress = common.HexToAddress("0xDe661C5aDaF15c243475C5c6BA96634983821593")

//DBType is the default database
var DBType = "sqlite3"

//DBPath is the default database connection string
var DBPath string

//Port is listening service port
var Port int

//ObserverKey is the key login to matrix server to observer other's presence
var ObserverKey = "0bb2d0315029cd6048c0a756c076f3dd80e84cff4ff4bd80aad4a8d1d7f62598"

//MatrixServer the matrix server for path finder use
var MatrixServer = "transport01.smartmesh.cn"

//DebugMode for debug setting
var DebugMode = false

//DefaultDataDir default work directory
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "photonpfs")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "photonpfs")
		} else {
			return filepath.Join(home, ".photonpfs")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}
func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func init() {
	//DBPath = filepath.Join(DefaultDataDir(), "photon.db")
}
