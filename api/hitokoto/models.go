package hitokoto

import (
	"gmwe/api/auth"

	"gorm.io/gorm"
)

type Hitokoto struct {
	gorm.Model
	Content           string `gorm:"unique;type:text COLLATE gmwe_unicode_ci"`
	UserID            int
	User              auth.User
	SubmittedByUserID *int `json:",omitempty"`
}
