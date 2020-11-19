package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/phanirithvij/fate/f8"
	"github.com/phanirithvij/fate/f8/buckets"
	"github.com/phanirithvij/fate/f8/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// User a simple user struct
type User struct {
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
	*entity.BaseEntity `gorm:"embedded"`
	Name               string `json:"name" gorm:"not null"`
	// Emails             []Email `json:"emails" gorm:"foreignKey:UserID;"`
	Emails []Email `json:"emails" gorm:"polymorphic:User;"`
	// PGSQL
	// Emails pq.StringArray `gorm:"type:varchar(254)[]" json:"emails"`
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
	port     = 5433
	dbame    = "f8"
	// port     = "5432"
)

// Main entrypoint for hacky development tests
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	posgres := &f8.DBConfig{
		DatabaseMode: f8.PgSQL,
		PGusername:   username,
		PGpassword:   password,
		PGhostname:   hostname,
		PGport:       port,
		PGdbname:     dbame,
		GormConfig:   &gorm.Config{},
	}
	db = posgres.PostGreSQLDB()
	storage, err := f8.New(f8.DB(db))
	if err != nil {
		log.Fatal(err)
	}

	// use the instance ourselves
	// db = storage.DB
	// if storage.DB == nil {
	// 	log.Fatal("DB could not be initialized")
	// }

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
	user.Emails = []Email{{Email: "pano@fm.dm"}, {Email: "dodo@gmm.ff"}}
	// PGSQL
	// user.Emails = pq.StringArray{"pano@fm.dm", "dodo@gmm.ff"}
	user.Name = "Phano"
	userID := "phano"
	// userID := "phano" + strconv.FormatInt(time.Now().Unix(), 10)
	user.BaseEntity, err = entity.Entity(
		entity.ID(userID),
		entity.StorageConfig(storage),
		entity.TableName(user.TableName()),
		// entity.BucketName("newDefault"),
		entity.BucketName(""),
		entity.BucketCount(3),
		entity.DB(db),
	)
	fmt.Println(user)
	err = user.Save()
	fmt.Println(user)

	// Now manuplate the entity's file system

	// get the default bucket
	buck, err := user.GetBucket("default")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("Bucket not found")
		}
	}
	fmt.Println(buck)
	// should be true
	fmt.Println("bucket exists?", buck.Exists())
	ok := user.DeleteBucket("default-1")
	if ok {
		fmt.Println("Deleted successfully")
	}
	// ok = buck.Delete()
	// if ok {
	// 	fmt.Println("Deleted successfully twice??")
	// }
	buck, err = user.GetBucket("default-1")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("Bucket1 not found")
		}
	}
	// should be false
	fmt.Println("bucket exists?", buck.Exists())

	ok = user.DeleteBucket("No such bucket")
	if !ok {
		fmt.Println("No, such bucket won't exist")
	}

	bucks := buckets.GetDeletedBuckets(db)
	fmt.Println(len(bucks))

	ok = buckets.CleanupBuckets(db)
	if ok {
		fmt.Println(len(bucks), "buckets permanently deleted successfully")
	}

	if v, ok := entity.EntityBucketMap[user.ID]; ok {
		fmt.Println("Printing entity bucket map")
		jss, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println(v)
		} else {
			fmt.Println(string(jss))
		}
	}
}

// TableName for the user
func (u User) TableName() string {
	return "users"
}

// Save a user
//
// Upserts the user
func (u *User) Save() error {
	cols := []string{"updated_at", "name"}
	// PGSQL
	// cols := []string{"updated_at", "name", "emails"}
	tx := db.Clauses(clause.OnConflict{
		// TODO all primaryKeys not just ID
		Columns: []clause.Column{{Name: "id"}},
		// TODO exept created_at everything
		DoUpdates: clause.AssignmentColumns(cols),
	}).Create(&u)
	log.Println("[main][WARNING]: Hardcoded feilds for user.Save", cols)
	return tx.Error
}

// AutoMigrate the user's schema
func AutoMigrate() (err error) {
	u := &User{}
	err = entity.AutoMigrate(db)
	if err != nil {
		return err
	}
	// PGSQL
	// err = db.AutoMigrate(u)
	err = db.AutoMigrate(u, &Email{})
	return err
}
