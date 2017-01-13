package resource

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"

	"github.com/Sky-And-Hammer/TM_EC"
	"github.com/Sky-And-Hammer/TM_EC/utils"
	"github.com/Sky-And-Hammer/roles"
	"github.com/Sky-And-Hammer/validations"
)

//	'Metaor' interface
type Metaor interface {
	GetName() string
	GetFieldName() string
	GetSetter() func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context)
	GetFormattedValuer() func(interface{}, *TM_EC.Context) interface{}
	GetValuer() func(interface{}, *TM_EC.Context) interface{}
	GetResource() Resourcer
	GetMetas() []Metaor
	HasPermission(roles.PermissionMode, *TM_EC.Context) bool
}

//	'ConfigureMetaBeforeInitializeInterface' if a strust's field's type implemented this interface, it will be called when initializing a meta
type ConfigureMetaBeforeInitializeInterface interface {
	ConfigureECMetaBeforeInitialize(Metaor)
}

//	'ConfigureMetaInterface' if a struct's field's type implemented this interface, it will be called after configed
type ConfigureMetaInterface interface {
	ConfigureECMeta(Metaor)
}

//	'MetaConfigInterface' meta configuration interface
type MetaConfigInterface interface {
	ConfigureMetaInterface
}

//	'MetaConfig' base meta config struct
type MetaConfig struct{}

//	'ConfigureECMeta' implement the MetaConfigInterface
func (MetaConfig) ConfigureECMeta(Metaor) {}

//	'Meta' meta struct definition
type Meta struct {
	Name            string
	FieldName       string
	FieldStruct     *gorm.StructField
	Setter          func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context)
	Valuer          func(interface{}, *TM_EC.Context) interface{}
	FormattedValuer func(interface{}, *TM_EC.Context) interface{}
	Config          MetaConfigInterface
	Resource        Resourcer
	Permission      *roles.Permission
}

//	'GetBaseResource' get base resource from meta
func (meta Meta) GetBaseResource() Resourcer {
	return meta.Resource
}

//	'GetName' get meta's name
func (meta Meta) GetName() string {
	return meta.Name
}

//	'getFieldName' get meta's field name
func (meta Meta) getFieldName() string {
	return meta.FieldName
}

//	'SetFieldName' set meta's field name
func (meta *Meta) SetFieldName(name string) {
	meta.FieldName = name
}

//	'GetSetter' get meta's setter
func (meta Meta) GetSetter() func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context) {
	return meta.Setter
}

//	'SetSetter' set meta's setter
func (meta *Meta) SetSetter(fc func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context)) {
	meta.Setter = fc
}

//	'GetValuer' get meta's valuer
func (meta Meta) GetValuer() func(interface{}, *TM_EC.Context) interface{} {
	return meta.Valuer
}

//	'SetValuer' set meta's valuer
func (meta *Meta) SetValuer(fc func(interface{}, *TM_EC.Context) interface{}) {
	meta.Valuer = fc
}

//	'GetFormattedValuer' get formatted valuer form meta
func (meta *Meta) GetFormattedValuer() func(interface{}, *TM_EC.Context) interface{} {
	if meta.FormattedValuer != nil {
		return meta.FormattedValuer
	}
	return meta.Valuer
}

//	'SetFormattedValuer' set formatted valuer form meta
func (meta *Meta) SetFormattedValuer(fc func(interface{}, *TM_EC.Context) interface{}) {
	meta.FormattedValuer = fc
}

//	'HasPermission' check has permission or not
func (meta Meta) HasPermission(mode roles.PermissionMode, context *TM_EC.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context.Roles...)
}

//	'SetPermission' set permission for meta
func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
}

