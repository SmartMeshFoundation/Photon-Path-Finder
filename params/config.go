package params

import (
	"math/big"
)

//ChainID for the block chain
var ChainID = big.NewInt(8888)

//DefaultFeePolicy 缺省按比例收费
var DefaultFeePolicy = 1 //model3.FeePolicyPercent

//DefaultFeeConstantPart 收费固定部分为0
var DefaultFeeConstantPart = big.NewInt(0)

//DefaultFeePercentPart 比例缺省万分之一
var DefaultFeePercentPart int64 = 10000
