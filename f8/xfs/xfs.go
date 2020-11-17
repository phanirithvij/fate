package xfs

import (
	"os"
	"time"
)

// XFileSystem a simple filesystem implementation
type XFileSystem struct {
	FileDir
}

// Bucket is equivalient to a filesystem with a name
type Bucket struct {
	XFileSystem
	/*
		Bucket ID will not be unique as evey entity has a `default` bucket
		Entity ID and bucket ID together will be unique
	*/
	// https://gorm.io/docs/composite_primary_key.html
	ID       string `gorm:"primarykey"`
	EntityID string `gorm:"primarykey"`
}

// copied from os.FileInfo interface{}

// FileDir a file or directory
type FileDir struct {
	Name    string      // base name of the file
	Size    int64       // length in bytes for regular files; system-dependent for others
	Mode    os.FileMode // file mode bits
	ModTime time.Time   // modification time
	IsDir   bool        // abbreviation for Mode.IsDir
}
