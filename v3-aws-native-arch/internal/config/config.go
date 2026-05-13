package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Database  DatabaseConfig
	Server    ServerConfig
	Email     EmailConfig
	App       AppConfig
	Security  SecurityConfig
	Redis     RedisConfig
	S3        S3Config
	Messaging MessagingConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
	Env  string
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider    string // "smtp", "gmail", "custom", "brevo_api" - just for reference, not enforced
	Host        string // SMTP server host (e.g., smtp.gmail.com, localhost)
	Port        int    // SMTP port (587 for TLS, 465 for SSL, 25 for plain)
	Username    string // SMTP username (leave empty if no auth required)
	Password    string // SMTP password (leave empty if no auth required)
	FromEmail   string // Sender email address
	FromName    string // Sender display name
	UseTLS      bool   // Use STARTTLS (port 587)
	UseSSL      bool   // Use SSL/TLS (port 465)
	Timeout     int    // Connection timeout in seconds
	MaxRetries  int    // Number of retry attempts on failure
	TestMode    bool   // If true, emails are logged but not sent
	BrevoAPIKey string // Brevo API key (for cloud deployments)
}

// AppConfig holds general app configuration
type AppConfig struct {
	Name        string
	Environment string
	BaseURL     string
	FrontendURL string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey string // Must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string // Redis server address (e.g., localhost:6379)
	Password string // Redis password (leave empty if no auth)
	DB       int    // Redis database number
	Enabled  bool   // Enable/disable Redis
}

// S3Config holds AWS S3 configuration
type S3Config struct {
	AccessKey string // AWS access key ID
	SecretKey string // AWS secret access key
	Region    string // AWS region (e.g., us-east-1)
	Bucket    string // S3 bucket name
	PublicURL string // Public URL for accessing files
	LocalPath string // Fallback local storage path
	Enabled   bool   // Enable/disable S3
}

