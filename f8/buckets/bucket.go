package buckets

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// XFileSystem a simple filesystem implementation
type XFileSystem struct {
	FileDirs []FileDir `gorm:"polymorphic:Bucket"`
	// FileDirs []FileDir `gorm:"foreignKey:BucketID"`
}

// copied from os.FileInfo interface{}

// FileDir a file or directory
type FileDir struct {
	gorm.Model
	// TODO: Once a file or dir is created it is our job to populate these fields
	Name       string      // base name of the file
	Path       string      `gorm:"primarykey"` // absolute path of the file
	Size       int64       // length in bytes for regular files; system-dependent for others
	Mode       os.FileMode // file mode bits
	ModTime    time.Time   // modification time
	IsDir      bool        // abbreviation for Mode.IsDir
	BucketID   string      `gorm:"primarykey"`
	BucketType string
	*os.File   `gorm:"-"`
}

// Bucket is equivalient to a filesystem with a name
type Bucket struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	XFileSystem `gorm:"embedded"`
	/*
		Bucket ID will not be unique as evey entity has a `default` bucket
		Entity ID and bucket ID together will be unique
	*/
	// https://stackoverflow.com/a/63409572/8608146
	// https://gorm.io/docs/composite_primary_key.html

	// Must be named as ID because
	// any othername fails polymorphic clause on XFileSystem

	// The Name of the bucket or the bucket name
	// Name       string `gorm:"uniqueIndex:buk_ent_idx;unique;primaryKey"`
	ID         string   `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
	EntityID   string   `gorm:"uniqueIndex:buk_ent_idx;primaryKey"`
	EntityType string   `gorm:"primaryKey"`
	db         *gorm.DB `gorm:"-" json:"-"`
	// Deleted to keep track of deleted buckets
	Deleted bool `gorm:"-"`
}

// newBucket returns a new bucket, if id is empty ID is default
func newBucket(id string) *Bucket {
	if id == "" {
		id = "default"
	}
	// Provision a bucket with an empty file system
	buck := &Bucket{
		ID:          id,
		XFileSystem: XFileSystem{FileDirs: []FileDir{}},
	}
	return buck
}

// NewBucket creates a new bucket for the given entity and attaches db
func NewBucket(bID string, db *gorm.DB) *Bucket {
	buck := newBucket(bID)
	buck.AttatchDB(db)
	return buck
}

// AttatchDB attaches the given db to the bucket
func (b *Bucket) AttatchDB(db *gorm.DB) {
	b.db = db
}

// Exists checks if the bucket already exists
func (b *Bucket) Exists() bool {
	if b == nil {
		log.Println("Bucket is nil handle error checking properly")
		return false
	}
	tx := b.db.First(b)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return false
		}
		log.Println("?")
		log.Fatal(tx.Error)
	}
	return true
}

// Delete deletes the bucket
//
// Use entity DeleteBucket instead
// This will delete the bucket but not from the entity map and list
func (b *Bucket) Delete() bool {
	if b == nil {
		log.Println("Bucket is nil handle error checking properly")
		return false
	}
	if b.db == nil {
		log.Println("Bucket DB is nil AttachDB call missed somewhere")
		return false
	}
	tx := b.db.Delete(b)
	// it doesn't exist at all
	if tx.RowsAffected == 0 {
		return false
	}
	if tx.Error != nil {
		log.Println(tx.Error)
		return false
	}
	log.Println("Deleted bucket", b.ID)
	b.Deleted = true
	return true
}

func (b *Bucket) String() string {
	xx, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		type bb *Bucket
		bx := (bb)(b)
		return fmt.Sprintf("%+v\n", bx)
	}
	return string(xx)
}

// GetDeletedBuckets returns the list of soft deleted buckets
//
// Use CleanupBuckets to permanently delete those
func GetDeletedBuckets(db *gorm.DB) (delbuckets []*Bucket) {
	// get deleted buckets
	tx := db.Unscoped().Where("deleted_at IS NOT NULL").Find(&delbuckets)
	if tx.Error != nil {
		log.Fatal(tx.Error)
	}
	return delbuckets
}

// CleanupBuckets cleans up the bucket table
//
// Better backup by getting them using the GetDeletedBuckets method
func CleanupBuckets(db *gorm.DB) bool {
	tx := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&Bucket{})
	if tx.Error != nil {
		log.Println(tx.Error)
		return false
	}
	return true
}

// NewFile retuns a new file
func NewFile(name string) (*FileDir, error) {
	if name == "" {
		return nil, errors.New("FileName was empty")
	}
	fdir := &FileDir{
		Name:  name,
		IsDir: false,
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return fdir, nil
}

// NewDir returns a new directory
func NewDir(name string) (*FileDir, error) {
	if name == "" {
		return nil, errors.New("Directory name was empty")
	}
	fdir := &FileDir{
		Name:  name,
		IsDir: true,
	}
	err := os.MkdirAll(name, 0766)
	if err != nil {
		return nil, err
	}
	return fdir, nil
}

// Move the file to a new location
func (f *FileDir) Move(dest string) (bool, error) {
	return false, errors.New("Not implemented")
}

// AutoMigrate for xfs
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Bucket{}, &FileDir{})
}

// BeforeCreate before creating fix the conflicts for primarykey
func (b *Bucket) BeforeCreate(tx *gorm.DB) (err error) {
	cols := []clause.Column{}
	colsNames := []string{}
	for _, field := range tx.Statement.Schema.PrimaryFields {
		cols = append(cols, clause.Column{Name: field.DBName})
		colsNames = append(colsNames, field.DBName)
	}
	// https://gorm.io/docs/create.html#Upsert-On-Conflict
	// https://github.com/go-gorm/gorm/issues/3611#issuecomment-729673788
	tx.Statement.AddClause(clause.OnConflict{
		Columns: cols,
		// DoUpdates: clause.AssignmentColumns(colsNames),
		DoNothing: true,
	})
	return nil
}

// BeforeUpdate before updating fix the conflicts for primarykey
func (b *Bucket) BeforeUpdate(tx *gorm.DB) (err error) {
	cols := []clause.Column{}
	colsNames := []string{}
	for _, field := range tx.Statement.Schema.PrimaryFields {
		cols = append(cols, clause.Column{Name: field.DBName})
		colsNames = append(colsNames, field.DBName)
	}
	colsNames = append(colsNames, "updated_at")
	// https://gorm.io/docs/create.html#Upsert-On-Conflict
	// https://github.com/go-gorm/gorm/issues/3611#issuecomment-729673788
	tx.Statement.AddClause(clause.OnConflict{
		Columns:   cols,
		DoUpdates: clause.AssignmentColumns(colsNames),
	})
	return nil
}
