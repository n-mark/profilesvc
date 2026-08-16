package messaging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"profile-service/internal/config"
	"profile-service/internal/models"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(b []byte) (bool, error)

type RabbitImpl struct {
	conn          *amqp.Connection
	publisher     *amqp.Channel
	publisherLock sync.Mutex
	consumers     map[string]HandlerFunc
	cfg           config.RabbitConfig
}

func (r *RabbitImpl) GetBillingResponseSource() Source {
	// RabbitMQ filters server-side via the queue binding, so no EventTypes here.
	return Source{Name: r.cfg.ProfileSvcConsumerForBillingAccount}
}

func NewRabbitImpl(cfg config.RabbitConfig) (*RabbitImpl, error) {
	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	publisher, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}

	r := &RabbitImpl{conn: conn, consumers: make(map[string]HandlerFunc), cfg: cfg, publisher: publisher}

	return r, nil
}


func (r *RabbitImpl) declareExchange(ch *amqp.Channel, exchange string) error {
	return ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
}

func (r *RabbitImpl) declareQueueAndBind(ch *amqp.Channel, queue string) error {
	if _, err := ch.QueueDeclare(
		queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("declare queue %q: %w", queue, err)
	}

	if r.cfg.BillingConsumeExchange == "" {
		return nil
	}

	rks := r.routingKeyFor(queue)
	if len(rks) == 0 {
		return nil
	}

	for _, rk := range rks {
		if err := ch.QueueBind(queue, rk, r.cfg.BillingConsumeExchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %q to %q with rk %q: %w", queue, r.cfg.BillingConsumeExchange, rk, err)
		}
	}

	return nil
}

func (r *RabbitImpl) routingKeyFor(queue string) []string {
	switch queue {
	case r.cfg.ProfileSvcConsumerForBillingAccount:
		return []string{r.cfg.BillingExchangeToConsumerRoutingKey}
	}
	return []string{}
}

func (r *RabbitImpl) RegisterConsumer(s Source, h HandlerFunc) {
	r.consumers[s.Name] = h
}

func (r *RabbitImpl) produceProfileEvent(routingKey, messageId string, body []byte) error {
	r.publisherLock.Lock()
	defer r.publisherLock.Unlock()

	return r.publisher.Publish(
		r.cfg.ProfileProduceExchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    messageId,
			Body:         body,
		},
	)
}

func (r *RabbitImpl) ReportProfileCreated(e models.ProfileCreatedEvent) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal profile.created: %w", err)
	}
	slog.Info("publishing profile.created", "user_id", e.UserId, "rk", r.cfg.ProfileCreatedRoutingKey)
	return r.produceProfileEvent(r.cfg.ProfileCreatedRoutingKey, e.EventId.String(), body)
}

func (r *RabbitImpl) Run() {
	defer r.conn.Close()
	defer r.publisher.Close()

	exchangesToDeclare := []string{r.cfg.BillingConsumeExchange, r.cfg.ProfileProduceExchange}
	for _, exchange := range exchangesToDeclare {
		if err := r.declareExchange(r.publisher, exchange); err != nil {
			slog.Error("declare topology", "op", "exchange", "err", err)
			return
		}
	}

	for queue := range r.consumers {
		if err := r.declareQueueAndBind(r.publisher, queue); err != nil {
			slog.Error("declare topology", "queue", queue, "err", err)
			return
		}
		slog.Info("queue ready", "queue", queue)
	}

	wg := &sync.WaitGroup{}
	for k, v := range r.consumers {
		wg.Add(1)
		go r.runConsumer(k, v, wg)
	}
	wg.Wait()
}

func (r *RabbitImpl) runConsumer(queue string, handler HandlerFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	ch, err := r.conn.Channel()
	if err != nil {
		slog.Error("create channel", "err", err)
		return
	}

	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		slog.Error("set qos", "err", err)
		return
	}

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		slog.Error("consume message", "err", err)
		return
	}

	for msg := range msgs {
		ok, err := handler(msg.Body)
		if err != nil {
			msg.Nack(false, false)
			continue
		}

		if !ok {
			msg.Ack(false)
			continue
		}

		msg.Ack(false)
	}
}
