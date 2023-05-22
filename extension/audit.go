package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/atomic"
	"io"
	"strings"
	"sync"
)

const auditLoggerExtension = "AuditLogger"

type AuditLogger struct {
	Writer io.Writer
}

var _ interface {
	//graphql.OperationInterceptor
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
	Resolvers *sync.Map
	Tags      *sync.Map
}

type Meta struct {
	Resolvers map[string]int
	Tags      map[string]map[string]int
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

func (a AuditLogger) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	op := graphql.GetOperationContext(ctx)
	m := &MetaData{
		Resolvers: &sync.Map{},
		Tags:      &sync.Map{},
	}
	op.Stats.SetExtension(auditLoggerExtension, m)
	return next(ctx)
}

func (a AuditLogger) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	op := graphql.GetOperationContext(ctx)
	m, ok := op.Stats.GetExtension(auditLoggerExtension).(*MetaData)
	if !ok {
		fmt.Println("[InterceptField] no extension context")
	}

	fc := graphql.GetFieldContext(ctx)
	callBy := fc.Object
	field := fc.Field.Name
	key := fmt.Sprintf("%s:%s", callBy, field)

	if fc.IsResolver {
		if callBy == "Query" || callBy == "Mutation" || callBy == "Subscription" {
			val, _ := m.Resolvers.LoadOrStore(key, &atomic.Int32{})
			if v, ok := val.(*atomic.Int32); ok {
				v.Add(1)
			}
		}
	}
	def := fc.Field.Definition
	if def.Directives.ForName("hasRole") != nil {
		val, _ := m.Tags.LoadOrStore("hasRole"+"|"+key, &atomic.Int32{})
		if v, ok := val.(*atomic.Int32); ok {
			v.Add(1)
		}
	}
	return next(ctx)
}

func (a AuditLogger) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	defer func() {
		op := graphql.GetOperationContext(ctx)
		operationName, _ := a.getOperation(ctx)

		m, ok := op.Stats.GetExtension(auditLoggerExtension).(*MetaData)
		if !ok {
			fmt.Println("[InterceptResoponse] No extension context")
		}

		resolvers := make(map[string]int)
		m.Resolvers.Range(func(key, value any) bool {
			k, ok := key.(string)
			if !ok {
				fmt.Println(fmt.Sprintf("cannot cast key(%v) to string", k))
				return false
			}
			v, ok := value.(*atomic.Int32)
			if !ok {
				fmt.Println(fmt.Sprintf("cannot cast value(%v) to string", v))
				return false
			}
			resolvers[k] = int(v.Load())
			return true
		})
		tags := make(map[string]map[string]int)
		tags["hasRole"] = make(map[string]int)
		tags["lang"] = make(map[string]int)
		m.Tags.Range(func(key, value any) bool {
			k, ok := key.(string)
			if !ok {
				fmt.Println(fmt.Sprintf("cannot cast key(%v) to string", k))
				return false
			}
			v, ok := value.(*atomic.Int32)
			if !ok {
				fmt.Println(fmt.Sprintf("cannot cast key(%v) to *atomic.Int32", k))
				return false
			}

			s := strings.Split(k, "|")
			if len(s) != 2 {
				fmt.Println(fmt.Sprintf("the length of the secuences splited by '|' should be 2, but it is actually %d", len(s)))
				return false
			}

			tags[s[0]][s[1]] = int(v.Load())
			return true
		})

		log := AuditLog{
			GraphQLOperation: operationName,
			Payload: DataPayload{
				RawQuery:  op.RawQuery,
				Variables: op.Variables,
			},
			Meta: Meta{
				Resolvers: resolvers,
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
	if op != nil {
		// query & mutation
		if op.Operation != nil {
			op := op.Operation
			if op.Name != "" {
				name = op.Name
			}
			if op.Operation != "" {
				opType = string(op.Operation)
			}
		} else {
			// subscription
			name = op.OperationName
			opType = "subscription"
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
