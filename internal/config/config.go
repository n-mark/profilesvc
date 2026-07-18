package config

import (
	"fmt"
	"net/url"
	"os"
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

func Load() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "social"),
		DBUser:     getEnv("DB_USER", "social_user"),
		DBPassword: getEnv("DB_PASSWORD", "social_pass"),
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		BrokerType: getEnv("BROKER_TYPE", "RABBITMQ"),
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
