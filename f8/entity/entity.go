package entity

import (
	"errors"
	"log"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/phanirithvij/fate/f8/buckets"
	"gorm.io/gorm"
)

// BaseEntity a base entity model
type BaseEntity struct {
	// ID entity id not auto generated
	ID string `gorm:"unique;primaryKey;not null"`
	// https://gorm.io/docs/models.html#Field-Level-Permission
	// would be great if this was a map but unfortunately sql no maps
	// Buckets []*buckets.Bucket `gorm:"foreignKey:EntityID"`
	Buckets []*buckets.Bucket `gorm:"polymorphic:Entity"`
	// Buckets []*buckets.Bucket `gorm:"polymorphic:Entity;<-:false"`
	// f8.BaseEntity `gorm:"-"`
	db                *gorm.DB `gorm:"-"`
	entityType        string   `gorm:"-"`
	defaultBucketName string   `gorm:"-"`
}

var (
	// EntityBucketMap a map of entity and bucket values
	//	eg:
	//	map["userid"]["default"]
	EntityBucketMap map[string]map[string]*buckets.Bucket = make(map[string]map[string]*buckets.Bucket)
)

// From this vararg approach
// https://github.com/faiface/gui/commit/ee3366ded862f02a1a5ee4ea856a06e46bb889ee#diff-f9c3d4c5cce2eabfcf19a1c38214e739a6619bdd22b15373e34be2e3e4589247R89

// Option is a functional option to the entity constructor New.
type Option func(*options)
type options struct {
	id                string
	numBuckets        int
	defaultBucketName string
	bucketNames       []string
	tableName         string
	db                *gorm.DB
}

// ID option sets the ID of the entity.
//
// Don't set it for an auto uuid
//
// If it's username better set it or you'll get uuids as usernames
func ID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// BucketCount option sets the num of buckets initially
func BucketCount(numBuckets int) Option {
	return func(o *options) {
		o.numBuckets = numBuckets
	}
}

// TableName option sets the name of the entity
// REQUIRED
func TableName(tableName string) Option {
	return func(o *options) {
		o.tableName = tableName
	}
}

// DB the gorm db instance for querying the buckets database
func DB(db *gorm.DB) Option {
	return func(o *options) {
		o.db = db
	}
}

// BucketName option sets the default bucket Name for the entity
//
// Must specify this or BucketNames if using BucketCount() as buckets will be created
// as name-0, name-1, name-2, ..., name-{count-1}
// default is `default`
func BucketName(bucketName string) Option {
	return func(o *options) {
		o.defaultBucketName = bucketName
	}
}

// BucketNames option sets the bucketNames for the entity
//
// Must specify these or BucketName if using BucketCount() as buckets will be created
// as names[0], names[1], names[2], ..., names[count-1]
func BucketNames(bucketNames []string) Option {
	return func(o *options) {
		o.bucketNames = bucketNames
	}
}

// Entity a new base entity
func Entity(opts ...Option) (*BaseEntity, error) {
	o := options{
		defaultBucketName: "default",
	}
	for _, opt := range opts {
		opt(&o)
	}
	if o.db == nil {
		return nil, errors.New("Must pass the gorm database instance")
	}
	if o.id == "" {
		o.id = uuid.New().String()
	}
	if o.tableName == "" {
		return nil, errors.New("Must specify the table name")
	}

	if o.defaultBucketName == "" {
		o.defaultBucketName = "default"
	}

	// whether we should use bucketNames[]
	usebNames := false
	if len(o.bucketNames) > 0 {
		if o.defaultBucketName != "default" {
			// both bucketNames and a name was specifed for an incremental bucket name
			return nil, errors.New("Use only one of BucketNames, BucketName")
		}
		if len(o.bucketNames) != o.numBuckets {
			return nil, errors.New("Number of bucket names must match the numBuckets")
		}
		usebNames = true
	}

	ent := &BaseEntity{
		ID:         o.id,
		Buckets:    []*buckets.Bucket{},
		entityType: o.tableName,
		db:         o.db,
	}
	if o.numBuckets == 0 {
		// number of buckets was not specified
		// create a default bucket
		o.numBuckets = 1
	}

	ent.defaultBucketName = o.defaultBucketName

	// populate the buckets from the db
	_ = ent.FetchBuckets()
	// if len(existing) > 0 {
	// 	// some buckets already exist
	// }
	// create initial buckets
	for i := 0; i < o.numBuckets; i++ {
		bID := o.defaultBucketName
		if usebNames {
			bID = o.bucketNames[i]
		} else {
			if i != 0 {
				// for i = 0 it will be default but not default-0
				bID = o.defaultBucketName + "-" + strconv.Itoa(i)
			}
		}
		// this will create a bucket but will no-op if already exists
		// so no need for error handling
		_, err := ent.CreateBucket(bID)
		if err != nil {
			log.Println(err)
		}
	}

	return ent, nil
}

