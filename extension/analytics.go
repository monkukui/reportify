package extension

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

const analyticsExtension = "Analytics"

type Analytics struct {
	Writer io.Writer
}

var _ interface {
	graphql.FieldInterceptor
	graphql.HandlerExtension
} = &Analytics{}

func (a Analytics) ExtensionName() string {
	return auditLoggerExtension
}

func (a Analytics) Validate(schema graphql.ExecutableSchema) error {
	if a.Writer == nil {
		return fmt.Errorf("AuditLogger Writer can not be nil")
	}
	return nil
}

func (a Analytics) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	fctx := graphql.GetFieldContext(ctx)
	if fctx.IsResolver {
		// resolvers
		field := fctx.Field.Name
		callBy := fctx.Object
		if callBy == "Query" || callBy == "Mutation" {
			_, err = fmt.Fprintf(a.Writer, "Root Resolver) %s -> %s\n", callBy, field)
			if err != nil {
				return nil, err
			}
		} else {
			// log parent data as well
			parentRes := fctx.Parent.Result
			_, err = fmt.Fprintf(a.Writer, "Resolver) %s -> %s (parent: %v)\n", callBy, field, parentRes)
			if err != nil {
				return nil, err
			}
		}

		return next(ctx)
	}
	// fields
	callBy := fctx.Object
	field := fctx.Field.Name
	def := fctx.Field.Definition
	// log if hasRole directive exists
	if def.Directives.ForName("hasRole") != nil {
		_, err = fmt.Fprintf(a.Writer, "Field) %s:%s with directives @hasRole\n", callBy, field)
		if err != nil {
			return nil, err
		}
	}

	return next(ctx)
}
