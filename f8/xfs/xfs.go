package xfs

import (
	"os"
	"time"

	"gorm.io/gorm"
)

// XFileSystem a simple filesystem implementation
type XFileSystem struct {
	FileDirs []FileDir `gorm:"foreignKey:BucketID"`
}

// Bucket is equivalient to a filesystem with a name
type Bucket struct {
	gorm.Model
	XFileSystem
	/*
		Bucket ID will not be unique as evey entity has a `default` bucket
		Entity ID and bucket ID together will be unique
	*/
	// https://gorm.io/docs/composite_primary_key.html
	ID       string `gorm:"unique_index:compositeindex;primarykey"`
	EntityID string `gorm:"unique_index:compositeindex"`
}

// copied from os.FileInfo interface{}

// FileDir a file or directory
type FileDir struct {
	gorm.Model
	Name     string      `gorm:"primarykey"` // base name of the file
	Size     int64       // length in bytes for regular files; system-dependent for others
	Mode     os.FileMode // file mode bits
	ModTime  time.Time   // modification time
	IsDir    bool        // abbreviation for Mode.IsDir
	BucketID string      `gorm:"primarykey"`
}

// AutoMigrate for xfs
func AutoMigrate(db *gorm.DB) error {
	return db.Debug().AutoMigrate(&FileDir{})
}
