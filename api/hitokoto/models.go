package hitokoto

import (
	"acey.k1yoshi.com/gmwe/api/auth"
	"gorm.io/gorm"
)

type Hitokoto struct {
	gorm.Model
	Content string `gorm:"unique"`
	UserID  int
	User    auth.User
}
