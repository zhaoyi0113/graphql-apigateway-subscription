package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	graphql "github.com/graph-gophers/graphql-go"
	ast "github.com/vektah/gqlparser/ast"
	parser "github.com/vektah/gqlparser/parser"
)

type Handler struct {
	schema       *graphql.Schema
	connectionDb *ConnectionDb
}

func New(schema *graphql.Schema) *Handler {
	h := Handler{schema, NewConnectionDb()}
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

type ConnectId struct {
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

func (h *Handler) GraphqlDefaultHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("event.RequestContext: %#v\n", event.RequestContext)
	switch {
	case event.RequestContext.EventType == "CONNECT":
		return h.graphqlConnectionHandler(ctx, event)
	case event.RequestContext.EventType == "MESSAGE":
		return h.graphqlMessageHandler(ctx, event), nil
	case event.RequestContext.EventType == "DISCONNECT":
		return h.graphqlDisconnectionHandler(ctx, event)
	default:
		fmt.Printf("Unknown connection type %s.", event.RequestContext.EventType)
	}
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 400}, errors.New("Invalid connection type:" + event.RequestContext.EventType)
}

func (h *Handler) graphqlDisconnectionHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Disconnect connection:", event.RequestContext.ConnectionID)
	h.connectionDb.Disconnect(event.RequestContext.ConnectionID)
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func (h *Handler) graphqlConnectionHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("receive event:", event)
	log.Println("event.RequestContext ConnectionID:", event.RequestContext.ConnectionID)
	h.connectionDb.SaveConnection(event.RequestContext.ConnectionID)
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func (h *Handler) graphqlMessageHandler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) events.APIGatewayProxyResponse {
	log.Println("receive event:", event)
	lc, _ := lambdacontext.FromContext(ctx)
	fmt.Printf("event.RequestContext: %#v\n", event.RequestContext)
	log.Println("Connection id:", event.RequestContext.ConnectionID)
	log.Println("context:", lc)
	log.Println("body", event.Body)
	log.Println("PathParameters", event.PathParameters)
	log.Println("QueryStringParameters", event.QueryStringParameters)

	var params GraphqlWSEvent
	if err := json.Unmarshal([]byte(event.Body), &params); err != nil {
		log.Println("Failed to parse body", err, event)
		return events.APIGatewayProxyResponse{Body: "Can't parse request body", StatusCode: 400}
	}
	if params.Type == "start" {
		fmt.Printf("Get start type %#v\n", params)
	}
	if params.Type == "connection_init" {
		fmt.Printf("Get connection_init type %#v\n", params)
		return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}
	}
	payload, _ := json.Marshal(params)
	fmt.Println("Exec graphql query payload: ", string(payload))

	ctx = context.WithValue(ctx, ConnectId{}, event.RequestContext.ConnectionID)

	doc, _ := parser.ParseQuery(&ast.Source{Input: params.Payload.Query})
	for _, o := range doc.Operations {
		fmt.Println("Operation:", o.Operation, o.Name)
		if o.Operation == "subscription" {
			channel := h.Subscribe(ctx, params.Payload.OperationName, params.Payload.Query, params.Payload.Variables)
			select {
			case r := <-channel:
				resp, _ := json.Marshal(r)
				fmt.Println("response from subscription ", string(resp))
				return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}
			case <-time.After(3 * time.Second):
				fmt.Println("subscription success")
				return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}
			}
		} else {
			h.Exec(ctx, params.Payload.OperationName, params.Payload.Query, params.Payload.Variables)
			return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}
		}
	}
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}
}

func (h *Handler) Subscribe(ctx context.Context, operationName string, query string, variables map[string]interface{}) <-chan interface{} {
	response := make(chan interface{})

	go func() {
		fmt.Println("subscribe", operationName, ",", query)
		res, err := h.schema.Subscribe(ctx, query, operationName, variables)
		if err != nil {
			fmt.Println("Subscription failed")
			log.Println(err)
			response <- err
		}
		select {
		case r := <-res:
			response <- r
		case <-time.After(3 * time.Second):
			fmt.Println("reponse subscription successfully.")
			response <- ""
		}
	}()
	return response
}

func (h *Handler) Exec(ctx context.Context, operationName string, query string, variables map[string]interface{}) {
	fmt.Println("exec", operationName)
	response := h.schema.Exec(ctx, query, operationName, variables)

	if response.Errors != nil {
		log.Println(operationName, "response error:", response.Errors)
	}
	j, _ := json.Marshal(&response.Data)
	fmt.Println("exec ", operationName, "response:", string(j))
}
