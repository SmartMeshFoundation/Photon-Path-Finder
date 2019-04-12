module github.com/SmartMeshFoundation/Photon-Path-Finder

replace (
	github.com/SmartMeshFoundation/Photon v1.0.0 => github.com/nkbai/Photon v1.2.0-rc0
	github.com/ethereum/go-ethereum v1.8.17 => github.com/nkbai/go-ethereum v0.1.2
	github.com/mattn/go-xmpp v0.0.1 => github.com/nkbai/go-xmpp v0.0.1
)

require (
	github.com/SmartMeshFoundation/Photon v1.0.0
	github.com/SmartMeshFoundation/matrix-regservice v0.0.0-20190219025223-14bc68e5eba7
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/ethereum/go-ethereum v1.8.17
	github.com/jinzhu/gorm v1.9.1
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a
	github.com/karalabe/hid v0.0.0-20180420081245-2b4488a37358
	github.com/lib/pq v1.0.0
	github.com/mattn/go-colorable v0.1.0
	github.com/mattn/go-isatty v0.0.4
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/mattn/go-xmpp v0.0.1
	github.com/nkbai/goutils v0.0.0-20181219015612-2fa82e8abe13
	github.com/nkbai/log v0.0.0-20180519141659-86998e435e8c // indirect
	github.com/sirupsen/logrus v1.1.1
	github.com/stretchr/testify v1.2.2
	golang.org/x/crypto v0.0.0-20181106171534-e4dc69e5b2fd
	golang.org/x/net v0.0.0-20181023162649-9b4f9f5ad519
	golang.org/x/sys v0.0.0-20181023152157-44b849a8bc13
	gopkg.in/urfave/cli.v1 v1.20.0
	gopkg.in/yaml.v2 v2.2.1
)
