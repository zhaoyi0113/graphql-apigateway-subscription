package resolver

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	handler "github.com/zhaoyi0113/graphql-apigateway-subscription/handler"
)

type Resolver struct {
	event        chan *MessageEvent
	subscribers  chan *Subscriber
	connectionDb *handler.ConnectionDb
}

type GetChatArgs struct {
	Id graphql.ID
}

type MessageEvent struct {
	msg          string     `json:"msg"`
	id           graphql.ID `json:"id"`
	topic        string     `json:"topic"`
	connectionId string     `json:"connectionId"`
}

type Subscriber struct {
	// stop   <-chan struct{}
	topic        string
	events       chan<- *MessageEvent
	connectionId string
}

func NewResolver() *Resolver {
	r := &Resolver{
		event:        make(chan *MessageEvent),
		subscribers:  make(chan *Subscriber),
		connectionDb: handler.NewConnectionDb(),
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
	connId := ctx.Value(handler.ConnectId{}).(string)
	fmt.Println("connectoin id:", ctx.Value(handler.ConnectId{}))
	id := graphql.ID(id)
	message := MessageEvent{msg: args.Message, topic: args.Topic, connectionId: connId}
	r.event <- &message
	return id
}

func (r *Resolver) Event(ctx context.Context, args *struct {
	On string
}) <-chan *MessageEvent {
	fmt.Println("resolver on event", args.On)
	fmt.Println("connectoin id:", ctx.Value(handler.ConnectId{}))
	connId := ctx.Value(handler.ConnectId{}).(string)
	ch := make(chan *MessageEvent)
	r.subscribers <- &Subscriber{events: ch, topic: args.On, connectionId: connId}
	return ch
}

func (r *Resolver) broadcastChat() {
	subscribers := map[string]*Subscriber{}
	for {
		select {
		case s := <-r.subscribers:
			fmt.Println("add a subscriber", s)
			subscribers[uuid.New().String()] = s
			r.connectionDb.SaveSubscriber(s.connectionId, s.topic)
		case e := <-r.event:
			fmt.Println("publish event", e)
			items := r.connectionDb.GetSubscribers(e.connectionId, e.topic)
			fmt.Println("Get subscribers:", items)
			for _, item := range items {
				fmt.Println("Get subscribers:", item)
			}
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