//	'PerInitialize' when will be run beform initialize, used to fill some basic necessary information
func (meta *Meta) PerInitialize() error {
	if meta.Name == "" {
		utils.ExistWithMsg("Meta should have name: %v", reflect.TypeOf(meta))
	} else if meta.FieldName == "" {
		meta.FieldName = meta.Name
	}

	var parseNestedField = func(value reflect.Value, name string) (reflect.Value, string) {
		fields := strings.Split(name, ".")
		value = reflect.Indirect(value)
		for _, field := range fields[:len(fields)-1] {
			value = value.FieldByName(field)
		}
		return value, fields[len(fields)-1]
	}

	var getField = func(fields []*gorm.StructField, name string) *gorm.StructField {
		for _, field := range fields {
			if field.Name == name || field.DBName == name {
				return field
			}
		}
		return nil
	}

	var nestedField = strings.Contains(meta.FieldName, ".")
	var scope = &gorm.Scope{
		Value: meta.Resource.GetResource().Value,
	}
	if nestedField {
		subModel, name := parseNestedField(reflect.ValueOf(meta.Resource.GetResource().Validatiors), meta.FieldName)
		meta.FieldStruct = getField(scope.New(subModel.Interface()).GetStructFields(), name)
	} else {
		meta.FieldStruct = getField(scope.GetStructFields(), meta.FieldName)
	}
	return nil
}

