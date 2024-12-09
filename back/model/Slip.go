package model

import "gorm.io/gorm"

type Slip struct {
    gorm.Model
    TransRef  string `gorm:"unique;not null"`
}

func (Slip) TableName() string {
	return "slip"
}
