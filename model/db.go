package model

import (
	"fmt"

	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

//SetUpDB init db
func SetUpDB(dbtype, path string) {
	var err error
	db, err = gorm.Open(dbtype, path)
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)
	//db.SetLogger(gorm.Logger{revel.TRACE})
	db.SetLogger(log.New(os.Stdout, "\r\n", 0))
	db.AutoMigrate(&Channel{})
	if err = db.AutoMigrate(&ChannelParticipantInfo{}).Error; err != nil {
		panic(err)
	}
	db.Model(&ChannelParticipantInfo{}).AddForeignKey("channel_id", "channels(id)", "CASCADE", "CASCADE") // Foreign key need to define manually
	db.AutoMigrate(&SettledChannel{})
	db.AutoMigrate(&latestBlockNumber{})
	db.AutoMigrate(&tokenNetwork{})
	db.AutoMigrate(&AccountFee{}, &AccountTokenFee{}, &TokenFee{}, &NodeStatus{})
	db = db.Debug()
	db.Create(lb)
	return
}

func CloseDB() {
	err := db.Close()
	if err != nil {
		log.Printf(fmt.Sprintf("closedb err %s", err))
	}
}

func SetupTestDB() {
	dbPath := "/tmp/test.db"
	os.Remove(dbPath)
	SetUpDB("sqlite3", dbPath)
}
