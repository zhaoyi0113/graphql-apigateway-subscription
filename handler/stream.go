package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

func GetApiClient() *apigatewaymanagementapi.Client {
	defaultRegion := "ap-southeast-2"
	url := "https://jkpmcizu0i.execute-api.ap-southeast-2.amazonaws.com/dev/"
	resolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == apigatewaymanagementapi.ServiceID && region == defaultRegion {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           url,
				SigningRegion: defaultRegion,
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(defaultRegion),
		config.WithEndpointResolver(resolver),
	)
	if err != nil {
		log.Panic("cfg err")
	}

	return apigatewaymanagementapi.NewFromConfig(cfg)
}
func StreamHandler(ctx context.Context, e events.DynamoDBEvent) {
	fmt.Println("Receive stream event", e)
	api := GetApiClient()
	db := NewConnectionDb(ctx)
	for _, record := range e.Records {
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		topic := record.Change.NewImage["topic"].String()
		message := record.Change.NewImage["message"].String()
		fmt.Println("get event", topic, message)
		subscribers := db.GetSubscribers(topic)
		fmt.Println("Get subscribers:", subscribers)
		for _, subscriber := range subscribers {
			sub := map[string]string{}
			attributevalue.UnmarshalMap(subscriber, &sub)
			fmt.Printf("subscribe data %#v\n", sub)
			connectionId := sub["connectionId"]
			payload := map[string]interface{}{
				"type": "data",
				"id":   sub["eventId"],
				"payload": map[string]interface{}{
					"data": map[string]interface{}{
						"event": map[string]interface{}{
							"msg":   message,
							"topic": topic,
						},
					},
				},
			}
			j, _ := json.Marshal(payload)
			fmt.Println("publish to", connectionId, string(j))
			output, err := api.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: &connectionId,
				Data:         j,
			})
			if err != nil {
				log.Println("Failed to post to connection", err)
				continue
			}
			fmt.Println("post to connection response", output)
		}
	}
}
