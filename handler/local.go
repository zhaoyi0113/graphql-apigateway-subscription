package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	ast "github.com/vektah/gqlparser/ast"
	parser "github.com/vektah/gqlparser/parser"
	"github.com/zhaoyi0113/graphql-apigateway-subscription/schema"
	"golang.org/x/exp/slices"
)

const tName = "Connection"

func setupLocalTable(ctx context.Context) {
	cfg, _ := ctx.Value("awsconfig").(aws.Config)
	db := dynamodb.NewFromConfig(cfg)
	var lastEvaluatedTableName *string
	existTable := false
	for {
		tables, err := db.ListTables(ctx, &dynamodb.ListTablesInput{
			ExclusiveStartTableName: lastEvaluatedTableName,
		})
		fmt.Println("table name", tables)
		if slices.Contains(tables.TableNames, tName) {
			existTable = true
			break
		}
		if err != nil {
			log.Panic("Failed to list tables", err)
		}
		if tables.LastEvaluatedTableName == nil {
			break
		}
		lastEvaluatedTableName = tables.LastEvaluatedTableName
	}
	if !existTable {
		output, err := db.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(tName),
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("type"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       types.KeyTypeHash,
				}, {
					AttributeName: aws.String("type"),
					KeyType:       types.KeyTypeRange,
				},
			},
			BillingMode: types.BillingModePayPerRequest,
		})
		if err != nil {
			log.Panic("Failed to create db table", err)
		}
		fmt.Printf("create table response %#v\n", output)
	}
}

func SetupLocalEnv(ctx context.Context) context.Context {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-2"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			},
		)))
	if err != nil {
		log.Panic("Failed to load aws config")
	}
	ctx = context.WithValue(ctx, "awsconfig", cfg)
	setupLocalTable(ctx)
	return ctx
}

func Test(h *Handler) {
	event := GraphqlWSEvent{
		Id:   "001",
		Type: "start",
		Payload: GraphqlQuery{
			OperationName: "event",
			Query:         "subscription event { event(on: \"xxxx\" ) { msg }}",
		},
	}
	e, _ := json.Marshal(event)
	response := h.graphqlMessageHandler(context.Background(), events.APIGatewayWebsocketProxyRequest{
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			EventType:    "MESSAGE",
			ConnectionID: "1",
		},
		Body: string(e),
	})
	fmt.Println("client response", response)

	time.Sleep(3 * time.Second)

	var s, _ = schema.GetSchema()
	doc, _ := parser.ParseSchema(&ast.Source{Input: s})
	fmt.Printf("doc: %#v\n", doc)

	fmt.Println("sendChat mutation")
	res := h.Exec(context.TODO(), "sendChat", "mutation sendChat{\n sendChat(topic: \"1\", message: \"hello\") }\n", nil)
	fmt.Println("response:", res)
	time.Sleep(50 * time.Second)
}
