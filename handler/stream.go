package handler

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func StreamHandler(ctx context.Context, e events.DynamoDBEvent) {
	fmt.Println("Receive stream event", e)
	// db := NewConnectionDb()
	for _, record := range e.Records {
		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		for name, value := range record.Change.NewImage {
			if value.DataType() == events.DataTypeString {
				fmt.Printf("Attribute name: %s, value: %s\n", name, value.String())
			}
		}
	}
}
