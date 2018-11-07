module github.com/SmartMeshFoundation/Photon-Path-Finder

require (
	github.com/SmartMeshFoundation/Photon v0.9.3
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/ethereum/go-ethereum v1.8.17
	github.com/jinzhu/gorm v1.9.1
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/kataras/go-errors v0.0.3
	github.com/mattn/go-colorable v0.0.9
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/nkbai/dijkstra v0.1.0
	golang.org/x/crypto v0.0.1
	golang.org/x/net v0.0.1
	golang.org/x/sys v0.0.1
	golang.org/x/tools v0.0.1
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace (
	github.com/ethereum/go-ethereum v1.8.17 => github.com/nkbai/go-ethereum v1.9.1
	github.com/mattn/go-xmpp v0.0.1 => github.com/nkbai/go-xmpp v0.0.1
	golang.org/x/crypto v0.0.1 => github.com/golang/crypto v0.0.0-20181106171534-e4dc69e5b2fd
	golang.org/x/net v0.0.1 => github.com/golang/net v0.0.0-20181106065722-10aee1819953
	golang.org/x/sys v0.0.1 => github.com/golang/sys v0.0.0-20181106135930-3a76605856fd
	golang.org/x/tools v0.0.1 => github.com/golang/tools v0.0.0-20181106213628-e21233ffa6c3
)
