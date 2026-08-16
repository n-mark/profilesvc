package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	ServerAddr string
	BrokerType string
}

type RabbitConfig struct {
	DSN                                 string
	ProfileProduceExchange              string
	BillingConsumeExchange              string
	ProfileSvcConsumerForBillingAccount string
	BillingExchangeToConsumerRoutingKey string
	ProfileCreatedRoutingKey            string
}

// KafkaConfig mirrors RabbitConfig: exchanges become topics, routing keys
// become `event_type` values inside the payload, and the consumer queue
// becomes a consumer group.
type KafkaConfig struct {
	Brokers                        []string
	ProfileTopic                   string
	BillingTopic                   string
	BillingGroup                   string
	ProfileCreatedEventType        string
	BillingAccountCreatedEventType string
}

func Load() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "social"),
		DBUser:     getEnv("DB_USER", "social_user"),
		DBPassword: getEnv("DB_PASSWORD", "social_pass"),
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		BrokerType: getEnv("BROKER_TYPE", "KAFKA"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetRabbitConfig() RabbitConfig {
	user := os.Getenv("RABBIT_USERNAME")
	password := os.Getenv("RABBIT_PASSWORD")
	host := os.Getenv("RABBIT_HOST")
	port := os.Getenv("RABBIT_PORT")
	billingExchange := getEnv("RABBIT_BILLING_CONSUME_EXCHANGE", "billing")
	profileExchange := getEnv("RABBIT_PROFILE_PRODUCE_EXCHANGE", "profile")
	consumer := getEnv("RABBIT_PROFILESVC_BILLING_CONSUMER", "profilesvc.consumer.for.billing.account")
	rk := getEnv("RABBIT_PROFILESVC_BILLING_RK", "billing.account.created")
	profileCreatedRk := getEnv("RABBIT_PROFILE_CREATED_RK", "profile.created")

	u := url.URL{Scheme: "amqp",
		User: url.UserPassword(user, password),
		Host: fmt.Sprintf("%s:%s", host, port)}

	return RabbitConfig{DSN: u.String(),
		ProfileProduceExchange:              profileExchange,
		ProfileCreatedRoutingKey:            profileCreatedRk,
		BillingConsumeExchange:              billingExchange,
		ProfileSvcConsumerForBillingAccount: consumer,
		BillingExchangeToConsumerRoutingKey: rk}
}

func GetKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:                        strings.Split(getEnv("KAFKA_BROKERS", "kafka-1:9092,kafka-2:9092,kafka-3:9092"), ","),
		ProfileTopic:                   getEnv("KAFKA_PROFILE_TOPIC", "profile"),
		BillingTopic:                   getEnv("KAFKA_BILLING_TOPIC", "billing"),
		BillingGroup:                   getEnv("KAFKA_BILLING_GROUP", "profilesvc.billing"),
		ProfileCreatedEventType:        getEnv("KAFKA_PROFILE_CREATED_EVENT_TYPE", "profile.created"),
		BillingAccountCreatedEventType: getEnv("KAFKA_BILLING_ACCOUNT_CREATED_EVENT_TYPE", "billing.account.created"),
	}
}
