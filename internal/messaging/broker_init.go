package messaging

import (
	"fmt"
	"profile-svc/internal/config"
	"profile-svc/internal/models"
)

type Broker interface {
	RegisterConsumer(dataSourceName string, h HandlerFunc)
	Run()
	ReportProfileCreated(e models.ProfileCreatedEvent) error
	GetBillingResponseDataSourceName() string
}

func InitBroker(cfg config.Config) (Broker, error) {
	var b Broker

	if "RABBITMQ" == cfg.BrokerType {
		br, err := NewRabbitImpl(config.GetRabbitConfig())
		if err != nil {
			return nil, fmt.Errorf("can't init rabbitmq impl: %s", err)
		}
		b = br
	}

	return b, nil
}
