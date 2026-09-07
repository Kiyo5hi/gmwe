package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"gmwe/api/consts"

	sqlite3 "github.com/mattn/go-sqlite3"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	sql.Register("gmwe-sqlite", &sqlite3.SQLiteDriver{ConnectHook: func(c *sqlite3.SQLiteConn) error {
		// One collator per connection; preserve case/accent-insensitive application text.
		comparison := collate.New(language.Und, collate.Loose)
		return c.RegisterCollation("gmwe_unicode_ci", func(a, b string) int {
			return comparison.CompareString(strings.TrimRight(a, " "), strings.TrimRight(b, " "))
		})
	}})
}

func Open(uri string) (*gorm.DB, error) {
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_synchronous=FULL"
	engine, err := gorm.Open(sqlite.Dialector{DriverName: "gmwe-sqlite", DSN: uri}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	pool, err := engine.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	return engine, nil
}

type gmweDB struct {
	Engine *gorm.DB
}

var (
	dbInstance *gmweDB
	once       sync.Once
)

func DB() (_ *gmweDB) {
	once.Do(func() {
		var db *gorm.DB
		db, err := Open(consts.DB_URI)
		if err != nil {
			panic(err)
		}
		dbInstance = &gmweDB{
			Engine: db,
		}
	})
	return dbInstance
}
