package auth

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Name     string
	Password string `json:"-"`
}
