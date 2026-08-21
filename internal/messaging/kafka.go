package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"

	"profile-service/internal/config"
	"profile-service/internal/models"
)

type consumer struct {
	source  Source
	handler HandlerFunc
}

// KafkaImpl is the Kafka counterpart of RabbitImpl.
//
// Topology mapping: exchange -> topic (`profile` for produce, `billing` for
// consume), routing key -> `event_type` inside the payload (filtered on the
// client side, see Source.EventTypes), queue -> consumer group.
type KafkaImpl struct {
	cfg       config.KafkaConfig
	writer    *kafka.Writer
	consumers []consumer
}

func NewKafkaImpl(cfg config.KafkaConfig) *KafkaImpl {
	w := &kafka.Writer{
		Addr: kafka.TCP(cfg.Brokers...),
		// Same key -> same partition -> per-user ordering is preserved.
		Balancer: &kafka.Hash{},
		// acks=all + min.insync.replicas=2 => no data loss if one broker dies.
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		BatchTimeout:           50 * time.Millisecond,
	}
	return &KafkaImpl{cfg: cfg, writer: w}
}

func (k *KafkaImpl) GetBillingResponseSource() Source {
	return Source{
		Name:       k.cfg.BillingTopic,
		Group:      k.cfg.BillingGroup,
		EventTypes: []string{k.cfg.BillingAccountCreatedEventType},
	}
}

func (k *KafkaImpl) RegisterConsumer(s Source, h HandlerFunc) {
	k.consumers = append(k.consumers, consumer{source: s, handler: h})
}

func (k *KafkaImpl) ReportProfileCreated(e models.ProfileCreatedEvent) error {
	if e.EventType == "" {
		e.EventType = k.cfg.ProfileCreatedEventType
	}

	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal profile.created: %w", err)
	}

	slog.Info("publishing profile.created",
		"topic", k.cfg.ProfileTopic, "user_id", e.UserId, "event_type", e.EventType)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: k.cfg.ProfileTopic,
		Key:   []byte(strconv.FormatInt(e.UserId, 10)),
		Value: body,
	})
}

func (k *KafkaImpl) Run() {
	defer k.writer.Close()

	wg := &sync.WaitGroup{}
	for _, c := range k.consumers {
		wg.Add(1)
		go k.runConsumer(c, wg)
	}
	wg.Wait()
}

func (k *KafkaImpl) runConsumer(c consumer, wg *sync.WaitGroup) {
	defer wg.Done()

	const reconnectBackoff = 2 * time.Second

	for {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: k.cfg.Brokers,
			Topic:   c.source.Name,
			GroupID: c.source.Group,
			// Manual commits: commit only after the handler succeeded, so a crash
			// leads to redelivery instead of a lost message (at-least-once).
			CommitInterval: 0,
			MaxWait:        time.Second,
		})

		slog.Info("kafka consumer started",
			"topic", c.source.Name, "group", c.source.Group, "event_types", c.source.EventTypes)

		// consumeLoop only returns on a fatal read error (e.g. broker restart).
		// Recreate the reader so the consumer keeps working without a rollout
		// restart, instead of letting the goroutine die forever.
		consumeLoop(r, c)
		_ = r.Close()

		slog.Warn("kafka consumer reconnecting", "topic", c.source.Name, "group", c.source.Group)
		time.Sleep(reconnectBackoff)
	}
}

func consumeLoop(r *kafka.Reader, c consumer) {
	for {
		msg, err := r.FetchMessage(context.Background())
		if err != nil {
			slog.Error("fetch message", "topic", c.source.Name, "err", err)
			return
		}

		if !matchesEventType(msg.Value, c.source.EventTypes) {
			// Not ours (the topic carries several event types): commit so we
			// do not re-read it forever.
			if err := r.CommitMessages(context.Background(), msg); err != nil {
				slog.Error("commit filtered message", "topic", c.source.Name, "err", err)
			}
			continue
		}

		ok, err := c.handler(msg.Value)
		if err != nil {
			// No commit -> redelivery. Handlers must be idempotent.
			slog.Error("handle message", "topic", c.source.Name, "offset", msg.Offset, "err", err)
			continue
		}
		if !ok {
			slog.Debug("message skipped", "topic", c.source.Name, "offset", msg.Offset)
		}

		if err := r.CommitMessages(context.Background(), msg); err != nil {
			slog.Error("commit message", "topic", c.source.Name, "offset", msg.Offset, "err", err)
		}
	}
}

// matchesEventType is the client-side replacement for AMQP routing-key
// wildcards (e.g. `billing.account.created`).
func matchesEventType(body []byte, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}

	var envelope struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		// Malformed payloads are handed to the handler, which reports the error.
		return true
	}

	for _, a := range allowed {
		if a == envelope.EventType {
			return true
		}
	}
	return false
}
