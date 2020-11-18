package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// const (
// 	username = "postgres"
// 	password = "522191"
// 	hostname = "localhost"
// 	port     = "5433"
// 	// port     = "5432"
// 	dbame = "gorm"
// )

// // XFileSystem a simple filesystem implementation
// type XFileSystem struct {
// 	FileDirs []FileDir `gorm:"polymorphic:Bucket"`
// 	// FileDirs []FileDir `gorm:"foreignKey:BucketID"`
// }

// // copied from os.FileInfo interface{}

// // FileDir a file or directory
// type FileDir struct {
// 	gorm.Model
// 	// TODO: Once a file or dir is created it is our job to populate these fields
// 	Name       string      `gorm:"primarykey"` // base name of the file
// 	Size       int64       // length in bytes for regular files; system-dependent for others
// 	Mode       os.FileMode // file mode bits
// 	ModTime    time.Time   // modification time
// 	IsDir      bool        // abbreviation for Mode.IsDir
// 	BucketID   string      `gorm:"primarykey"`
// 	BucketType string
// 	info       os.FileInfo `gorm:"-"`
// }

// // Bucket is equivalient to a filesystem with a name
// type Bucket struct {
// 	CreatedAt   time.Time
// 	UpdatedAt   time.Time
// 	DeletedAt   gorm.DeletedAt `gorm:"index"`
// 	XFileSystem `gorm:"embedded"`
// 	/*
// 		Bucket ID will not be unique as evey entity has a `default` bucket
// 		Entity ID and bucket ID together will be unique
// 	*/
// 	// https://stackoverflow.com/a/63409572/8608146
// 	// https://gorm.io/docs/composite_primary_key.html

// 	// Must be named as ID because
// 	// any othername fails polymorphic clause on XFileSystem

// 	// The Name of the bucket or the bucket name
// 	// Name       string `gorm:"uniqueIndex:buk_ent_idx;unique;primaryKey"`
// 	ID         string `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
// 	EntityID   string `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
// 	EntityType string `gorm:"primaryKey"`
// }

// // UserDDD ...
// type UserDDD struct {
// 	gorm.Model
// 	Buckets []*BucketDDD `gorm:"foreignKey:EntityID"`
// }

// // BucketDDD ...
// type BucketDDD struct {
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt gorm.DeletedAt `gorm:"index"`
// 	ID        string         `gorm:"unique;primaryKey"`
// 	EntityID  string         `gorm:"unique;primaryKey"`
// 	FileDirs  []FileDirDDD   `gorm:"foreignKey:BucketID"`
// }

// // FileDirDDD ...
// type FileDirDDD struct {
// 	gorm.Model
// 	Name     string `gorm:"primarykey"`
// 	BucketID string `gorm:"primarykey"`
// }

// func main() {

// 	dsn := fmt.Sprintf(
// 		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
// 		username, password, hostname, port, dbame,
// 	)

// 	var err error
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	// db, err := gorm.Open(sqlite.Open("sample.db"), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	db = db.Debug()
// 	db.AutoMigrate(&UserDDD{}, &BucketDDD{}, &FileDirDDD{})

// 	var buckets []Bucket
// 	tx := db.Find(&buckets)
// 	fmt.Println(tx.RowsAffected)
// 	fmt.Println(tx.Error)
// }