// FetchBuckets fetches existing buckets and adds them to the map and list
func (e *BaseEntity) FetchBuckets() (bucks []*buckets.Bucket) {
	log.Println("FetchBuckets(...)")
	bucks = e.GetBuckets()
	e.Buckets = append(e.Buckets, bucks...)

	// populate map
	if _, ok := EntityBucketMap[e.ID]; !ok {
		EntityBucketMap[e.ID] = make(map[string]*buckets.Bucket)
	}
	for _, b := range bucks {
		b.AttatchDB(e.db)
		if val, ok := EntityBucketMap[e.ID][b.ID]; !ok {
			log.Println("Added fetched", b.ID, "to map")
			EntityBucketMap[e.ID][b.ID] = b
		} else {
			// already exists which is unlikely
			log.Println("Duplicate bucket exists", val, e.ID, b.ID)
		}
	}
	log.Println(bucks)
	return bucks
}

// GetBuckets fetches existing buckets and adds them to the map and list
func (e *BaseEntity) GetBuckets() (bucks []*buckets.Bucket) {
	tx := e.db.Where(
		"entity_id = ? AND entity_type = ?",
		e.ID, e.entityType,
	).Find(&bucks)
	if tx.Error != nil {
		log.Println("Failed to fetch buckets", tx.Error)
		return nil
	}
	return bucks
}

// OverwriteBuckets overwrites the local buckets from the database
func (e *BaseEntity) OverwriteBuckets() {
	e.Buckets = e.GetBuckets()
	// reset and populate map
	EntityBucketMap[e.ID] = make(map[string]*buckets.Bucket)
	for _, b := range e.Buckets {
		b.AttatchDB(e.db)
		if _, ok := EntityBucketMap[e.ID][b.ID]; !ok {
			EntityBucketMap[e.ID][b.ID] = b
		}
	}
	// for mapID := range EntityBucketMap[e.ID] {
	// 	// if updated bucket id is not found in the map
	// 	inList := true
	// 	for _, b := range e.Buckets {
	// 		if _, ok := EntityBucketMap[e.ID][b.ID]; ok {
	// 			// new bucket already exists in map no need to loop
	// 			inList = true
	// 			break
	// 		}
	// 		// overwrite don't care if it exists or not
	// 		EntityBucketMap[e.ID][b.ID] = b
	// 		// if mapID is b.ID them mapID exists in
	// 		inList = mapID == b.ID
	// 	}
	// 	// not found in buckets list so remove it from the map
	// 	if !inList {
	// 		delete(EntityBucketMap[e.ID], mapID)
	// 	}
	// }
}

// CreateBucket creates a new bucket for the entity
// and appends it to the entity owned bucket list
func (e *BaseEntity) CreateBucket(bID string) (buck *buckets.Bucket, err error) {
	if _, ok := EntityBucketMap[e.ID]; !ok {
		EntityBucketMap[e.ID] = make(map[string]*buckets.Bucket)
	}
	if _, ok := EntityBucketMap[e.ID][bID]; !ok {
		buck = buckets.NewBucket(bID, e.db)
		EntityBucketMap[e.ID][bID] = buck
		e.Buckets = append(e.Buckets, buck)
		log.Println("Added", buck.ID, "to map")
		return buck, nil
	}
	return nil, errors.New("Bucket already exists " + bID)
}

// GetBucket returns a bucket with given name reading from DB
//
// pass an empty string to get the default bucket
func (e *BaseEntity) GetBucket(bID string) (buck *buckets.Bucket, err error) {
	if e.db == nil {
		return nil, errors.New("DB was nil for some reason")
	}
	if bID == "" {
		bID = e.defaultBucketName
	}
	buck = &buckets.Bucket{
		ID:         bID,
		EntityID:   e.ID,
		EntityType: e.entityType,
	}
	buck.AttatchDB(e.db)

	tx := e.db.First(buck)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if _, ok := EntityBucketMap[e.ID]; !ok {
		EntityBucketMap[e.ID] = make(map[string]*buckets.Bucket)
	}
	// Add to map
	if _, ok := EntityBucketMap[e.ID][bID]; !ok {
		EntityBucketMap[e.ID][bID] = buck
	}
	// TODO add to list if it doesn't exist
	return buck, nil
}

// DeleteBucket deletes a bucket from the entity
func (e *BaseEntity) DeleteBucket(bID string) bool {
	if _, ok := EntityBucketMap[e.ID]; !ok {
		EntityBucketMap[e.ID] = make(map[string]*buckets.Bucket)
	}

	var buck *buckets.Bucket
	// get bucket from map
	if val, ok := EntityBucketMap[e.ID][bID]; ok && val != nil {
		buck = val
		log.Println("buck is if", val)
	} else {
		var err error
		buck, err = e.GetBucket(bID)
		log.Println("buck is else", buck)
		if err != nil {
			log.Println("Bucket doesn't exist")
			return false
		}
	}
	log.Println("buck is after if", buck)
	// try delete
	ok := buck.Delete()

	if ok {
		// find in list
		idx := sort.Search(len(e.Buckets), func(i int) bool {
			return e.Buckets[i].ID == buck.ID
		})
		// delete from list
		e.Buckets = remove(e.Buckets, idx)
		// delete from map
		delete(EntityBucketMap[e.ID], bID)
	}

	return ok
}

// https://stackoverflow.com/a/37335777/8608146
// TODO: Bug fix
// This function fails sometimes with list index out of bounds
// Very randomly need some sync locks
func remove(s []*buckets.Bucket, i int) []*buckets.Bucket {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
}

// AutoMigrate auto migrations required for the database
//
// Note: EntityBase will not auto migrate because it's the parent's responsibility
func AutoMigrate(db *gorm.DB) error {
	return buckets.AutoMigrate(db)
}
