package db

import (
	"gmwe/api/auth"
	"gmwe/api/db"
	"gmwe/api/hitokoto"
)

func InitDB() error {
	return db.DB().Engine.AutoMigrate(&auth.User{}, &hitokoto.Hitokoto{})
}
