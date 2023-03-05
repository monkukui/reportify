package extension

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
)

const auditLoggerExtension = "AuditLogger"

type AuditLogger struct {
	Writer io.Writer
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = &AuditLogger{}

func (a AuditLogger) ExtensionName() string {
	return auditLoggerExtension
}

func (a AuditLogger) Validate(schema graphql.ExecutableSchema) error {
	if a.Writer == nil {
		return fmt.Errorf("AuditLogger Writer can not be nil")
	}
	return nil
}

func (a AuditLogger) MutateOperationContext(ctx context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	_, err := fmt.Fprintf(a.Writer, "Operation Name) %s\n", rc.OperationName)
	if err != nil {
		return nil
	}
	_, err = fmt.Fprintf(a.Writer, "Query) %s\n", rc.RawQuery)
	if err != nil {
		return nil
	}
	return nil
}
