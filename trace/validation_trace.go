package trace

import (
	"github.com/graph-gophers/graphql-go/errors"
)

type TraceValidationFinishFunc = func([]*errors.QueryError)

type ValidationTracer interface {
	TraceValidation() TraceValidationFinishFunc
}

type NoopValidationTracer struct{}

func (NoopValidationTracer) TraceValidation() TraceValidationFinishFunc {
	return func(errs []*errors.QueryError) {}
}
