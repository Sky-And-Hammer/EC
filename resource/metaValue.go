package resource

import (
	"reflect"

	"github.com/Sky-And-Hammer/TM_EC"
)

//	'MetaValues' a slice of MetaValue
type MetaValues struct {
	Values []*MetaValue
}

func (mvs MetaValues) Get(name string) *MetaValue {
	for _, value := range mvs.Values {
		if value.Name == name {
			return value
		}
	}

	return nil
}

//	'MetaValue' a struct used to hold inforamtion when convert inputs from HTTP form, JSON, CSV fields and so on to meta values
//	It will includes file name, field value and it's configured Meta, if it is a nested resource, will includeds nested metas in it's MetaValues
type MetaValue struct {
	Name       string
	Value      interface{}
	Index      int
	MetaValues *MetaValues
	Meta       Metaor
	error      error
}

func decodeMetaValuesToField(res Resourcer, field reflect.Value, metaValue *MetaValue, context *TM_EC.Context) {
	if field.Kind() == reflect.Struct {
		value := reflect.New(field.Type())
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			field.Set(value.Elem())
		}
	} else if field.Kind() == reflect.Slice {
		if metaValue.Index == 0 {
			field.Set(reflect.Zero(field.Type()))
		}

		fieldType := field.Type().Elem()
		var isPtr bool
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtr = true
		}

		value := reflect.New(fieldType)
		associationProcessor := DecodeToResource(res, value.Interface(), metaValue.MetaValues, context)
		associationProcessor.Start()
		if !associationProcessor.SkipLeft {
			if !reflect.DeepEqual(reflect.Zero(fieldType).Interface(), value.Elem().Interface()) {
				if isPtr {
					field.Set(reflect.Append(field, value))
				} else {
					field.Set(reflect.Append(field, value.Elem()))
				}
			}
		}
	}
}
