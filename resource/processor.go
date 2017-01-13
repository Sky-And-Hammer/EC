package resource

import (
	"errors"
	"reflect"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"

	"github.com/Sky-And-Hammer/TM_EC"
	"github.com/Sky-And-Hammer/roles"
)

//	'ErrProcessorSkipLeft' skip left processors error, if returned this error in validation, beform callbacks, then EC will stop process following processers
var ErrProcessorSkipLeft = errors.New("resource: skip left")

type processor struct {
	Result     interface{}
	Resource   Resourcer
	Context    *TM_EC.Context
	MetaValues *MetaValues
	SkipLeft   bool
	newRecord  bool
}

//	'DecodeToResource' decode meta values to resource result
func DecodeToResource(res Resourcer, result interface{}, metaValues *MetaValues, context *TM_EC.Context) *processor {
	scope := &gorm.Scope{
		Value: result,
	}
	return &processor{
		Resource:   res,
		Result:     result,
		Context:    context,
		MetaValues: metaValues,
		newRecord:  scope.PrimaryKeyZero(),
	}
}

func (processor *processor) checkSkipLeft(errs ...error) bool {
	if processor.SkipLeft {
		return true
	}

	for _, err := range errs {
		if err == ErrProcessorSkipLeft {
			processor.SkipLeft = true
			break
		}
	}
	return processor.SkipLeft
}

func (processor *processor) Initialize() error {
	err := processor.Resource.CallFindOne(processor.Resource, processor.MetaValues, processor.Context)
	processor.checkSkipLeft(err)
	return err
}

func (processor *processor) Validate() error {
	var errors TM_EC.Errors
	if processor.checkSkipLeft() {
		return nil
	}

	for _, fc := range processor.Resource.GetResource().Validatiors {
		if errors.AddError(fc(processor.Result, processor.MetaValues, processor.Context)); !errors.HasError() {
			if processor.checkSkipLeft(errors.GetErrors()...) {
				break
			}
		}
	}
	return errors
}

func (processor *processor) decode() (errors []error) {
	if processor.checkSkipLeft() || processor.MetaValues == nil {
		return
	}

	for _, metaValue := range processor.MetaValues.Values {
		meta := metaValue.Meta
		if meta == nil {
			continue
		}

		if processor.newRecord && !meta.HasPermission(roles.Create, processor.Context) {
			continue
		} else if !meta.HasPermission(roles.Update, processor.Context) {
			continue
		}

		if setter := meta.GetSetter(); setter != nil {
			setter(processor.Result, metaValue, processor.Context)
			continue
		}

		res := metaValue.Meta.GetResource()
		if res == nil {
			continue
		}

		field := reflect.Indirect(reflect.ValueOf(processor.Result)).FieldByName(meta.GetFieldName())
		decodeMetaValuesToField(res, field, metaValue, processor.Context)
	}
	return
}

func (processor *processor) Commit() error {
	var errors TM_EC.Errors
	errors.AddError(processor.decode()...)
	if processor.checkSkipLeft(errors.GetErrors()...) {
		return nil
	}

	for _, fc := range processor.Resource.GetResource().Processors {
		if err := fc(processor.Result, processor.MetaValues, processor.Context); err != nil {
			if processor.checkSkipLeft(err) {
				break
			}
			errors.AddError(err)
		}
	}
	return errors
}

func (processor *processor) Start() error {
	var errors TM_EC.Errors
	processor.Initialize()
	if errors.AddError(processor.Validate()); !errors.HasError() {
		errors.AddError(processor.Commit())
	}

	if errors.HasError() {
		return errors
	}

	return nil
}
