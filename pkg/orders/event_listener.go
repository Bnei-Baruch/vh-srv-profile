package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

const (
	ComponentAPI              = "api"
	ComponentRobokasaImporter = "robokasa_importer"

	TypeCreateAccount     = "create_account"
	TypeUpdateAccount     = "update_account"
	TypeDeleteAccount     = "delete_account"
	TypeHardDeleteAccount = "hard_delete_account"
	TypeCreateOrder       = "create_order"
	TypeUpdateOrder       = "update_order"
	TypeDeleteOrder       = "delete_order"
	TypeCreatePayment     = "create_payment"
	TypeUpdatePayment     = "update_payment"
	TypeDeletePayment     = "delete_payment"
	TypeDeleteSpecial     = "delete_special"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Component string                 `json:"component"`
	Actor     string                 `json:"actor"`
	Payload   map[string]interface{} `json:"payload"`
}

type EventHandler func(Event)

type EventListener struct {
	nc          *nats.Conn
	js          jetstream.JetStream
	consumer    jetstream.Consumer
	consumerCtx jetstream.ConsumeContext

	queue    chan Event
	handlers []EventHandler
}

func NewEventListener() (*EventListener, error) {
	el := new(EventListener)

	var err error
	el.nc, err = nats.Connect(common.Config.NatsUrl)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	el.js, err = jetstream.New(el.nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream.New: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	el.consumer, err = el.js.CreateOrUpdateConsumer(ctx, "VH_SRV_ORDERS", jetstream.ConsumerConfig{
		Name:        common.ServiceName,
		Durable:     common.ServiceName,
		Description: "Events listener of vh-srv-profile",
	})
	if err != nil {
		return nil, fmt.Errorf("jetstream.CreateOrUpdateConsumer: %w", err)
	}

	el.queue = make(chan Event, 2^10)
	el.handlers = make([]EventHandler, 0)
	return el, nil
}

func (el *EventListener) Run() error {
	var err error
	el.consumerCtx, err = el.consumer.Consume(el.handleMessage)
	if err != nil {
		return fmt.Errorf("jetstream consumer.Consume: %w", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("ERROR: EventListener runner goroutine panic: %v\n", r)
				debug.PrintStack()

				el.consumerCtx.Stop()
				if err := el.Run(); err != nil {
					log.Fatalf("EventListener runner goroutine error re-run after panic: %v\n", err)
				}
			}
		}()

		for event := range el.queue {
			for _, handler := range el.handlers {
				handler(event)
			}
		}
		log.Println("DEBUG: EventListener runner goroutine exit")
	}()

	return nil
}

func (el *EventListener) Close() {
	el.consumerCtx.Stop()
	el.nc.Close()
	close(el.queue)
}

func (el *EventListener) RegisterHandler(handler EventHandler) {
	el.handlers = append(el.handlers, handler)
}

func (el *EventListener) handleMessage(msg jetstream.Msg) {
	log.Printf("DEBUG: EventListener.handleMessage: %s\n", msg.Data())

	var event Event
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		log.Printf("ERROR: EventListener.handleMessage json.Unmarshal: %v \n", err)
	}

	el.queue <- event

	msg.Ack()
}
