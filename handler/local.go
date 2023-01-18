package handler

import (
	"context"
	"fmt"
	"time"
)

func SetupLocalEnv(h *Handler) {
	// event := GraphqlWSEvent{
	// 	Id:   "001",
	// 	Type: "start",
	// 	Payload: GraphqlQuery{
	// 		OperationName: "event",
	// 		Query:         "subscription event { event(on: \"xxxx\" ) { msg }}",
	// 	},
	// }
	// e, _ := json.Marshal(event)
	// response, _ := h.GraphqlDefaultHandler(context.Background(), events.APIGatewayWebsocketProxyRequest{
	// 	RequestContext: events.APIGatewayWebsocketProxyRequestContext{
	// 		EventType:    "MESSAGE",
	// 		ConnectionID: "1",
	// 	},
	// 	Body: string(e),
	// })
	// fmt.Println(response)

	// time.Sleep(3 * time.Second)

	fmt.Println("sendChat mutation")
	res := h.Exec(context.TODO(), "sendChat", "mutation sendChat{\n sendChat(topic: \"1\", message: \"hello\") }\n", nil)
	fmt.Println("response:", res)
	time.Sleep(50 * time.Second)
}
