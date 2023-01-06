package resolver

import "github.com/graph-gophers/graphql-go"

type MessageResolve struct {
	message MessageEvent
}

func (r *MessageResolve) Msg() string {
	return r.message.msg
}

func (r *MessageResolve) Id() graphql.ID {
	return r.message.id
}
