package auth

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;type:text COLLATE gmwe_unicode_ci"`
	Name     string
	Password string `json:"-"`
}
