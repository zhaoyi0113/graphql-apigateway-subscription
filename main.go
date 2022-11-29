package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	graphql "github.com/graph-gophers/graphql-go"
	handler "github.com/zhaoyi0113/graphql-apigateway-subscription/handler"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/resolver"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/schema"
)

var graphqlSchema *graphql.Schema

func init() {
	var s, err = schema.GetSchema()
	if err != nil {
		panic("Failed to load schema")
	}
	println("schema:", s)
	graphqlSchema = graphql.MustParseSchema(s, &resolver.Resolver{})
}

func main() {
	h := handler.New(graphqlSchema)
	lambda.Start(h.GraphqlHandler)
	// http.Handle("/query", &relay.Handler{Schema: schema})
	// http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
	// 	var params struct {
	// 		Query         string                 `json:"query"`
	// 		OperationName string                 `json:"operationName"`
	// 		Variables     map[string]interface{} `json:"variables"`
	// 	}
	// 	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	println("query:", params.Query)
	// 	println("OperationName:", params.OperationName)
	// 	println("var:", params.Variables)
	// 	response := graphqlSchema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)
	// 	responseJSON, err := json.Marshal(response)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(responseJSON)
	// })

	// http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "graphiql.html")
	// }))

	// log.Fatal(http.ListenAndServe(":8080", nil))
}
