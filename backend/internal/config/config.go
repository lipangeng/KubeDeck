package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Auth     AuthConfig     `json:"auth" yaml:"auth"`
	K8s      K8sConfig      `json:"kubernetes" yaml:"kubernetes"`
	AI       AIConfig       `json:"ai" yaml:"ai"`
	Logging  LoggingConfig  `json:"logging" yaml:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	TLS          TLSConfig     `json:"tls" yaml:"tls"`
	CORS         CORSConfig    `json:"cors" yaml:"cors"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string `json:"type" yaml:"type"` // sqlite, mysql, postgres
	DSN             string `json:"dsn" yaml:"dsn"`   // Connection string
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret       string                 `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiry       time.Duration          `json:"jwt_expiry" yaml:"jwt_expiry"`
	EncryptionKey   string                 `json:"encryption_key" yaml:"encryption_key"`
	EnableOAuth2    bool                   `json:"enable_oauth2" yaml:"enable_oauth2"`
	OAuth2Providers []OAuth2ProviderConfig `json:"oauth2_providers" yaml:"oauth2_providers"`
}

// OAuth2ProviderConfig holds OAuth2 provider configuration
type OAuth2ProviderConfig struct {
	Name         string   `json:"name" yaml:"name"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
	IssuerURL    string   `json:"issuer_url" yaml:"issuer_url"`
}

// K8sConfig holds Kubernetes configuration
type K8sConfig struct {
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	InCluster  bool   `json:"in_cluster" yaml:"in_cluster"`
}

// AIConfig holds AI configuration
type AIConfig struct {
	Enabled   bool               `json:"enabled" yaml:"enabled"`
	Providers []AIProviderConfig `json:"providers" yaml:"providers"`
}

// AIProviderConfig holds AI provider configuration
type AIProviderConfig struct {
	Type    string `json:"type" yaml:"type"` // ollama, openai
	BaseURL string `json:"base_url" yaml:"base_url"`
	APIKey  string `json:"api_key" yaml:"api_key"`
	Model   string `json:"model" yaml:"model"`
	Timeout int    `json:"timeout" yaml:"timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level" yaml:"level"`   // debug, info, warn, error
	Format     string `json:"format" yaml:"format"` // json, text
	OutputFile string `json:"output_file" yaml:"output_file"`
}

// Manager manages application configuration
type Manager struct {
	mu       sync.RWMutex
	config   *Config
	watchers []func(*Config)
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
			},
		},
		Database: DatabaseConfig{
			Type:            "sqlite",
			DSN:             "data/kubedeck.db",
			MaxIdleConns:    5,
			MaxOpenConns:    25,
			ConnMaxLifetime: "5m",
		},
		Auth: AuthConfig{
			JWTExpiry:    24 * time.Hour,
			EnableOAuth2: false,
		},
		K8s: K8sConfig{
			InCluster: true,
			Namespace: "default",
		},
		AI: AIConfig{
			Enabled: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{
		config: Default(),
	}
}

// Load loads configuration from file
func (m *Manager) Load(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Use defaults if file doesn't exist
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format from extension
	var config Config
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	} else if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported config file format: %s", path)
	}

	// Merge with defaults
	m.mergeConfig(&config)

	return nil
}

// mergeConfig merges loaded config with defaults
func (m *Manager) mergeConfig(loaded *Config) {
	// Server
	if loaded.Server.Host != "" {
		m.config.Server.Host = loaded.Server.Host
	}
	if loaded.Server.Port != 0 {
		m.config.Server.Port = loaded.Server.Port
	}
	if loaded.Server.ReadTimeout != 0 {
		m.config.Server.ReadTimeout = loaded.Server.ReadTimeout
	}
	if loaded.Server.WriteTimeout != 0 {
		m.config.Server.WriteTimeout = loaded.Server.WriteTimeout
	}

	// Database
	if loaded.Database.Type != "" {
		m.config.Database.Type = loaded.Database.Type
	}
	if loaded.Database.DSN != "" {
		m.config.Database.DSN = loaded.Database.DSN
	}

	// Auth
	if loaded.Auth.JWTSecret != "" {
		m.config.Auth.JWTSecret = loaded.Auth.JWTSecret
	}
	if loaded.Auth.EncryptionKey != "" {
		m.config.Auth.EncryptionKey = loaded.Auth.EncryptionKey
	}
	m.config.Auth.EnableOAuth2 = loaded.Auth.EnableOAuth2
	if len(loaded.Auth.OAuth2Providers) > 0 {
		m.config.Auth.OAuth2Providers = loaded.Auth.OAuth2Providers
	}

	// K8s
	if loaded.K8s.Kubeconfig != "" {
		m.config.K8s.Kubeconfig = loaded.K8s.Kubeconfig
	}
	if loaded.K8s.Namespace != "" {
		m.config.K8s.Namespace = loaded.K8s.Namespace
	}
	m.config.K8s.InCluster = loaded.K8s.InCluster

	// AI
	m.config.AI.Enabled = loaded.AI.Enabled
	if len(loaded.AI.Providers) > 0 {
		m.config.AI.Providers = loaded.AI.Providers
	}

	// Logging
	if loaded.Logging.Level != "" {
		m.config.Logging.Level = loaded.Logging.Level
	}
	if loaded.Logging.Format != "" {
		m.config.Logging.Format = loaded.Logging.Format
	}
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Update updates configuration and notifies watchers
func (m *Manager) Update(config *Config) {
	m.mu.Lock()
	m.config = config
	watchers := make([]func(*Config), len(m.watchers))
	copy(watchers, m.watchers)
	m.mu.Unlock()

	// Notify watchers
	for _, watcher := range watchers {
		go watcher(config)
	}
}

// Watch registers a configuration change watcher
func (m *Manager) Watch(fn func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watchers = append(m.watchers, fn)
}

// Global manager instance
var global *Manager
var globalOnce sync.Once

// Global returns the global config manager
func Global() *Manager {
	globalOnce.Do(func() {
		global = NewManager()
	})
	return global
}

// GetConfig returns global configuration
func GetConfig() *Config {
	return Global().Get()
}

// LoadConfig loads configuration from file into global manager
func LoadConfig(path string) error {
	return Global().Load(path)
}
