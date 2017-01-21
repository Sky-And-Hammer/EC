package utils

import (
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"
	//	A pure Go postgres driver for Go's database/sql package
	"github.com/lib/pq"
)

func TestDB() *gorm.DB {
	var db *gorm.DB
	var err error
	var dbuser, dbpwd, dbname = "root", "19911220", "ec_test"
	if os.Getenv("DB_USER") != "" {
		dbuser = os.Getenv("DB_USER")
	}

	if os.Getenv("DB_PWD") != "" {
		dbpwd = os.Getenv("DB_PWD")
	}

	if os.Getenv("TEST_DB") == "postgres" {
		db, err := gorm.Open("prostgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", dbuser, dbpwd, dbname))
	} else {
		db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	}

	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		db.LogMode(true)
	}

	return db
}