// MessagingConfig holds SNS topic ARNs and SQS queue URLs for the AWS-native
// messaging layer that replaces self-managed Kafka.
type MessagingConfig struct {
	// TopicARNs maps each topic name constant to its SNS ARN.
	// e.g. "order.created" → "arn:aws:sns:us-east-1:123456789012:order-created"
	TopicARNs map[string]string

	// SQS queue URLs — one per worker consumer group.
	PaymentWorkerQueue           string
	TicketWorkerQueue            string
	TicketEmailWorkerQueue       string
	EmailWorkerQueue             string
	NotificationWorkerQueue      string
	EventCancellationWorkerQueue string
	RefundProcessorWorkerQueue   string
	WaitlistWorkerQueue          string

	// Analytics worker subscribes to six separate queues (one per topic).
	AnalyticsOrderCreatedQueue     string
	AnalyticsOrderConfirmedQueue   string
	AnalyticsPaymentCompletedQueue string
	AnalyticsOrderCancelledQueue   string
	AnalyticsEventCancelledQueue   string
	AnalyticsTicketsGeneratedQueue string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "ticketing_system"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Email: EmailConfig{
			Provider:    getEnv("EMAIL_PROVIDER", "smtp"),
			Host:        getEnv("EMAIL_HOST", "localhost"),
			Port:        getEnvAsInt("EMAIL_PORT", 587),
			Username:    getEnv("EMAIL_USERNAME", ""),
			Password:    getEnv("EMAIL_PASSWORD", ""),
			FromEmail:   getEnv("EMAIL_FROM", "noreply@ticketing.com"),
			FromName:    getEnv("EMAIL_FROM_NAME", "Ticketing System"),
			UseTLS:      getEnvAsBool("EMAIL_USE_TLS", true),
			UseSSL:      getEnvAsBool("EMAIL_USE_SSL", false),
			Timeout:     getEnvAsInt("EMAIL_TIMEOUT", 30),
			MaxRetries:  getEnvAsInt("EMAIL_MAX_RETRIES", 3),
			TestMode:    getEnvAsBool("EMAIL_TEST_MODE", true),
			BrevoAPIKey: getEnv("BREVO_API_KEY", ""),
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Ticketing System"),
			Environment: getEnv("APP_ENV", "development"),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", "dev-key-32-bytes-length-aes!!123"), // Default 32-byte key for development
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", false),
		},
		S3: S3Config{
			AccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:    getEnv("AWS_REGION", "us-east-1"),
			Bucket:    getEnv("S3_BUCKET", ""),
			PublicURL: getEnv("S3_PUBLIC_URL", "http://localhost:8080/uploads"),
			LocalPath: getEnv("LOCAL_STORAGE_PATH", "./uploads"),
			Enabled:   getEnvAsBool("S3_ENABLED", false),
		},
		Messaging: MessagingConfig{
			TopicARNs: map[string]string{
				"order.created":               getEnv("SNS_ARN_ORDER_CREATED", ""),
				"order.payment_confirmed":      getEnv("SNS_ARN_ORDER_PAYMENT_CONFIRMED", ""),
				"order.confirmed":             getEnv("SNS_ARN_ORDER_CONFIRMED", ""),
				"order.cancelled":             getEnv("SNS_ARN_ORDER_CANCELLED", ""),
				"order.completed":             getEnv("SNS_ARN_ORDER_COMPLETED", ""),
				"reservation.expired":         getEnv("SNS_ARN_RESERVATION_EXPIRED", ""),
				"inventory.released":          getEnv("SNS_ARN_INVENTORY_RELEASED", ""),
				"event.cancelled":             getEnv("SNS_ARN_EVENT_CANCELLED", ""),
				"payment.completed":           getEnv("SNS_ARN_PAYMENT_COMPLETED", ""),
				"tickets.generated":           getEnv("SNS_ARN_TICKETS_GENERATED", ""),
				"notification.email.requested": getEnv("SNS_ARN_NOTIFICATION_EMAIL", ""),
				"refund.approved":             getEnv("SNS_ARN_REFUND_APPROVED", ""),
				"refund.completed":            getEnv("SNS_ARN_REFUND_COMPLETED", ""),
			},
			PaymentWorkerQueue:             getEnv("SQS_URL_PAYMENT_WORKER", ""),
			TicketWorkerQueue:              getEnv("SQS_URL_TICKET_WORKER", ""),
			TicketEmailWorkerQueue:         getEnv("SQS_URL_TICKET_EMAIL_WORKER", ""),
			EmailWorkerQueue:               getEnv("SQS_URL_EMAIL_WORKER", ""),
			NotificationWorkerQueue:        getEnv("SQS_URL_NOTIFICATION_WORKER", ""),
			EventCancellationWorkerQueue:   getEnv("SQS_URL_EVENT_CANCELLATION_WORKER", ""),
			RefundProcessorWorkerQueue:     getEnv("SQS_URL_REFUND_PROCESSOR_WORKER", ""),
			WaitlistWorkerQueue:            getEnv("SQS_URL_WAITLIST_WORKER", ""),
			AnalyticsOrderCreatedQueue:     getEnv("SQS_URL_ANALYTICS_ORDER_CREATED", ""),
			AnalyticsOrderConfirmedQueue:   getEnv("SQS_URL_ANALYTICS_ORDER_CONFIRMED", ""),
			AnalyticsPaymentCompletedQueue: getEnv("SQS_URL_ANALYTICS_PAYMENT_COMPLETED", ""),
			AnalyticsOrderCancelledQueue:   getEnv("SQS_URL_ANALYTICS_ORDER_CANCELLED", ""),
			AnalyticsEventCancelledQueue:   getEnv("SQS_URL_ANALYTICS_EVENT_CANCELLED", ""),
			AnalyticsTicketsGeneratedQueue: getEnv("SQS_URL_ANALYTICS_TICKETS_GENERATED", ""),
		},
	}
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Print("failed to load env variables")
	}

	// Validate encryption key length

	keyLen := len(config.Security.EncryptionKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 16, 24, or 32 bytes (current: %d bytes)", keyLen)
	}

	// Email configuration validation (optional - some SMTP servers don't require auth)
	// Uncomment if you want to enforce authentication:
	// if config.Email.Username == "" || config.Email.Password == "" {
	//     return nil, fmt.Errorf("email configuration is incomplete: username and password are required")
	// }

	return config, nil
}

// LoadOrPanic loads configuration or panics
func LoadOrPanic() *Config {
	config, err := Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
	return config
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if running in development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
