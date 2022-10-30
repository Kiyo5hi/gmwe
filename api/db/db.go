package db

import (
	"sync"

	"acey.k1yoshi.com/gmwe/api/consts"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
		db, err := gorm.Open(mysql.Open(consts.DB_URI), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		dbInstance = &gmweDB{
			Engine: db,
		}
	})
	return dbInstance
}
