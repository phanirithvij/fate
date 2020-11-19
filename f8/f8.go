package f8

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/shibukawa/configdir"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	relativeWarn = `
Storage directory was a relative path
which will result in data being at different places
based on the working directory of the executable`
)

// DBKind the type of the database
//
// Default Sqlite is used because postgres configuration cannot be guessed easily
type DBKind string

const (
	// PgSQL Postgresql will be used as the internal database
	PgSQL DBKind = "postgres"
	// Sqlite implies Sqlite will be used as the database
	Sqlite DBKind = "sqlite"
)

// StorageConfig configuration for the storage
type StorageConfig struct {
	StorageDir string
	// DatabaseMode the buckets will use this to store their data
	DatabaseMode DBKind
	// DB the underlying gorm database instance
	//
	// This instance can be used if needed externally
	DB       *gorm.DB
	DBConfig *DBConfig
}

// DBConfig the configuration for the database
type DBConfig struct {
	// Postgres username
	PGusername,
	// Postgres password
	PGpassword,
	// Postgres hostname default "localhost"
	PGhostname,
	// Postgres database name default "f8"
	PGdbname string
	// Postgres port default 5432
	PGport int
	// Sqlite database file path default f8.db
	LitePath   string
	GormConfig gorm.Config
}

// Option is a functional option to the entity constructor New.
type Option func(*options)
type options struct {
	db         *gorm.DB
	storageDir string
	dbKind     DBKind
	dbConfig   *DBConfig
}

// DB may optionally pass a gorm DB instance to the constructor
//
// It will be used directory
func DB(db *gorm.DB) Option {
	return func(o *options) {
		o.db = db
	}
}

// SetDBKind default is sqlite
//
// The dabase used will be f8.db
func SetDBKind(kind DBKind) Option {
	return func(o *options) {
		o.dbKind = kind
	}
}

// SetDBConfig the configuration configuration for the database
func SetDBConfig(conf *DBConfig) Option {
	return func(o *options) {
		o.dbConfig = conf
	}
}

// StorageDir the storage directory
//
// by default it is `storage` and it
func StorageDir(db *gorm.DB) Option {
	return func(o *options) {
		o.db = db
	}
}

// InitDB will initialize the grom database
func (s *StorageConfig) InitDB(existing *gorm.DB) *gorm.DB {
	if existing != nil {
		return existing
	}
	log.Println("Initializing grom database ...")
	conf := s.DBConfig
	log.Println("Config is", conf)
	switch s.DatabaseMode {
	case PgSQL:
		var err error
		if conf.PGdbname == "" {
			conf.PGdbname = "f8"
		}
		conninfo := fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable",
			conf.PGusername,
			conf.PGpassword,
			conf.PGhostname,
			conf.PGport,
		)
		dbx, err := sql.Open("postgres", conninfo)
		if err != nil {
			log.Println("Failed to connect to the postgres DB")
			log.Fatal(err)
		}

		// try to create the database for f8
		_, err = dbx.Exec("create database if not exists" + conf.PGdbname)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Created database", conf.PGdbname, "successfully")
		// no need of the sql connection dispose it
		dbx.Close()

		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			conf.PGusername,
			conf.PGpassword,
			conf.PGhostname,
			conf.PGport,
			conf.PGdbname,
		)
		db, err := gorm.Open(postgres.Open(dsn), &conf.GormConfig)
		if err != nil {
			log.Fatal(err)
		}
		s.DB = db
		return db
	case Sqlite:
		log.Println("Initializing sqlite database ...")
		var err error
		db, err := gorm.Open(sqlite.Open(s.DBConfig.LitePath), &conf.GormConfig)
		if err != nil {
			log.Fatal(err)
		}
		s.DB = db
		log.Println("DB", s.DB)
		return db

	default:
		log.Println("Unknown database requested", s.DatabaseMode)
		return nil
	}

}

// New retuns a new f8 storage object
func New(opts ...Option) (s *StorageConfig, err error) {
	o := options{
		storageDir: "storage",
		dbKind:     Sqlite,
		dbConfig: &DBConfig{
			LitePath: "f8.db",
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.storageDir == "" {
		return nil, errors.New("Storage directory is empty")
	}

	if !filepath.IsAbs(o.storageDir) {
		log.Println(relativeWarn)
		o.storageDir, err = filepath.Abs(o.storageDir)
		if err != nil {
			log.Println("Failed to get the absolute path of the storage directory")
			return nil, err
		}
	}

	log.Println(os.UserConfigDir())
	log.Println(configdir.New("f8", "f8-storage"))

	err = os.MkdirAll(o.storageDir, 0766)
	if err != nil {
		log.Println("Failed to create the storage directory")
		return nil, err
	}

	s = &StorageConfig{
		StorageDir:   o.storageDir,
		DBConfig:     o.dbConfig,
		DatabaseMode: o.dbKind,
	}

	s.InitDB(s.DB)
	return s, err
}
