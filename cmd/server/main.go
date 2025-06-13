package main

import (
	"log"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"tsuribari/internal/config"
	"tsuribari/internal/handlers"
	"tsuribari/internal/middleware"
	"tsuribari/internal/queue"
	"tsuribari/internal/storage"
)

func extractHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func extractVhost(amqpURL string) string {
	u, err := url.Parse(amqpURL)
	if err != nil {
		return ""
	}

	vhost := strings.TrimPrefix(u.Path, "/")
	if vhost == "" {
		return "/"
	}
	return "/" + vhost
}

func setupLogging() {
	if os.Getenv("DEBUG") != "" {
		// Debug mode: log to console
		log.SetOutput(os.Stdout)
		log.SetPrefix("tsuribari: ")
		return
	}

	// Production mode: log to syslog
	syslogWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "tsuribari")
	if err != nil {
		// Fallback to console if syslog fails
		log.Printf("Failed to connect to syslog, falling back to console: %v", err)
		log.SetOutput(os.Stdout)
		log.SetPrefix("tsuribari: ")
		return
	}

	log.SetOutput(syslogWriter)
	log.SetFlags(0) // syslog handles timestamps
}

func main() {
	// Setup logging first
	setupLogging()

	// Set Gin to release mode if DEBUG is not set
	if os.Getenv("DEBUG") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize storage
	couchDB, err := storage.NewCouchDB(cfg.CouchDB.URL, cfg.CouchDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to CouchDB:", err)
	}
	log.Printf("Connected to CouchDB: %s/%s", extractHostname(cfg.CouchDB.URL), cfg.CouchDB.Database)

	// Initialize queue
	rabbitMQ, err := queue.NewRabbitMQ(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Queue)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	log.Printf("Connected to RabbitMQ: %s%s", extractHostname(cfg.RabbitMQ.URL), extractVhost(cfg.RabbitMQ.URL))
	defer rabbitMQ.Close()

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(couchDB, rabbitMQ)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Webhook endpoints with middleware
	webhookGroup := router.Group("/webhooks")
	webhookGroup.Use(middleware.IPFilter(cfg.Security.TrustedIPs))
	webhookGroup.Use(middleware.HMACValidator(cfg.Security.Secrets))
	{
		webhookGroup.POST("/:organisation", webhookHandler.HandleWebhook)
		webhookGroup.POST("/:organisation/:pipeline", webhookHandler.HandleWebhook)
	}

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	log.Fatal(router.Run(addr))
}
