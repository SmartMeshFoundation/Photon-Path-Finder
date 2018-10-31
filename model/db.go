package model3

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
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
	db.AutoMigrate(&AccountFee{}, &AccountTokenFee{}, &TokenFee{})
	db = db.Debug()
	db.Create(lb)
	return
}
