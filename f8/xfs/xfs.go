package xfs

import "os"

// XFileSystem a simple filesystem implementation
type XFileSystem struct {
}

// Bucket is equivalient to a filesystem with a name
type Bucket struct {
	XFileSystem
	Name string
}

// File a file
type File struct {
	os.FileInfo
}

// Dir a directory
type Dir struct {
	os.FileInfo
}
