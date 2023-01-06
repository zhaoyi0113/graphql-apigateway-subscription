package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

type GraphqlQuery struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type GraphqlWSEvent struct {
	Id      string
	Type    string
	Payload GraphqlQuery
}

func (h *Handler) GraphqlHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	println("receive event", event.Headers, event.Path, event.Body)
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(event.Body), &params); err != nil {
		log.Println("Failed to parse body", err, event)
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
	log.Println("event.RequestContext ConnectionID:", event.RequestContext.ConnectionID)
	log.Println("context:", lc)
	log.Println("body", event.Body)
	log.Println("PathParameters", event.PathParameters)
	log.Println("QueryStringParameters", event.QueryStringParameters)
	// type params struct {
	// 	Query         string                 `json:"query"`
	// 	OperationName string                 `json:"operationName"`
	// 	Variables     map[string]interface{} `json:"variables"`
	// }
	// var bodyType struct {
	// 	Id      string `json:"id"`
	// 	Type    string `json:"type"`
	// 	Payload params `json:"payload"`
	// }
	// if err := json.Unmarshal([]byte(event.Body), &bodyType); err != nil {
	// 	log.Println("Failed to parse body", err, event)
	// 	return events.APIGatewayProxyResponse{Body: "Can't parse request body", StatusCode: 400}, nil
	// }
	// fmt.Println("payload", bodyType.Payload)
	// response := h.schema.Exec(ctx, bodyType.Payload.Query, bodyType.Payload.OperationName, bodyType.Payload.Variables)
	// responseJSON, err := json.Marshal(response)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{Body: "Can't execute graphql request", StatusCode: 500}, nil
	// }

	// return events.APIGatewayProxyResponse{Body: string(responseJSON), StatusCode: 200}, nil
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func (h *Handler) GraphqlDefaultSubscriptionHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("receive event:", event)
	lc, _ := lambdacontext.FromContext(ctx)
	log.Println("event.RequestContext:", event.RequestContext.ConnectionID)
	log.Println("context:", lc)
	log.Println("body", event.Body)
	log.Println("PathParameters", event.PathParameters)
	log.Println("QueryStringParameters", event.QueryStringParameters)

	var params GraphqlWSEvent
	if err := json.Unmarshal([]byte(event.Body), &params); err != nil {
		log.Println("Failed to parse body", err, event)
		return events.APIGatewayProxyResponse{Body: "Can't parse request body", StatusCode: 400}, nil
	}
	if params.Type == "start" {
		fmt.Printf("Get start type %#v\n", params)
	}
	if params.Type == "connection_init" {
		fmt.Printf("Get connection_init type %#v\n", params)
		return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
	}
	fmt.Println("Exec graphql query:", params)
	channel, err := h.schema.Subscribe(ctx, params.Payload.Query, params.Payload.OperationName, params.Payload.Variables)
	fmt.Println("channel", channel)
	if err != nil {
		log.Println("Subscribe failed", err)
		return events.APIGatewayProxyResponse{Body: "Can't execute graphql request", StatusCode: 500}, nil
	}
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func (h *Handler) Subscribe(ctx context.Context, operationName string, query string, variables map[string]interface{}) <-chan interface{} {
	response := make(chan interface{})
	go func() {
		fmt.Println("subscribe", operationName)
		res, err := h.schema.Subscribe(ctx, query, operationName, variables)
		if err != nil {
			log.Println(err)
		}
		r := <-res
		response <- r
	}()
	return response
}

func (h *Handler) Exec(ctx context.Context, operationName string, query string, variables map[string]interface{}) {
	fmt.Println("exec", operationName)
	response := h.schema.Exec(ctx, query, operationName, variables)
	j, _ := json.Marshal(&response.Data)
	fmt.Println("exec ", operationName, "response:", string(j))
}
