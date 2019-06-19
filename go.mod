module github.com/SmartMeshFoundation/Photon-Path-Finder

replace (
	github.com/SmartMeshFoundation/Photon v1.0.0 => github.com/nkbai/Photon v1.2.0-rc0
	github.com/ethereum/go-ethereum v1.8.17 => github.com/nkbai/go-ethereum v0.1.2
	github.com/mattn/go-xmpp v0.0.1 => github.com/nkbai/go-xmpp v0.0.1
)

require (
	github.com/SmartMeshFoundation/Photon v1.0.0
	github.com/SmartMeshFoundation/matrix-regservice v0.0.0-20190219025223-14bc68e5eba7 // indirect
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/ethereum/go-ethereum v1.8.17
	github.com/jinzhu/gorm v1.9.1
	github.com/mattn/go-colorable v0.1.0
	github.com/mattn/go-xmpp v0.0.1
	github.com/nkbai/goutils v0.0.0-20181219015612-2fa82e8abe13
	github.com/nkbai/log v0.0.0-20180519141659-86998e435e8c // indirect
	github.com/stretchr/testify v1.2.2
	gopkg.in/urfave/cli.v1 v1.20.0
)
