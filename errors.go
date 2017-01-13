package TM_EC

import (
	"strings"
)

type Errors struct {
	errors []error
}

func (errs Errors) Error() string {
	var errors []string
	for _, err := range errs.errors {
		errors = append(errors, err.Error())
	}

	return strings.Join(errors, ";")
}

func (errs *Errors) AddError(errors ...error) {
	for _, err := range errs.errors {
		if err != nil {
			if e, ok := err.(errorsInterface); ok {
				errs.errors = append(errs.errors, e.GetErrors()...)
			} else {
				errs.errors = append(errs.errors, err)
			}
		}
	}
}

func (errs Errors) HasError() bool {
	return len(errs.errors) != 0
}

//	'GetErrors'
func (errs Errors) GetErrors() []error {
	return errs.errors
}

type errorsInterface interface {
	GetErrors() []error
}
