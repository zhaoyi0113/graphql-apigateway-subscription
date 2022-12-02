package resolver

import "github.com/graph-gophers/graphql-go"

type MessageResolve struct {
	message Message
}

func (r *MessageResolve) Message() string {
	return r.message.Message
}

func (r *MessageResolve) Id() graphql.ID {
	return r.message.Id
}
