package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ConnectionDb struct {
	db *dynamodb.Client
}

type ConnectionItemType struct {
	Id          string `dynamodbav:"id"`
	Type        string `dynamodbav:"type"`
	CreatedTime string `dynamodbav:"createdTime"`
	Status      string `dynamodbav:"status"`
}

type SubscriberItemType struct {
	Id          string `dynamodbav:"id"`
	Type        string `dynamodbav:"type"`
	CreatedTime string `dynamodbav:"createdTime"`
	Topic       string `dynamodbav:"topic"`
}

type EventItemType struct {
	Id          string `dynamodbav:"id"`
	Type        string `dynamodbav:"type"`
	CreatedTime string `dynamodbav:"createdTime"`
	Topic       string `dynamodbav:"topic"`
	Message     string `dynamodbav:"message"`
}

type ItemType string

const (
	CONNECTION ItemType = "connection.core"
	SUBSCRIBER ItemType = "connection.subscriber"
	Event      ItemType = "event."
)

const tableName = "Connection"

func NewConnectionDb() *ConnectionDb {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-2"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB client
	dynamodbClient := ConnectionDb{}
	dynamodbClient.db = dynamodb.NewFromConfig(cfg)
	return &dynamodbClient
}

func (c *ConnectionDb) SaveConnection(id string) {
	fmt.Println("Save connection", id, "on db")
	itemMap := ConnectionItemType{
		Id:          id,
		Type:        string(CONNECTION),
		CreatedTime: time.Now().Format(time.RFC3339),
		Status:      "CONNECTED",
	}
	item, err := attributevalue.MarshalMap(itemMap)
	if err != nil {
		log.Panic("Failed to marsh item", id)
	}
	fmt.Println("Marshaled item", item)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}
	res, err := c.db.PutItem(context.Background(), input)
	if err != nil {
		log.Panic("Failed to save item on db", id, err)
	}
	fmt.Println("Save item response", res)
}

func (c *ConnectionDb) Disconnect(id string) {
	fmt.Println("Disconnect connection", id)
	key := struct {
		Id   string `dynamodbav:"id"`
		Type string `dynamodbav:"type"`
	}{Id: id, Type: string(CONNECTION)}
	item, err := attributevalue.MarshalMap(key)
	if err != nil {
		log.Panic("Failed to mashal key", id)
	}
	output, err := c.db.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key: item,
	})
	if err != nil {
		log.Panic("Cant delete item", id)
	}
	fmt.Println("Delete item output", output)
}

func (c *ConnectionDb) SaveSubscriber(id string, topic string) {
	fmt.Println("Save connection", id, topic, "on db")
	itemMap := SubscriberItemType{
		Id:          id,
		Type:        string(SUBSCRIBER),
		CreatedTime: time.Now().Format(time.RFC3339),
		Topic:       topic,
	}
	item, err := attributevalue.MarshalMap(itemMap)
	if err != nil {
		log.Panic("Failed to marsh item", id)
	}
	fmt.Println("Marshaled item", item)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}
	res, err := c.db.PutItem(context.Background(), input)
	if err != nil {
		log.Panic("Failed to save item on db", id, err)
	}
	fmt.Println("Save item response", res)
}

func (c *ConnectionDb) SaveEvent(id string, topic string, message string) {

	fmt.Println("Save connection", id, topic, "on db")
	itemMap := EventItemType{
		Id:          id,
		Type:        string(Event) + fmt.Sprint(time.Now().UnixMilli()),
		CreatedTime: time.Now().Format(time.RFC3339),
		Topic:       topic,
		Message:     message,
	}
	item, err := attributevalue.MarshalMap(itemMap)
	if err != nil {
		log.Panic("Failed to marsh item", id)
	}
	fmt.Println("Marshaled item", item)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}
	res, err := c.db.PutItem(context.Background(), input)
	if err != nil {
		log.Panic("Failed to save item on db", id, err)
	}
	fmt.Println("Save item response", res)
}

func (c *ConnectionDb) GetSubscribers(id string) []map[string]types.AttributeValue {
	key := struct {
		Id   string `dynamodbav:"id"`
		Type string `dynamodbav:"type"`
	}{Id: id, Type: string(SUBSCRIBER)}
	item, err := attributevalue.MarshalMap(key)
	if err != nil {
		log.Panic("Failed to marsh item", id, err)
	}
	fmt.Println("Fetch item from db", id, item)
	out, err := c.db.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                aws.String(tableName),
		KeyConditionExpression:   aws.String("id = :id and #type > :type"),
		ExpressionAttributeNames: map[string]string{"#type": "type"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id":   &types.AttributeValueMemberS{Value: id},
			":type": &types.AttributeValueMemberS{Value: string(SUBSCRIBER)},
		},
	})
	fmt.Println("Fetched item from db", out)
	if err != nil {
		log.Panic("Failed to get item", id, err)
	}
	return out.Items
}
