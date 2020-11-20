package f8

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// We need postgres driver
	"github.com/lib/pq"
	"github.com/phanirithvij/fate/f8/browser"
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
	LitePath string
	// DatabaseMode the buckets will use this to store their data
	DatabaseMode DBKind
	GormConfig   *gorm.Config
}

// Option is a functional option to the entity constructor New.
type Option func(*options)
type options struct {
	appName    string
	db         *gorm.DB
	storageDir string
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

// SetDBConfig the configuration configuration for the database
func SetDBConfig(conf *DBConfig) Option {
	return func(o *options) {
		o.dbConfig = conf
	}
}

// StorageDir the storage directory
//
// by default it is
func StorageDir(dir string) Option {
	return func(o *options) {
		o.storageDir = dir
	}
}

// AppName the application's name
//
// Change this if the default application storage path by default it is `f8`
func AppName(name string) Option {
	return func(o *options) {
		o.appName = name
	}
}

// InitDB will initialize the grom database
func (s *StorageConfig) InitDB(existing *gorm.DB) *gorm.DB {
	if existing != nil {
		return existing
	}
	log.Println("Initializing grom database ...")
	conf := s.DBConfig
	if conf == nil {
		log.Println("Warning: Config was null")
		return nil
	}
	switch conf.DatabaseMode {
	case PgSQL:
		s.DB = conf.PostGreSQLDB()
		return s.DB
	case Sqlite:
		s.DB = conf.SqliteDB()
		return s.DB

	default:
		log.Println("Unknown database requested", conf.DatabaseMode)
		return nil
	}

}

// SqliteDB an sqlite database
func (conf *DBConfig) SqliteDB() *gorm.DB {
	if conf.DatabaseMode != Sqlite {
		log.Println("Warning: not a sqlite database")
		return nil
	}
	var err error
	db, err := gorm.Open(sqlite.Open(conf.LitePath), conf.GormConfig)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// PostGreSQLDB retuns a grom database from the config
func (conf *DBConfig) PostGreSQLDB() *gorm.DB {
	if conf.DatabaseMode != PgSQL {
		log.Println("Warning: not a postgres database")
		return nil
	}
	var err error
	if conf.PGdbname == "" {
		conf.PGdbname = "f8"
	}
	// https://stackoverflow.com/a/55556599/8608146
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
	_, err = dbx.Exec("create database " + conf.PGdbname)
	// no need of the sql connection dispose it
	// no need to defer because we can close it immediately
	dbx.Close()
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// https://pkg.go.dev/github.com/lib/pq#hdr-Errors
			if err.Code.Name() == "duplicate_database" {
				log.Println("[f8][Info]: Database f8 Already exists")
			}
		} else {
			// not a pq error so assume fail
			log.Fatal(err)
		}
	}

	log.Println("Created database", conf.PGdbname, "successfully")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.PGusername,
		conf.PGpassword,
		conf.PGhostname,
		conf.PGport,
		conf.PGdbname,
	)
	db, err := gorm.Open(postgres.Open(dsn), conf.GormConfig)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// New retuns a new f8 storage object
func New(opts ...Option) (s *StorageConfig, err error) {
	o := options{
		appName: "f8",
		dbConfig: &DBConfig{
			LitePath:     "f8.db",
			DatabaseMode: Sqlite,
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.storageDir == "" {
		// Default storage directory
		// log.Println(os.UserConfigDir())
		appConfg := configdir.New(o.appName, "f8-storage")
		o.storageDir = appConfg.QueryFolders(configdir.Global)[0].Path
	} else {
		// storage directory was specified
		if !filepath.IsAbs(o.storageDir) {
			log.Println(relativeWarn)
			o.storageDir, err = filepath.Abs(o.storageDir)
			if err != nil {
				log.Println("Failed to get the absolute path of the storage directory")
				return nil, err
			}
		}
	}

	err = os.MkdirAll(o.storageDir, 0766)
	if err != nil {
		log.Println("Failed to create the storage directory")
		return nil, err
	}
	log.Println("The storage directory is", o.storageDir)

	s = &StorageConfig{
		StorageDir: o.storageDir,
		DBConfig:   o.dbConfig,
	}

	s.InitDB(s.DB)
	return s, err
}

// StartBrowser starts a filebrowser instance
func (s *StorageConfig) StartBrowser() {
	browser.StartBrowser(s.StorageDir)
}
