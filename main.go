package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	graphql "github.com/graph-gophers/graphql-go"
	handler "github.com/zhaoyi0113/graphql-apigateway-subscription/handler"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/resolver"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/schema"
)

var graphqlSchema *graphql.Schema
var h *handler.Handler

func init() {
	var s, err = schema.GetSchema()
	if err != nil {
		panic("Failed to load schema")
	}
	println("schema:", s)
	graphqlSchema = graphql.MustParseSchema(s, &resolver.Resolver{})
	h = handler.New(graphqlSchema)
}

func setupLocalEnv() {
	// http.Handle("/query", &relay.Handler{Schema: graphqlSchema})
	println("set up local env")
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body)
		if b, err := io.ReadAll(r.Body); err == nil {
			resp, err := h.GraphqlHandler(r.Context(), events.APIGatewayProxyRequest{Body: string(b)})
			if err != nil {
				println(err)
				println("Failed to execute")
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(resp.Body))
		}
	})
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphiql.html")
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	lambdaEnv := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if len(lambdaEnv) == 0 {
		setupLocalEnv()
	} else {
		lambda.Start(h.GraphqlHandler)
	}
}
