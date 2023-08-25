package auth

import (
	"fmt"

	"acey.k1yoshi.com/api/db"
	gmwe "acey.k1yoshi.com/api/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func (UserService) CreateUser(user *User) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	res := db.DB().Engine.Create(user)
	if res.Error != nil {
		err = gmwe.ResourceConflictException{
			Message: fmt.Sprintf("failed to create user; user with username=%s already exists", user.Username),
		}
	}
	return res.Error
}

func (UserService) FindUserByUsername(un string) (u *User, err gmwe.GMWEException) {
	us := User{}
	res := db.DB().Engine.Where("username = ?", &un).First(&us)
	if res.Error != nil {
		return nil, gmwe.NotFoundException{
			Message: fmt.Sprintf("failed to find user with username=%s", un),
		}
	}
	return &us, nil
}

func (UserService) FindUserByID(id int) (u *User, err gmwe.GMWEException) {
	us := User{}
	res := db.DB().Engine.Where("id = ?", &id).First(&us)
	if res.Error != nil {
		return nil, gmwe.NotFoundException{
			Message: fmt.Sprintf("failed to find user with id=%d", id),
		}
	}
	return &us, nil
}

func (UserService) GetUsers() (us *[]User, err gmwe.GMWEException) {
	u := []User{}
	res := db.DB().Engine.Find(&u)
	if res.Error != nil {
		return nil, gmwe.NotFoundException{
			Message: fmt.Sprintf("failed to get users; %s", res.Error.Error()),
		}
	}
	return &u, nil
}
