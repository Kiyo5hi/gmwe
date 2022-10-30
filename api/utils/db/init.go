package db

import (
	"acey.k1yoshi.com/gmwe/api/auth"
	"acey.k1yoshi.com/gmwe/api/db"
	"acey.k1yoshi.com/gmwe/api/hitokoto"
)

func InitDB() (err error) {
	err = db.DB().Engine.AutoMigrate(&auth.User{}, &hitokoto.Hitokoto{})
	if err != nil {
		return err
	}
	return fillData()
}

func fillData() (err error) {
	userService := new(auth.UserService)
	_, err = userService.FindUserByUsername("acey")
	if err != nil {
		acey := auth.User{
			Username: "acey",
			Name:     "Acey Tong",
			Password: "LoveFromKiyoshi",
		}
		err = userService.CreateUser(&acey)
		if err != nil {
			return err
		}
	}

	_, err = userService.FindUserByUsername("kiyoshi")
	if err != nil {
		kiyoshi := auth.User{
			Username: "kiyoshi",
			Name:     "Kiyoshi Guo",
			Password: "LoveFromAcey",
		}
		err = userService.CreateUser(&kiyoshi)
		if err != nil {
			return err
		}
	}

	_, err = userService.FindUserByUsername("aron")
	if err != nil {
		aron := auth.User{
			Username: "aron",
			Name:     "Aron",
			Password: "WeLoveAron",
		}
		err = userService.CreateUser(&aron)
		if err != nil {
			return err
		}
	}

	return nil
}
