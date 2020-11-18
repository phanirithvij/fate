package main

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User ...
type User struct {
	gorm.Model
	Buckets []*Bucket `gorm:"polymorphic:Entity"`
}

// Bucket ...
type Bucket struct {
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Name       string         `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
	EntityID   string         `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
	EntityType string         `gorm:"uniqueIndex:buk_ent_idx;"`
	FileDirs   []FileDir      `gorm:"polymorphic:Bucket;"`
}

// FileDir ...
type FileDir struct {
	gorm.Model
	Name       string `gorm:"primarykey"`
	BucketID   string `gorm:"primarykey"`
	BucketType string
}

func main() {
	var err error
	db, err := gorm.Open(sqlite.Open("sample.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug()
	db.AutoMigrate(&User{}, &Bucket{}, &FileDir{})
}
