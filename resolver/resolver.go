package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Resolver struct{ message Message }

type GetChatArgs struct {
	Id graphql.ID
}

type Message struct {
	Message string     `json:"message"`
	Id      graphql.ID `json:"id"`
}

func (r *Resolver) GetChat(ctx context.Context, args GetChatArgs) string { return "Hello, world!" }

type SendChatArgs struct {
	Message string
}

var id = "1b1404d7-5c2b-4a14-bf9e-8bdc494e7234"

func (r *Resolver) SendChat(ctx context.Context, args SendChatArgs) graphql.ID {
	return graphql.ID(id)
}

func (r *Resolver) Event(ctx context.Context, args *struct {
	On string
	// }) (string, error) {
}) (<-chan *MessageResolve, error) {
	println("on event", args.On)
	ch := make(chan *MessageResolve, 1)
	s := ""
	msg := Message{s, graphql.ID(id)}
	msgArgs := MessageResolve{message: msg}
	ch <- &msgArgs
	return ch, nil
}
