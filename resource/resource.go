package resource

import (
	"reflect"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"

	"github.com/Sky-And-Hammer/TM_EC"
	"github.com/Sky-And-Hammer/TM_EC/utils"
	"github.com/Sky-And-Hammer/roles"
)

//	'Resourcer' interface
type Resourcer interface {
	GetResource() *Resource
	GetMetas([]string) []Metaor
	CallFindMany(interface{}, *TM_EC.Context) error
	CallFindOne(interface{}, *MetaValues, *TM_EC.Context) error
	CallSave(interface{}, *TM_EC.Context) error
	CallDelete(interface{}, *TM_EC.Context) error
	NewSlice() interface{}
	NewStruct() interface{}
}

//	'ConfigureReouseceBeforeInitializeInterface' if a struct implemented this interface, it will be called before everything when create a resource with the struct
type ConfigureResourceBeforeInitializeInterface interface {
	ConfigureECResourceBeforeInitialize(Resourcer)
}

//	'ConfigureResourceInterface' if a struct implemented this interface, it will be called after configured by user
type ConfigureResourceInterface interface {
	ConfigureECResource(Resourcer)
}

//	'Resource' is a struct that including basic definition of EC resource
type Resource struct {
	Name            string
	Value           interface{}
	FindManyHandler func(interface{}, *TM_EC.Context) error
	FindOneHandler  func(interface{}, *MetaValues, *TM_EC.Context) error
	SaveHandler     func(interface{}, *TM_EC.Context) error
	DeleteHandler   func(interface{}, *TM_EC.Context) error
	Permission      *roles.Permission
	Validatiors     []func(interface{}, *MetaValues, *TM_EC.Context) error
	Processors      []func(interface{}, *MetaValues, *TM_EC.Context) error
	primaryField    *gorm.Field
}

//	'New' initialize EC resource
func New(value interface{}) *Resource {
	var (
		name = utils.HumanizeString(utils.ModelType(value).Name())
		res  = &Resource{
			Value: value,
			Name:  name,
		}
	)

	res.FindOneHandler = res.findOneHandler
	res.FindManyHandler = res.findManyHandler
	res.SaveHandler = res.saveHandler
	res.DeleteHandler = res.deleteHandler
	return res
}

//	'GetResource' return isself to match interface 'Resourcer'
func (res *Resource) GetResource() *Resource {
	return res
}

//	'AddValidator' add validator to resource, it will invoked when creating, updating, and will rollback the change if validator return any error
func (res *Resource) AddValidator(fc func(interface{}, *MetaValues, *TM_EC.Context) error) {
	res.Validatiors = append(res.Validatiors, fc)
}

//	'addProcessor' add processor to resource, it is used to process data before creating, updating, will rollback the change if it return any error
func (res *Resource) addProcessor(fc func(interface{}, *MetaValues, *TM_EC.Context) error) {
	res.Processors = append(res.Processors, fc)
}

//	'NewStruct' initialize a struct for the resource
func (res *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(res.Value)).Type()).Interface()
}

//	'NewSlice' initializ a slice of struct for the resource
func (res *Resource) NewSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.TypeOf(res.Value))
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}

//	'GetMetas' get defined metas, to match interface "Resourcer"
func (res *Resource) GetMetas([]string) []Metaor {
	panic("not defined")
}

//	'HasPermission' check permission of resource
func (res *Resource) HasPermission(mode roles.PermissionMode, context *TM_EC.Context) bool {
	if res == nil || res.Permission == nil {
		return true
	}

	return res.Permission.HasPermission(mode, context.Roles...)
}

//	'PrimaryField' return gorm's primary field
func (res *Resource) PrimaryField() *gorm.Field {
	if res.primaryField == nil {
		scope := gorm.Scope{
			Value: res.Value,
		}
		res.primaryField = scope.PrimaryField()
	}

	return res.primaryField
}

//	'PrimaryDBName' return db column name of the resource's primary field
func (res *Resource) PrimaryDBName() (name string) {
	filed := res.primaryField
	if filed != nil {
		name = filed.Name
	}
	return
}

//	'PrimaryFieldName' return struct column name of the resource's primary field
func (res *Resource) PrimaryFieldName() (name string) {
	field := res.primaryField
	if field != nil {
		name = field.Name
	}
	return
}
