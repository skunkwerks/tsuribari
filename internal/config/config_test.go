package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Create a new viper instance for this test
	v := viper.New()

	// Set a non-existent config path to ensure defaults are used
	v.SetConfigName("nonexistent")
	v.AddConfigPath("/tmp/nonexistent/path/that/does/not/exist")

	// Set defaults on this instance
	v.SetDefault("server.port", "4003")
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("couchdb.url", "http://admin:passwd@127.0.0.1:5984")
	v.SetDefault("couchdb.database", "koans")
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq.exchange", "koans.topic")
	v.SetDefault("rabbitmq.queue", "koans.workflow")

	// Try to read config (should fail and use defaults)
	v.ReadInConfig() // Ignore error

	var config Config
	err := v.Unmarshal(&config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test defaults
	if config.Server.Port != "4003" {
		t.Errorf("Expected default port 4003, got %s", config.Server.Port)
	}

	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Expected default host 127.0.0.1, got %s", config.Server.Host)
	}

	if config.CouchDB.URL != "http://admin:passwd@127.0.0.1:5984" {
		t.Errorf("Expected default CouchDB URL, got %s", config.CouchDB.URL)
	}

	if config.CouchDB.Database != "koans" {
		t.Errorf("Expected default database koans, got %s", config.CouchDB.Database)
	}

	if config.RabbitMQ.URL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("Expected default RabbitMQ URL, got %s", config.RabbitMQ.URL)
	}

	if config.RabbitMQ.Exchange != "koans.topic" {
		t.Errorf("Expected default exchange koans.topic, got %s", config.RabbitMQ.Exchange)
	}

	if config.RabbitMQ.Queue != "koans.workflow" {
		t.Errorf("Expected default queue koans.workflow, got %s", config.RabbitMQ.Queue)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create a new viper instance for this test
	v := viper.New()

	// Create temporary config file
	configContent := `
server:
  host: "0.0.0.0"
  port: "8080"

couchdb:
  url: "http://test:test@localhost:5984"
  database: "test_db"

security:
  trusted_ips:
    - "192.168.1.0/24"
    - "10.0.0.0/8"
  secrets:
    testorg: "testsecret123"
    demo: "demosecret456"
`

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Configure this viper instance to use the temp file
	v.SetConfigFile(tmpFile.Name())

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Expected no error reading config, got %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Expected no error unmarshaling, got %v", err)
	}

	// Test loaded values
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Server.Host)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", config.Server.Port)
	}

	if config.CouchDB.URL != "http://test:test@localhost:5984" {
		t.Errorf("Expected test CouchDB URL, got %s", config.CouchDB.URL)
	}

	if config.CouchDB.Database != "test_db" {
		t.Errorf("Expected database test_db, got %s", config.CouchDB.Database)
	}

	if len(config.Security.TrustedIPs) != 2 {
		t.Errorf("Expected 2 trusted IPs, got %d", len(config.Security.TrustedIPs))
	}

	if len(config.Security.TrustedIPs) > 0 && config.Security.TrustedIPs[0] != "192.168.1.0/24" {
		t.Errorf("Expected first IP 192.168.1.0/24, got %s", config.Security.TrustedIPs[0])
	}

	if len(config.Security.Secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(config.Security.Secrets))
	}

	if config.Security.Secrets["testorg"] != "testsecret123" {
		t.Errorf("Expected testorg secret testsecret123, got %s", config.Security.Secrets["testorg"])
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create a new viper instance for this test
	v := viper.New()

	// Create temp file with invalid YAML
	invalidYAML := `
server:
  host: "0.0.0.0"
  port: 8080
invalid yaml structure
  missing: colon
`

	tmpFile, err := os.CreateTemp("", "invalid*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Configure this viper instance to use the invalid file
	v.SetConfigFile(tmpFile.Name())

	err = v.ReadInConfig()
	if err == nil {
		t.Error("Expected error with invalid YAML, got nil")
	}
}
