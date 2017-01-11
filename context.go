package TM_EC

import (
	"net/http"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"
)

//	'CurrentUser' is an interface, which is used for ec admin to get current logged user
type CurrentUser interface {
	DisplayName() string
}

//	'Context' is ec context, which is used for many ec components, used to share infoation between them
type Context struct {
	Request     *http.Request
	Writer      *http.ResponseWriter
	ResourceID  string
	Config      *Config
	Roles       []string
	DB          *gorm.DB
	CurrentUser CurrentUser
	Errors
}

//	'Clone' clone current context
func (context *Context) Clone() *Context {
	var clone = *context
	return &clone
}

//	'GetDB' get db from current context
func (context *Context) GetDB() *gorm.DB {
	if context.DB != nil {
		return context.DB
	}

	return context.Config.DB
}

//	'SetDB' set db into current context
func (context *Context) SetDB(DB *gorm.DB) {
	context.DB = DB
}
