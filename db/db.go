package db

import (
	"errors"
	"fmt"
	"os"

	//	The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"

	"github.com/Sky-And-Hammer/EC/config"
)

var (
	DB *gorm.DB
)

func init() {
	var err error
	dbConfig := 
}
