package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	ID      uint
	Name    string
	PhoneID uint
	Phone   Phone
}

type Phone struct {
	ID    uint
	Color string
	Brand string
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/jenkins?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	//db.AutoMigrate(&Student{})
	//var r Student
	//db.Create(&Phone{ID: 1, Color: "yellow", Brand: "oppo"})
	fmt.Println(db.Create(&Student{ID: 2, Name: "lishi", PhoneID: 1}).Error)
	//db.Preload("Phone").Take(&r)
	//fmt.Println(r)

}
