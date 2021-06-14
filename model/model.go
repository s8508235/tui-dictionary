package model

import (
	"gorm.io/gorm"
)

type Dictionary struct {
	gorm.Model
	Word       string
	Definition []byte
}
