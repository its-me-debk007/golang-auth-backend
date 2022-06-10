package models

import (
	"time"
)

type Otps struct {
	Email     string    `json:"email" gorm:"primarykey"`
	Otp       int       `json:"otp"`
	CreatedAt time.Time `json:"created_at"`
}
