package hitokoto

import (
	"fmt"
	"math/rand"

	"acey.k1yoshi.com/api/auth"
	"acey.k1yoshi.com/api/db"
	gmwe "acey.k1yoshi.com/api/errors"
)

type HitokotoService struct{}

func (HitokotoService) CreateHitokoto(h *Hitokoto) (err gmwe.GMWEException) {
	db := db.DB().Engine
	us := new(auth.UserService)
	u, erro := us.FindUserByID(h.UserID)
	if erro != nil {
		return gmwe.InvalidArgumentException{
			Message: erro.Error(),
		}
	}
	res := db.Create(h)
	if res.Error != nil {
		err = gmwe.ResourceConflictException{
			Message: fmt.Sprintf("failed to create hitokoto with %v; resource already exists", *h),
		}
		return err
	}
	h.User = *u
	return nil
}

func (HitokotoService) RandomHitokoto() (h *Hitokoto, err gmwe.GMWEException) {
	db := db.DB().Engine
	var count int64
	db.Model(&Hitokoto{}).Count(&count)
	if count == 0 {
		err = gmwe.NotFoundException{
			Message: "failed to find hitokoto; table is empty",
		}
		return nil, err
	}
	hs := Hitokoto{}
	res := db.Offset(rand.Intn(int(count))).Preload("User").First(&hs)
	if res.Error != nil {
		return nil, err
	}
	return &hs, nil
}
