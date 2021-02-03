package trace

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/selected"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// Ensure our own types respects the interface
var _ Tracer = OpenTracingTracer{}
var _ Tracer = NoopTracer{}

type TraceRequestFinishFunc func([]*errors.QueryError)
type TraceOperationFinishFunc func()
type TraceFieldFinishFunc func(*errors.QueryError)

type Tracer interface {
	TraceRequest(ctx context.Context, queryString string, opType string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, TraceRequestFinishFunc)

	TraceQuery(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc)
	TraceMutation(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc)
	TraceSubscription(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc)

	TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, TraceFieldFinishFunc)
}

type OpenTracingTracer struct{}

func (OpenTracingTracer) TraceRequest(ctx context.Context, queryString string, opType string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, TraceRequestFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "GraphQL request")
	span.SetTag("graphql.type", opType)
	span.SetTag("graphql.query", queryString)

	if operationName != "" {
		span.SetTag("graphql.operationName", operationName)
	}

	if len(variables) != 0 {
		span.LogFields(log.Object("graphql.variables", variables))
	}

	return spanCtx, func(errs []*errors.QueryError) {
		if len(errs) > 0 {
			msg := errs[0].Error()
			if len(errs) > 1 {
				msg += fmt.Sprintf(" (and %d more errors)", len(errs)-1)
			}
			ext.Error.Set(span, true)
			span.SetTag("graphql.error", msg)
		}
		span.Finish()
	}
}

func (OpenTracingTracer) TraceQuery(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "GraphQL Query")
	span.SetTag("graphql.type", root.Identifier())

	return spanCtx, func() {
		span.Finish()
	}
}

func (OpenTracingTracer) TraceMutation(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "GraphQL Mutation")
	span.SetTag("graphql.type", root.Identifier())

	return spanCtx, func() {
		span.Finish()
	}
}

func (OpenTracingTracer) TraceSubscription(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "GraphQL Subscription")
	span.SetTag("graphql.type", root.Identifier())

	return spanCtx, func() {
		span.Finish()
	}
}

func (OpenTracingTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, TraceFieldFinishFunc) {
	if trivial {
		return ctx, noop
	}

	span, spanCtx := opentracing.StartSpanFromContext(ctx, label)
	span.SetTag("graphql.type", typeName)
	span.SetTag("graphql.field", fieldName)
	for name, value := range args {
		span.SetTag("graphql.args."+name, value)
	}

	return spanCtx, func(err *errors.QueryError) {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("graphql.error", err.Error())
		}
		span.Finish()
	}
}

func noop(*errors.QueryError) {}

type NoopTracer struct{}

func (NoopTracer) TraceRequest(ctx context.Context, queryString string, opType string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, TraceRequestFinishFunc) {
	return ctx, func(errs []*errors.QueryError) {}
}

func (NoopTracer) TraceQuery(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	return ctx, func() {}
}

func (NoopTracer) TraceMutation(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	return ctx, func() {}
}

func (NoopTracer) TraceSubscription(ctx context.Context, root selected.Field) (context.Context, TraceOperationFinishFunc) {
	return ctx, func() {}
}

func (NoopTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, TraceFieldFinishFunc) {
	return ctx, func(err *errors.QueryError) {}
}
