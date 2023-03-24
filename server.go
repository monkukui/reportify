package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/jlwt90/gqlgen-usage-analysis/extension"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jlwt90/gqlgen-usage-analysis/graph"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	c := graph.Config{Resolvers: &graph.Resolver{}}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, id string) (res interface{}, err error) {
		return next(ctx)
	}
	c.Directives.Lang = func(ctx context.Context, obj interface{}, next graphql.Resolver, region string) (res interface{}, err error) {
		return next(ctx)
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))
	srv.Use(extension.AuditLogger{Writer: os.Stdout})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
