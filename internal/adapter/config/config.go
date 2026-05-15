package config

import (
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Http      *HttpConfig
	DynamoDB  *DynamoDBConfig
	SNS       *SNSConfig
	SQS       *SQSConfig
	Clients   *HTTPClientsConfig
	DogStatsD *DogStatsDConfig
	Env       string
	Version   string
	Service   string
}

type HttpConfig struct {
	AllowedOrigins []string
	Port           string
	Url            string
	ApiUrl         string
	PublicBaseURL  string
}

type DynamoDBConfig struct {
	Region    string
	Endpoint  string
	TableName string
}

type SNSConfig struct {
	OrdersTopicARN string
}

type SQSConfig struct {
	StockAvailableURL   string
	StockUnavailableURL string
	StockUpdatedURL     string
	PaymentEventsURL    string
}

type HTTPClientsConfig struct {
	UsersBaseURL string
	StockBaseURL string
	TimeoutMs    int
}

type DogStatsDConfig struct {
	Addr     string
	Disabled bool
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := getEnv(key, "")
	if v == "" {
		return defaultValue
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return defaultValue
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func Init() (*Config, error) {
	_ = godotenv.Load()

	var allowedOriginList []string
	allowedOrigins := getEnv("HTTP_ALLOWED_ORIGINS", "*")
	if allowedOrigins != "*" && len(allowedOrigins) > 0 {
		allowedOriginList = strings.Split(allowedOrigins, ",")
	}

	http := &HttpConfig{
		AllowedOrigins: allowedOriginList,
		Port:           getEnv("HTTP_PORT", "8080"),
		Url:            getEnv("HTTP_URL", "http://localhost"),
		ApiUrl:         getEnv("API_URL", "http://localhost"),
		PublicBaseURL:  getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
	}

	dynamoDB := &DynamoDBConfig{
		Region:    getEnv("AWS_REGION", "us-east-1"),
		Endpoint:  getEnv("DYNAMODB_ENDPOINT", ""),
		TableName: getEnv("DYNAMODB_TABLE", "orders"),
	}

	sns := &SNSConfig{
		OrdersTopicARN: getEnv("ORDERS_TOPIC_ARN", ""),
	}

	sqs := &SQSConfig{
		StockAvailableURL:   getEnv("SQS_STOCK_AVAILABLE_URL", ""),
		StockUnavailableURL: getEnv("SQS_STOCK_UNAVAILABLE_URL", ""),
		StockUpdatedURL:     getEnv("SQS_STOCK_UPDATED_URL", ""),
		PaymentEventsURL:    getEnv("SQS_PAYMENT_EVENTS_URL", ""),
	}

	clients := &HTTPClientsConfig{
		UsersBaseURL: getEnv("USERS_BASE_URL", "http://localhost:8081"),
		StockBaseURL: getEnv("STOCK_BASE_URL", "http://localhost:8082"),
		TimeoutMs:    getEnvInt("HTTP_CLIENT_TIMEOUT_MS", 2000),
	}

	agentHost := getEnv("DD_AGENT_HOST", "")
	dogstatsdAddr := ""
	dogstatsdDisabled := true
	if agentHost != "" {
		dogstatsdAddr = net.JoinHostPort(agentHost, "8125")
		dogstatsdDisabled = false
	}

	return &Config{
		Http:     http,
		DynamoDB: dynamoDB,
		SNS:      sns,
		SQS:      sqs,
		Clients:  clients,
		Env:      getEnv("ENV", "development"),
		Version:  getEnv("API_VERSION", "1.0.0"),
		Service:  getEnv("DD_SERVICE", "tech-challenge-api"),
		DogStatsD: &DogStatsDConfig{
			Addr:     dogstatsdAddr,
			Disabled: dogstatsdDisabled,
		},
	}, nil
}
