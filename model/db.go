package model

import (
	"fmt"
	log2 "log"
	"path"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/SmartMeshFoundation/Photon/log"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/params"

	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //for gorm
	_ "github.com/jinzhu/gorm/dialects/sqlite"   //for gorm
)

var db *gorm.DB

//SetUpDB init db
func SetUpDB(dbtype, path string) {
	var err error
	db, err = gorm.Open(dbtype, path)
	if err != nil {
		panic("failed to connect database")
	}
	if dbtype=="sqlite3"{
		//由于sqlite多线程访问性能很低,测试的时候用sqlite都会效率很低
		//db.DB().SetMaxIdleConns(1)
		//db.DB().SetMaxOpenConns(1)
	}
	//if params.DebugMode {
	//	db = db.Debug()
	//	db.LogMode(true)
	//}
	//db.SetLogger(gorm.Logger{revel.TRACE})
	db.SetLogger(log2.New(os.Stdout, "\r\n", 0))
	db.AutoMigrate(&Channel{})
	if err = db.AutoMigrate(&ChannelParticipantInfo{}).Error; err != nil {
		panic(err)
	}
	//db.Model(&ChannelParticipantInfo{}).AddForeignKey("channel_id", "channels(channel_id)", "CASCADE", "CASCADE") // Foreign key need to define manually
	db.AutoMigrate(&SettledChannel{})
	db.AutoMigrate(&latestBlockNumber{})
	db.AutoMigrate(&tokenNetwork{})
	db.AutoMigrate(&AccountFee{}, &AccountTokenFee{}, &TokenFee{}, &NodeStatus{})
	db.AutoMigrate(&xmpp{})
	db.AutoMigrate(&observerKey{})
	db.AutoMigrate(&ChannelParticipantFee{})
	db.FirstOrCreate(lb)
	params.ObserverKey = GetObserverKey
	return
}

//CloseDB release connection
func CloseDB() {
	err := db.Close()
	if err != nil {
		log.Error(fmt.Sprintf("closedb err %s", err))
	}
}

//SetupTestDB for test only
func SetupTestDB() {
	dbPath := path.Join(os.TempDir(), fmt.Sprintf("test%s.db", utils.RandomString(10)))
	log.Trace(fmt.Sprintf(dbPath))
	err := os.Remove(dbPath)
	if err != nil {
		log.Error(fmt.Sprintf("remove err %s",err))
	}
	SetUpDB("sqlite3", dbPath)
}
