package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func SetupLocalEnv(h *Handler) {
	event := GraphqlWSEvent{
		Id:   "001",
		Type: "start",
		Payload: GraphqlQuery{
			OperationName: "event",
			Query:         "subscription event { event(on: \"xxxx\" ) { msg }}",
		},
	}
	e, _ := json.Marshal(event)
	response, _ := h.GraphqlDefaultHandler(context.Background(), events.APIGatewayWebsocketProxyRequest{
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			EventType:    "MESSAGE",
			ConnectionID: "1",
		},
		Body: string(e),
	})
	fmt.Println(response)
	// ch := h.Subscribe(context.TODO(), "event", "subscription event {\n  event(on: \"xxxx\") {\n    msg\n    __typename\n  }\n} ", nil)
	// go func() {
	// 	for resp := range ch {
	// 		j, _ := json.Marshal(resp)
	// 		fmt.Println("Receive published message", string(j))
	// 	}
	// }()

	time.Sleep(3 * time.Second)
	fmt.Println("sendChat mutation")
	h.Exec(context.TODO(), "sendChat", "mutation sendChat{\n sendChat(topic: \"1\", message: \"hello\") }\n", nil)
	time.Sleep(50 * time.Second)
}
