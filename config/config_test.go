package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsOpenAIConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "base.yaml")
	content := `
env: local
server:
  host: 127.0.0.1
  port: 9090
composition:
  max_upload_bytes: 1234
openai:
  api_key: test-key
  model: gpt-5.5
  base_url: https://api.openai.com/v1/responses
  timeout: 30s
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.OpenAI.APIKey != "test-key" {
		t.Fatalf("api key=%q want=%q", cfg.OpenAI.APIKey, "test-key")
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("port=%d want=%d", cfg.Server.Port, 9090)
	}
}
