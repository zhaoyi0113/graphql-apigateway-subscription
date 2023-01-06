package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func SetupLocalEnv(h *Handler) {
	ch := h.Subscribe(context.TODO(), "event", "subscription event {\n  event(on: \"xxxx\") {\n    msg\n    __typename\n  }\n}", nil)
	go func() {
		for resp := range ch {
			j, _ := json.Marshal(resp)
			fmt.Println("Receive published message", string(j))
		}
	}()

	time.Sleep(3 * time.Second)
	h.Exec(context.TODO(), "sendChat", "mutation sendChat{\n sendChat(topic: \"1\", message: \"hello\") }\n", nil)
	time.Sleep(50 * time.Second)
}
