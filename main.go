package main

import (
	"encoding/json"
	"errors"
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
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
	*entity.BaseEntity `gorm:"embedded"`
	Name               string `json:"name" gorm:"not null"`
	// Emails             []Email `json:"emails" gorm:"foreignKey:UserID;"`
	// Emails []Email `json:"emails" gorm:"polymorphic:User;"`
	Emails pq.StringArray `gorm:"type:varchar(254)[]" json:"emails"`
}

// TODO Consider https://github.com/go-gorm/datatypes for metadata or details

// Email email for the user
type Email struct {
	gorm.Model
	Email    string `gorm:"uniqueindex:user_email_idx" json:"email"`
	UserID   string `gorm:"uniqueindex:user_email_idx"`
	UserType string
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, hostname, port, dbame,
	)
	var err error
	// db, err = gorm.Open(sqlite.Open("sample.db"), &gorm.Config{})
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// TODO debug flag
	if true {
		// db = db.Debug()
	}
	// Not supported error
	// db.DryRun = true

	err = AutoMigrate()
	if err != nil {
		log.Println("AutoMigrate failed")
		log.Fatal(err)
	}

	user := new(User)
	// user.Emails = []Email{{Email: "pano@fm.dm"}, {Email: "dodo@gmm.ff"}}
	user.Emails = pq.StringArray{"pano@fm.dm", "dodo@gmm.ff"}
	user.Name = "Phano"
	// userID := "phano"
	userID := "phano" + strconv.FormatInt(time.Now().Unix(), 10)
	user.BaseEntity, err = entity.Entity(
		entity.ID(userID),
		entity.TableName(user.TableName()),
		entity.BucketName("default"),
		entity.BucketCount(3),
		entity.DB(db),
	)
	err = user.Save()
	fmt.Println(user)

	// Now manuplate the entity's file system

	// get the default bucket
	buck, err := user.GetBucket("")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("Bucket not found")
		}
	}
	fmt.Println(buck)
	fmt.Println("bucket exists?", buck.Exists())
	ok := user.DeleteBucket(buck.ID)
	if ok {
		fmt.Println("Deleted successfully")
	}
	ok = buck.Delete()
	if ok {
		fmt.Println("Deleted successfully twice??")
	}
	fmt.Println("bucket exists?", buck.Exists())

	// if err != nil {
	// 	if perr, ok := err.(*pq.Error); ok {
	// 		// Here err is of type *pq.Error, you may inspect all its fields, e.g.:
	// 		fmt.Println("pq error:", perr.Code.Name())
	// 	} else {
	// 		log.Fatal(err)
	// 	}
	// }
}

// TableName for the user
func (u User) TableName() string {
	return "users"
}

// Save a user
func (u *User) Save() error {
	tx := db.Create(&u)
	return tx.Error
}

// AutoMigrate the user's schema
func AutoMigrate() (err error) {
	u := &User{}
	err = entity.AutoMigrate(db)
	if err != nil {
		return err
	}
	err = db.AutoMigrate(u)
	// err = db.AutoMigrate(u, &Email{})
	return err
}
