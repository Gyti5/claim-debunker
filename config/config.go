package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env         string            `yaml:"env"`
	Server      ServerConfig      `yaml:"server"`
	Composition CompositionConfig `yaml:"composition"`
	OpenAI      OpenAIConfig      `yaml:"openai"`
	Security    SecurityConfig    `yaml:"security"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type CompositionConfig struct {
	MaxUploadBytes int64 `yaml:"max_upload_bytes"`
}

type OpenAIConfig struct {
	APIKey  string        `yaml:"api_key"`
	Model   string        `yaml:"model"`
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

type SecurityConfig struct {
	CompositionAPIKey             string `yaml:"composition_api_key"`
	CompositionRateLimitPerMinute int    `yaml:"composition_rate_limit_per_minute"`
}

func Load(path string) (Config, error) {
	cfg := Config{
		Env: "local",
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Composition: CompositionConfig{
			MaxUploadBytes: 10 << 20,
		},
		OpenAI: OpenAIConfig{
			APIKey:  "",
			Model:   "gpt-5.5",
			BaseURL: "https://api.openai.com/v1/responses",
			Timeout: 45 * time.Second,
		},
		Security: SecurityConfig{
			CompositionAPIKey:             "replace-with-random-secret",
			CompositionRateLimitPerMinute: 60,
		},
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	applyEnvOverrides(&cfg)

	if cfg.Server.Port <= 0 {
		return Config{}, fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Composition.MaxUploadBytes <= 0 {
		return Config{}, fmt.Errorf(
			"invalid composition max upload bytes: %d",
			cfg.Composition.MaxUploadBytes,
		)
	}
	if cfg.OpenAI.Model == "" {
		return Config{}, fmt.Errorf("openai model is required")
	}
	if cfg.OpenAI.BaseURL == "" {
		return Config{}, fmt.Errorf("openai base url is required")
	}
	if cfg.OpenAI.Timeout <= 0 {
		return Config{}, fmt.Errorf("openai timeout must be positive")
	}
	if cfg.Security.CompositionAPIKey == "" {
		return Config{}, fmt.Errorf("security composition api key is required")
	}
	if cfg.Security.CompositionRateLimitPerMinute <= 0 {
		return Config{}, fmt.Errorf("security composition rate limit per minute must be positive")
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	cfg.Env = getEnv("APP_ENV", cfg.Env)
	cfg.Server.Host = getEnv("SERVER_HOST", cfg.Server.Host)
	cfg.Server.Port = getEnvAsInt("SERVER_PORT", cfg.Server.Port)
	cfg.Composition.MaxUploadBytes = getEnvAsInt64("COMPOSITION_MAX_UPLOAD_BYTES", cfg.Composition.MaxUploadBytes)

	cfg.OpenAI.APIKey = getEnv("OPENAI_API_KEY", cfg.OpenAI.APIKey)
	cfg.OpenAI.Model = getEnv("OPENAI_MODEL", cfg.OpenAI.Model)
	cfg.OpenAI.BaseURL = getEnv("OPENAI_BASE_URL", cfg.OpenAI.BaseURL)
	cfg.OpenAI.Timeout = getEnvAsDuration("OPENAI_TIMEOUT", cfg.OpenAI.Timeout)

	cfg.Security.CompositionAPIKey = getEnv("COMPOSITION_API_KEY", cfg.Security.CompositionAPIKey)
	cfg.Security.CompositionRateLimitPerMinute = getEnvAsInt(
		"COMPOSITION_RATE_LIMIT_PER_MINUTE",
		cfg.Security.CompositionRateLimitPerMinute,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
