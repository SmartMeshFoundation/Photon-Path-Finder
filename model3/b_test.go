package model3

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
)

type Customer struct {
	CustomerID   int `gorm:"primary_key"`
	CustomerName string
	Contacts     []Contact `gorm:"ForeignKey:CustID"` //you need to do like this
}

type Contact struct {
	ContactID   int `gorm:"primary_key"`
	CountryCode int
	MobileNo    uint
	CustID      int
}

func TestB(t *testing.T) {

	db, err := gorm.Open("sqlite3", "test2.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	db.DropTableIfExists(&Contact{}, &Customer{})
	db.AutoMigrate(&Customer{}, &Contact{})
	db.Model(&Contact{}).AddForeignKey("cust_id", "customers(customer_id)", "CASCADE", "CASCADE") // Foreign key need to define manually

	Custs1 := Customer{CustomerName: "John", Contacts: []Contact{
		{CountryCode: 91, MobileNo: 956112},
		{CountryCode: 91, MobileNo: 997555}}}

	Custs2 := Customer{CustomerName: "Martin", Contacts: []Contact{
		{CountryCode: 90, MobileNo: 808988},
		{CountryCode: 90, MobileNo: 909699}}}

	Custs3 := Customer{CustomerName: "Raym", Contacts: []Contact{
		{CountryCode: 75, MobileNo: 798088},
		{CountryCode: 75, MobileNo: 965755}}}

	Custs4 := Customer{CustomerName: "Stoke", Contacts: []Contact{
		{CountryCode: 80, MobileNo: 805510},
		{CountryCode: 80, MobileNo: 758863}}}

	db.Create(&Custs1)
	db.Create(&Custs2)
	db.Create(&Custs3)
	db.Create(&Custs4)

	customers := &Customer{}
	contacts := &Contact{}

	db.Debug().Where("customer_name=?", "Martin").Preload("Contacts").Find(customers) //db.Debug().Where("customer_name=?","John").Preload("Contacts").Find(&customers)
	fmt.Println("Customers", customers)
	fmt.Println("Contacts", contacts)

	//Update
	db.Debug().Model(&Contact{}).Where("cust_id=?", 3).Update("country_code", 77)
	//Delete
	db.Debug().Where("customer_name=?", customers.CustomerName).Delete(&customers)
	fmt.Println("After Delete", customers)

}
