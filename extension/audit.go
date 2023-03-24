package extension

import (
	"context"
	"encoding/json"
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
	graphql.FieldInterceptor
	graphql.ResponseInterceptor
	graphql.HandlerExtension
} = &AuditLogger{}

func (a AuditLogger) ExtensionName() string {
	return auditLoggerExtension
}

type DataPayload struct {
	RawQuery  string
	Variables map[string]interface{}
}

type MetaData struct {
	Resolvers map[string]bool
	Tags      map[string]map[string]bool
}

type Meta struct {
	Resolvers []string
	Tags      map[string][]string
}

type AuditLog struct {
	GraphQLOperation string
	Payload          DataPayload
	Meta             Meta
}

func (a AuditLogger) Validate(_ graphql.ExecutableSchema) error {
	if a.Writer == nil {
		return fmt.Errorf("AuditLogger Writer can not be nil")
	}
	return nil
}

func (a AuditLogger) MutateOperationContext(_ context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	m := &MetaData{
		Resolvers: map[string]bool{},
		Tags: map[string]map[string]bool{
			"hasRole": {},
			"lang":    {},
		},
	}
	rc.Stats.SetExtension(auditLoggerExtension, m)
	return nil
}

func (a AuditLogger) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	op := graphql.GetOperationContext(ctx)
	m := op.Stats.GetExtension(auditLoggerExtension).(*MetaData)

	fc := graphql.GetFieldContext(ctx)
	callBy := fc.Object
	field := fc.Field.Name
	key := fmt.Sprintf("%s:%s", callBy, field)

	if fc.IsResolver {
		if callBy == "Query" || callBy == "Mutation" || callBy == "Subscription" {
			if _, found := m.Resolvers[key]; !found {
				m.Resolvers[key] = true
			}
		}
	}
	def := fc.Field.Definition
	if def.Directives.ForName("hasRole") != nil {
		if v, found := m.Tags["hasRole"]; found {
			if _, found = v[key]; !found {
				v[key] = true
			}
		}
	}
	return next(ctx)
}

func (a AuditLogger) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	defer func() {
		op := graphql.GetOperationContext(ctx)
		operationName, _ := a.getOperation(ctx)

		m := op.Stats.GetExtension(auditLoggerExtension).(*MetaData)
		tags := make(map[string][]string, len(m.Tags))
		for k, v := range m.Tags {
			tags[k] = a.keysOf(v)
		}

		log := AuditLog{
			GraphQLOperation: operationName,
			Payload: DataPayload{
				RawQuery:  op.RawQuery,
				Variables: op.Variables,
			},
			Meta: Meta{
				Resolvers: a.keysOf(m.Resolvers),
				Tags:      tags,
			},
		}
		a.logRequest(log)
	}()

	return next(ctx)
}

func (a AuditLogger) logRequest(l AuditLog) {
	b, err := json.Marshal(l)
	if err != nil {
		_ = fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(b))
}

func (a AuditLogger) getOperation(ctx context.Context) (name, opType string) {
	defer func() {
		if err := recover(); err != nil {
			_ = fmt.Errorf("failed to get operation name: %v", err)
		}
	}()

	op := graphql.GetOperationContext(ctx)
	name = "unnamed operation"
	if op != nil && op.Operation != nil {
		op := op.Operation
		if op.Name != "" {
			name = op.Name
		}
		if op.Operation != "" {
			opType = string(op.Operation)
		}
	}

	return name, opType
}

func (a AuditLogger) keysOf(m map[string]bool) []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	return keys
}
