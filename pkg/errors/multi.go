package errors

import (
	"strings"
)

type MultiError struct {
	errs []error
}

func Multi(errs ... error) *MultiError {
	return &MultiError{
		errs: errs,
	}
}

func (m *MultiError) Error() string {
	errs := make([]string, 0, len(m.errs))
	for _, e := range m.errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, ", ")
}
