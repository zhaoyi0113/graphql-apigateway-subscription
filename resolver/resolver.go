package resolver

import (
	"context"
	"fmt"
	"time"

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
	msg   string     `json:"msg"`
	id    graphql.ID `json:"id"`
	topic string     `json:"topic"`
}

type Subscriber struct {
	// stop   <-chan struct{}
	topic        string
	events       chan<- *MessageEvent
	connectionId string
}

func NewResolver(ctx context.Context) *Resolver {
	r := &Resolver{
		event:        make(chan *MessageEvent),
		subscribers:  make(chan *Subscriber),
		connectionDb: handler.NewConnectionDb(ctx),
	}
	go r.broadcastChat()
	return r
}

func (r *Resolver) GetChat(ctx context.Context, args GetChatArgs) *MessageEvent {
	msg := MessageEvent{
		msg:   "Hello world!",
		topic: "xxx",
		id:    "aa",
	}
	return &msg
}

type SendChatArgs struct {
	Message string
	Topic   string
}

var id = "1b1404d7-5c2b-4a14-bf9e-8bdc494e7234"

func (r *Resolver) SendChat(ctx context.Context, args SendChatArgs) graphql.ID {
	fmt.Println("send chat mutation", args.Message, args.Topic)
	id := graphql.ID(id)
	message := MessageEvent{msg: args.Message, topic: args.Topic}
	r.connectionDb.SaveEvent(args.Topic, args.Message)
	r.event <- &message
	return id
}

func (r *Resolver) Event(ctx context.Context, args *struct {
	Topic string
}) chan *MessageEvent {
	fmt.Println("resolver on event", args.Topic)
	fmt.Println("connectoin id:", ctx.Value(handler.ConnectId{}))
	connId := ctx.Value(handler.ConnectId{}).(string)
	eventId := ctx.Value(handler.EventId{}).(string)
	ch := make(chan *MessageEvent)
	go func() {
		r.connectionDb.SaveSubscriber(connId, args.Topic, eventId)
		ch <- &MessageEvent{}
	}()
	return ch
}

func (r *Resolver) broadcastChat() {
	subscribers := map[string]*Subscriber{}
	for {
		select {
		case s := <-r.subscribers:
			fmt.Println("add a subscriber", s)
			subscribers[uuid.New().String()] = s
		case e := <-r.event:
			fmt.Println("publish event", e)
			time.Sleep(3 * time.Second)
			items := r.connectionDb.GetSubscribers(e.topic)
			fmt.Println("Get subscriber:", items)
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
