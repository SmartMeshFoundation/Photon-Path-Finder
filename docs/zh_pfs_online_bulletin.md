# Photon Path Finding Service Online bulletin 

PFS路由收费查询服务已经在`Spectrum`主网和测试网上线，开发者可以直接使用，下面是一些需要更改的服务配置

- 合约地址
- pfs'host


## Spectrum 测试网

参数|参数信息
--|--
合约地址|0xa2150A4647908ab8D0135F1c4BFBB723495e8d12
pfs'ip|transport01.smartmesh.cn
端口|7001


## Spectrum 主网

参数|参数信息
--|--
合约地址|0x28233F8e0f8Bd049382077c6eC78bE9c2915c7D4
pfs'ip|transport01.smartmesh.cn
端口|7000

## 怎么使用？

- 确保photon节点合约地址与pfs合约地址一致。例如你在测试网上使用pfs，则对应的节点启动参数合约地址应该是：`0xa2150A4647908ab8D0135F1c4BFBB723495e8d12`
- photon节点启动时 需要加上 `--fee` 和 `--pfs`参数
  - `fee` 收费参数，添加这个参数后，photon才能设置/查询收费
  - `pfs` pfs'host，添加这个参数后，photon本地设置的收费信息才能更新到pfs服务器上，例如测试网上参数信息应该是：`--pfs  http://transport01.smartmesh.cn:7001` 

- 主要接口,详细参数信息请参考[接口文档](https://photonnetwork.readthedocs.io/en/latest/pfs/)
  - 更新balance_proof `Put /pfs/1/:peer/balance`
  - 查询节点在某个通道收费信息 `Get /pfs/1/channel_rate/:channel/:peer`
  - 查询节点在某种token上收费信息 `Get /pfs/1/token_rate/:token/:peer`
  - 查询节点(account)收费信息 `Get /pfs/1/account_rate/:peer`
  - 查询路由 `Post /pfs/1/paths`


- photon 设置[收费接口](https://photonnetwork.readthedocs.io/en/latest/rest_api/#post-api1fee_policy)



`POST /api/1/fee_policy`

请求参数
```json
{
    "account_fee":{
        "fee_constant":5,
        "fee_percent":10000
    },
    "token_fee_map":{
        "0x83073FCD20b9D31C6c6B3aAE1dEE0a539458d0c5":{
            "fee_constant":5,
            "fee_percent":10000
        }
    },
    "channel_fee_map":{
        "0xa7712241a1a10abdada1c228c6935a71a9db80aa0bf2a13b59940159aa4eb4b5":{
            "fee_constant":5,
            "fee_percent":10000
        }
    }
}
```
- fee_constant  固定收费费用
- fee_percent  比例收费费率 

其中`fee_constant`为固定费率,比如5代表手续费固定部分为5个token,设置为0即不收费

`fee_percent`为比例费率,计算方式为 交易金额/fee_percent,比如交易金额50000,fee_percent=10000,那么手续费比例部分=50000/10000=5,设置为0即不收费

总收费费用=固定收费+比例收费

一个节点有三种收费方式：
- account_fee 节点收费
- token_fee  节点在某种token上收费
- channel_fee  节点在某个通道收费

三种收费方式的优先级是：`channel_fee`>`token_fee`>`account_fee`