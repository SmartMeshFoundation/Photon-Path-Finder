package model3

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//var Db *gorm.Db
type Place struct {
	ID     int `gorm:primary_key`
	Name   string
	Town   Town
	TownID int `gorm:"ForeignKey:id"` //this foreignKey tag didn't works
}

type Town struct {
	ID   int `gorm:"primary_key"`
	Name string
}

func TestA(t *testing.T) {
	//Init Db connection

	Db, _ := gorm.Open("sqlite3", "test2.db")
	defer Db.Close()

	Db.DropTableIfExists(&Place{}, &Town{})

	Db.AutoMigrate(&Place{}, &Town{})
	//We need to add foreign keys manually.
	Db.Model(&Place{}).AddForeignKey("town_id", "towns(id)", "CASCADE", "CASCADE")

	t1 := Town{
		Name: "Pune",
	}
	t2 := Town{
		Name: "Mumbai",
	}
	t3 := Town{
		Name: "Hyderabad",
	}

	p1 := Place{
		Name: "Katraj",
		Town: t1,
	}
	p2 := Place{
		Name: "Thane",
		Town: t2,
	}
	p3 := Place{
		Name: "Secundarabad",
		Town: t3,
	}

	Db.Save(&p1) //Saving one to one relationship
	Db.Save(&p2)
	Db.Save(&p3)

	fmt.Println("t1==>", t1, "p1==>", p1)
	fmt.Println("t2==>", t2, "p2s==>", p2)
	fmt.Println("t2==>", t3, "p2s==>", p3)

	//Delete
	Db.Where("name=?", "Hyderabad").Delete(&Town{})

	//Update
	Db.Model(&Place{}).Where("id=?", 1).Update("name", "Shivaji Nagar")

	//Select
	places := Place{}
	towns := Town{}
	fmt.Println("Before Association", places)
	Db.Where("name=?", "Shivaji Nagar").Find(&places)
	fmt.Println("After Association", places)
	err := Db.Model(&places).Association("town").Find(&places.Town).Error
	fmt.Println("After Association", towns, places, err)

	defer Db.Close()
}