//	'Initialize' initialize meta, will set valuer, setter if haven't configure it
func (meta *Meta) Initialize() error {
	var (
		nestedField = strings.Contains(meta.FieldName, ".")
		field       = meta.FieldStruct
		hasColumn   = meta.FieldStruct != nil
	)

	var fieldType reflect.Type
	if hasColumn {
		fieldType = field.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	if meta.Valuer == nil {
		if hasColumn {
			meta.Valuer = func(value interface{}, context *TM_EC.Context) interface{} {
				scope := context.GetDB().NewScope(value)
				fieldName := meta.FieldName
				if nestedField {
					fileds := strings.Split(fieldName, ".")
					fieldName = fileds[len(fileds)-1]
				}

				if f, ok := scope.FieldByName(fieldName); ok {
					if f.Relationship != nil && f.Field.CanAddr() && !scope.PrimaryKeyZero() {
						context.GetDB().Model(value).Related(f.Field.Addr().Interface(), meta.FieldName)
					}

					return f.Field.Interface()
				}

				return ""
			}
		} else {
			utils.ExistWithMsg("Meta %v is not supported for resource %v, no 'Valuer' configured for it", meta.FieldName, reflect.TypeOf(meta.Resource.GetResource().Value))
		}
	}

	if meta.Setter == nil && hasColumn {
		if relationship := field.Relationship; relationship != nil {
			if relationship.Kind == "belongs_to" || relationship.Kind == "many_to_many" {
				meta.Setter = func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context) {
					scope := &gorm.Scope{
						Value: resource,
					}
					reflectValue := reflect.Indirect(reflect.ValueOf(resource))
					field := reflectValue.FieldByName(meta.FieldName)
					if field.Kind() == reflect.Ptr {
						if field.IsNil() {
							field.Set(utils.NewValue(field.Type()).Elem())
						}

						for field.Kind() == reflect.Ptr {
							field = field.Elem()
						}
					}

					primaryKeys := utils.ToArray(metaValue.Value)
					if relationship.Kind == "belong_to" && len(relationship.ForeignFieldNames) == 1 {
						oldPrimaryKeys := utils.ToArray(reflectValue.FieldByName(relationship.ForeignFieldNames[0]).Interface())
						if fmt.Sprint(primaryKeys) == fmt.Sprint(oldPrimaryKeys) {
							return
						}

						if len(primaryKeys) == 0 {
							field := reflectValue.FieldByName(relationship.ForeignFieldNames[0])
							field.Set(reflect.Zero(field.Type()))
						}
					}

					if len(primaryKeys) > 0 {
						context.GetDB().Where(primaryKeys).Find(field.Addr().Interface())
					}

					if relationship.Kind == "many_to_many" {
						if !scope.PrimaryKeyZero() {
							context.GetDB().Model(resource).Association(meta.FieldName).Replace(field.Interface())
							field.Send(reflect.Zero(field.Type()))
						}
					}
				}
			}
		} else {
			meta.Setter = func(resource interface{}, metaValue *MetaValue, context *TM_EC.Context) {
				if metaValue == nil {
					return
				}

				var (
					value     = metaValue.Value
					fieldName = meta.FieldName
				)

				defer func() {
					if r := recover(); r != nil {
						context.AddError(validations.NewError(resource, meta.Name, fmt.Sprintf("Can't set value %v", value)))
					}
				}()

				if nestedField {
					fields := strings.Split(fieldName, ".")
					fieldName = fields[len(fields)-1]
				}

				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(fieldName)
				if field.Kind() == reflect.Ptr {
					if field.IsNil() && utils.ToString(value) != "" {
						field.Set(utils.NewValue(field.Type()).Elem())
					}

					if utils.ToString(value) == "" {
						field.Set(reflect.Zero(field.Type()))
						return
					}

					for field.Kind() == reflect.Ptr {
						field = field.Elem()
					}
				}

				if field.IsValid() && field.CanAddr() {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(utils.ToInt(value))
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.SetUint(utils.ToUInt(value))
					case reflect.Float32, reflect.Float64:
						field.SetFloat(utils.ToFloat(value))
					case reflect.Bool:
						if utils.ToString(value) == "true" {
							field.SetBool(true)
						} else {
							field.SetBool(false)
						}
					default:
						if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
							if value == nil && len(metaValue.MetaValues.Values) > 0 {
								decodeMetaValuesToField(meta.Resource, field, metaValue, context)
								return
							}

							if scanner.Scan(value) != nil {
								scanner.Scan(utils.ToString(value))
							}
						} else if reflect.TypeOf("").ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(utils.ToString(value)).Convert(field.Type()))
						} else if reflect.TypeOf([]string{}).ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(utils.ToArray(value)).Convert(field.Type()))
						} else if rValue := reflect.ValueOf(value); reflect.TypeOf(rValue.Type()).ConvertibleTo(field.Type()) {
							field.Set(rValue.Convert(field.Type()))
						} else if _, ok := field.Addr().Interface().(*time.Time); ok {
							if str := utils.ToString(value); str != "" {
								if newTime, err := utils.ParseTime(str, context); err == nil {
									field.Set(reflect.ValueOf(newTime))
								}
							} else {
								field.Set(reflect.Zero(field.Type()))
							}
						} else {
							var buffer = bytes.NewBufferString("")
							json.NewEncoder(buffer).Encode(value)
							if err := json.NewDecoder(strings.NewReader(buffer.String())).Decode(field.Addr().Interface()); err != nil {
								utils.ExistWithMsg("Can't set value %v to %v [meta %v]", reflect.TypeOf(value), field.Type(), meta)
							}
						}
					}
				}
			}
		}
	}

	if nestedField {
		oldValue := meta.Valuer
		meta.Valuer = func(value interface{}, context *TM_EC.Context) interface{} {
			return oldValue(getNestedModel(value, meta.FieldName, context), context)
		}

		oldSetter := meta.Setter
		meta.Setter = func(resource interface{}, metaValu *MetaValue, context *TM_EC.Context) {
			oldSetter(getNestedModel(resource, meta.FieldName, context), metaValu, context)
		}
	}

	return nil
}

func getNestedModel(value interface{}, fieldName string, context *TM_EC.Context) interface{} {
	model := reflect.Indirect(reflect.ValueOf(value))
	fields := strings.Split(fieldName, ".")
	for _, field := range fields[:len(fields)-1] {
		subModel := model.FieldByName(field)
		if key := subModel.FieldByName("Id"); !key.IsValid() || key.Uint() == 0 {
			if subModel.CanAddr() {
				context.GetDB().Model(model.Addr().Interface()).Related(subModel.Addr().Interface())
				model = subModel
			} else {
				break
			}
		} else {
			model = subModel
		}
	}

	if model.CanAddr() {
		return model.Addr().Interface()
	}

	return nil
}
