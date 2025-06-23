package config

import (
	"os"
)

type Config struct {
	OrderServiceURL   string
	PaymentServiceURL string
	Port              string
}

func NewConfig() *Config {
	return &Config{
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://order-service:8081"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8082"),
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
