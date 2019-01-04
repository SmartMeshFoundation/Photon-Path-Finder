package model

import (
	"fmt"
	"path"

	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    //for gorm
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
	//if params.DebugMode {
	//	db = db.Debug()
	//	db.LogMode(true)
	//}
	//db.SetLogger(gorm.Logger{revel.TRACE})
	db.SetLogger(log.New(os.Stdout, "\r\n", 0))
	db.AutoMigrate(&Channel{})
	if err = db.AutoMigrate(&ChannelParticipantInfo{}).Error; err != nil {
		panic(err)
	}
	//db.Model(&ChannelParticipantInfo{}).AddForeignKey("channel_id", "channels(channel_id)", "CASCADE", "CASCADE") // Foreign key need to define manually
	db.AutoMigrate(&SettledChannel{})
	db.AutoMigrate(&latestBlockNumber{})
	db.AutoMigrate(&tokenNetwork{})
	db.AutoMigrate(&AccountFee{}, &AccountTokenFee{}, &TokenFee{}, &NodeStatus{})

	db.FirstOrCreate(lb)
	return
}

//CloseDB release connection
func CloseDB() {
	err := db.Close()
	if err != nil {
		log.Printf(fmt.Sprintf("closedb err %s", err))
	}
}

//SetupTestDB for test only
func SetupTestDB() {
	dbPath := path.Join(os.TempDir(), "test.db")
	err := os.Remove(dbPath)
	if err != nil {
		//ignore
	}
	SetUpDB("sqlite3", dbPath)
}
