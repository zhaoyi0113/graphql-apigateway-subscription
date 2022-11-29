package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Resolver struct{}

type GetChatArgs struct {
	Id graphql.ID
}

func (r *Resolver) GetChat(ctx context.Context, args GetChatArgs) string { return "Hello, world!" }

type SendChatArgs struct {
	Message string
}

func (r *Resolver) SendChat(ctx context.Context, args SendChatArgs) graphql.ID {
	return graphql.ID("1b1404d7-5c2b-4a14-bf9e-8bdc494e7234")
}
