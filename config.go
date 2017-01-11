package TM_EC

import (
	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"
)

type Config struct {
	DB *gorm.DB
}
