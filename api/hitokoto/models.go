package hitokoto

import (
	"gmwe/api/auth"

	"gorm.io/gorm"
)

type Hitokoto struct {
	gorm.Model
	Content string `gorm:"unique"`
	UserID  int
	User    auth.User
}
