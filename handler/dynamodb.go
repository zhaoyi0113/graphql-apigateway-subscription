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
	Id           string `dynamodbav:"id"`
	Type         string `dynamodbav:"type"`
	CreatedTime  string `dynamodbav:"createdTime"`
	Topic        string `dynamodbav:"topic"`
	ConnectionId string `dynamodbav:"connectionId"`
	EventId      string `dynamodbav:"eventId"`
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
	SUBSCRIBER ItemType = "connection.subscriber."
	Event      ItemType = "event."
)

const tableName = "Connection"

func NewConnectionDb(ctx context.Context) *ConnectionDb {
	cfg, err := ctx.Value("awsconfig").(aws.Config)
	if err == false {
		localcfg, e := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-2"))
		if e != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}
		cfg = localcfg
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
		Key:       item,
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.Println("Cant delete item", id, err)
	}

	out, err := c.db.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                aws.String(tableName),
		IndexName:                aws.String("typeGsi"),
		KeyConditionExpression:   aws.String("#type = :type"),
		ExpressionAttributeNames: map[string]string{"#type": "type"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":type": &types.AttributeValueMemberS{Value: string(SUBSCRIBER) + id},
		},
	})
	fmt.Println("Fetched item from db", string(SUBSCRIBER)+id, out.Count)
	if err != nil {
		log.Panic("Failed to get item", id, err)
	}

	for _, o := range out.Items {
		data := map[string]string{}
		attributevalue.UnmarshalMap(o, &data)
		fmt.Println("Delete item", data["id"], data["type"])
		key = struct {
			Id   string `dynamodbav:"id"`
			Type string `dynamodbav:"type"`
		}{Id: data["id"], Type: data["type"]}
		k, _ := attributevalue.MarshalMap(key)
		c.db.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			Key:       k,
			TableName: aws.String(tableName),
		})
		fmt.Println("Delete item output", output)
	}

}

func (c *ConnectionDb) SaveSubscriber(id string, topic string, eventId string) {
	fmt.Println("Save connection", id, topic, "on db")
	itemMap := SubscriberItemType{
		Id:           topic,
		Type:         string(SUBSCRIBER) + id,
		CreatedTime:  time.Now().Format(time.RFC3339),
		Topic:        topic,
		ConnectionId: id,
		EventId:      eventId,
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

func (c *ConnectionDb) SaveEvent(topic string, message string) {

	fmt.Println("Save connection", topic, "on db")
	itemMap := EventItemType{
		Id:          topic,
		Type:        string(Event) + fmt.Sprint(time.Now().UnixMilli()),
		CreatedTime: time.Now().Format(time.RFC3339),
		Topic:       topic,
		Message:     message,
	}
	item, err := attributevalue.MarshalMap(itemMap)
	if err != nil {
		log.Panic("Failed to marsh item", topic)
	}
	fmt.Println("Marshaled item", item)
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}
	res, err := c.db.PutItem(context.Background(), input)
	if err != nil {
		log.Panic("Failed to save item on db", topic, err)
	}
	fmt.Println("Save item response", res)
}

func (c *ConnectionDb) GetSubscribers(topic string) []map[string]types.AttributeValue {
	fmt.Printf("Fetch item from db %#v, %#v\n", topic, string(SUBSCRIBER))
	out, err := c.db.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                aws.String(tableName),
		IndexName:                aws.String("topicGsi"),
		KeyConditionExpression:   aws.String("topic = :topic AND begins_with(#type, :type)"),
		ExpressionAttributeNames: map[string]string{"#type": "type"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":topic": &types.AttributeValueMemberS{Value: topic},
			":type":  &types.AttributeValueMemberS{Value: string(SUBSCRIBER)},
		},
	})
	fmt.Println("Fetched item from db", out.Count)
	if err != nil {
		log.Panic("Failed to get item", topic, err)
	}
	return out.Items
}
