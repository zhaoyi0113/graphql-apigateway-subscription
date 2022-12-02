package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	graphql "github.com/graph-gophers/graphql-go"
)

type Handler struct {
	schema *graphql.Schema
}

func New(schema *graphql.Schema) *Handler {
	h := Handler{schema}
	return &h
}

func (h *Handler) GraphqlHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	println("receive event", event.Headers, event.Path, event.Body)
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(event.Body), &params); err != nil {
		return events.APIGatewayProxyResponse{Body: "Can't parse request body", StatusCode: 400}, nil
	}
	println("query:", params.Query)
	println("OperationName:", params.OperationName)
	println("var:", params.Variables)
	response := h.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Can't execute graphql request", StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(responseJSON), StatusCode: 200}, nil
}

func (h *Handler) GraphqlSubscriptionHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("receive event:", event)
	lc, _ := lambdacontext.FromContext(ctx)
	log.Println("event.RequestContext:", event.RequestContext.ConnectionID)
	log.Print("context:", lc)
	// response := h.schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	// responseJSON, err := json.Marshal(response)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{Body: "Can't execute graphql request", StatusCode: 500}, nil
	// }

	// return events.APIGatewayProxyResponse{Body: string(responseJSON), StatusCode: 200}, nil
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}
