package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/phanirithvij/fate/f8/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User a simple user struct
type User struct {
	gorm.Model
	entity.BaseEntity `gorm:"embedded"`
	Username          string `gorm:"unique;not null" json:"username"`
	Name              string `json:"name" gorm:"unique;not null"`
	// Emails            []Email `json:"emails"`
	Emails pq.StringArray `gorm:"type:varchar(254)[]" json:"emails"`
}

// Email email for the user
// type Email struct {
// 	gorm.Model
// 	Email  string `gorm:"uniqueindex:compositeindex" json:"email"`
// 	UserID string `gorm:"uniqueindex:compositeindex"`
// }

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

// postgres pgadmin javascript mime type unblock on windows
// https://www.pgadmin.org/faq/
// https://stackoverflow.com/questions/39228657/disable-chrome-strict-mime-type-checking#comment114712270_58133872

const (
	username = "postgres"
	password = "522191"
	hostname = "localhost"
	port     = "5433"
	// port     = "5432"
	dbame = "gorm"
)

// Main entrypoint for hacky development tests
func main() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, hostname, port, dbame,
	)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// TODO debug flag
	if true {
		db = db.Debug()
	}
	// Not supported error
	// db.DryRun = true

	err = AutoMigrate()
	if err != nil {
		log.Println("AutoMigrate failed")
		log.Fatal(err)
	}

	user := new(User)
	user.BaseEntity = *entity.NewBase()
	// user.Emails = []Email{{Email: "pano@fm.dm"}, {Email: "dodo@gmm.ff"}}
	user.Emails = pq.StringArray{"pano@fm.dm", "dodo@gmm.ff"}
	user.Name = "Phano"
	user.Username = "phano" + strconv.FormatInt(time.Now().Unix(), 10)
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

// AutoMigrate the user's schema
func AutoMigrate() (err error) {
	u := &User{}
	err = entity.AutoMigrate(db)
	if err != nil {
		return err
	}
	// err = db.AutoMigrate(u, &Email{})
	err = db.AutoMigrate(u)
	// db.Model(u).
	return err
}
