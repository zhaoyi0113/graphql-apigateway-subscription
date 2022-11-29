package main

import (
	"log"
	"net/http"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/resolver"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/schema"
)

func main() {
	s, err := schema.GetSchema()
	if err != nil {
		panic("Failed to load schema")
	}
	println("schema:", s)

	schema := graphql.MustParseSchema(s, &resolver.Resolver{})
	println("compiled schema:", schema)
	http.Handle("/query", &relay.Handler{Schema: schema})

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphiql.html")
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
