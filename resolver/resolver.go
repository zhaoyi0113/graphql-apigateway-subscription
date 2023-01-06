package resolver

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
)

type Resolver struct {
	event       chan *MessageEvent
	subscribers chan *Subscriber
}

type GetChatArgs struct {
	Id graphql.ID
}

type MessageEvent struct {
	msg   string     `json:"msg"`
	id    graphql.ID `json:"id"`
	topic string     `json:"topic"`
}

type Subscriber struct {
	// stop   <-chan struct{}
	topic  string
	events chan<- *MessageEvent
}

func NewResolver() *Resolver {
	r := &Resolver{
		event:       make(chan *MessageEvent),
		subscribers: make(chan *Subscriber),
	}
	go r.broadcastChat()
	return r
}

func (r *Resolver) GetChat(ctx context.Context, args GetChatArgs) string { return "Hello, world!" }

type SendChatArgs struct {
	Message string
	Topic   string
}

var id = "1b1404d7-5c2b-4a14-bf9e-8bdc494e7234"

func (r *Resolver) SendChat(ctx context.Context, args SendChatArgs) graphql.ID {
	fmt.Println("send chat mutation")
	id := graphql.ID(id)
	message := MessageEvent{msg: args.Message, topic: args.Topic}
	r.event <- &message
	return id
}

func (r *Resolver) Event(ctx context.Context, args *struct {
	On string
}) <-chan *MessageEvent {
	fmt.Println("on event", args.On)
	ch := make(chan *MessageEvent)
	r.subscribers <- &Subscriber{events: ch, topic: args.On}
	return ch
}

func (r *Resolver) broadcastChat() {
	subscribers := map[string]*Subscriber{}
	for {
		select {
		case s := <-r.subscribers:
			fmt.Println("add a subscriber")
			subscribers[uuid.New().String()] = s
		case e := <-r.event:
			fmt.Println("publish event")
			for id, s := range subscribers {
				go func(id string, s *Subscriber) {
					select {
					case s.events <- e:
						fmt.Println("publish to event", e)
					default:
					}
				}(id, s)
			}
		}
	}
}

func (r *MessageEvent) Msg() string {
	return r.msg
}

func (r *MessageEvent) Id() graphql.ID {
	return r.id
}

func (r *MessageEvent) Topic() string {
	return r.topic
}
