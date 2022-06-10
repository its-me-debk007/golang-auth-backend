package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"primarykey"`
	Password string `json:"password"`
	IsVerified bool `json:"is_verified"`
}
