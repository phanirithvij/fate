package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/lib/pq"
	"github.com/phanirithvij/fate/f8/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User a simple user struct
type User struct {
	gorm.Model
	*entity.Base `gorm:"embedded"`
	Name         string         `json:"name"`
	Emails       pq.StringArray `gorm:"type:varchar(254)[]" json:"emails"`
}

func (u User) String() string {
	x, err := json.MarshalIndent(u, "", " ")
	if err != nil {
		// https://stackoverflow.com/a/64306225/8608146
		type userX User
		ux := (userX)(u)
		return fmt.Sprintf("%v", ux)
	}
	return string(x)
}

var (
	db *gorm.DB
)

// Main entrypoint for hacky development tests
func main() {
	dsn := "postgres://postgres:522191@localhost:5432/gorm?sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.Debug()

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Println("AutoMigrate failed")
		log.Fatal(err)
	}
	// var u int
	// tx := db.Raw("").Scan(&u)
	// sdb, err := tx.DB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// sdb.Query()

	user := new(User)
	user.Base = entity.NewBase()
	user.Emails = pq.StringArray{"3", "3"}
	user.Name = "Phano"
	err = user.Register()
	if err != nil {
		log.Fatal(err)
	}
}

// Register a user
func (u User) Register() error {
	tx := db.Create(&u)
	fmt.Println(u)
	return tx.Error
}
