package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	username = "postgres"
	password = "522191"
	hostname = "localhost"
	port     = "5433"
	// port     = "5432"
	dbame = "gorm"
)

// UserDDD ...
type UserDDD struct {
	gorm.Model
	Buckets []*BucketDDD `gorm:"foreignKey:EntityID"`
}

// BucketDDD ...
type BucketDDD struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	ID        string         `gorm:"unique;primaryKey"`
	EntityID  string         `gorm:"unique;primaryKey"`
	FileDirs  []FileDirDDD   `gorm:"foreignKey:BucketID"`
}

// FileDirDDD ...
type FileDirDDD struct {
	gorm.Model
	Name     string `gorm:"primarykey"`
	BucketID string `gorm:"primarykey"`
}

func main() {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, hostname, port, dbame,
	)

	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// db, err := gorm.Open(sqlite.Open("sample.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug()
	db.AutoMigrate(&UserDDD{}, &BucketDDD{}, &FileDirDDD{})
}
