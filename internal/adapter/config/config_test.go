package config

import (
	"testing"
)

func TestConfig_Init_LoadsEnvFile(t *testing.T) {
	SetMockEnv(map[string]string{
		"HTTP_PORT":            "9999",
		"HTTP_URL":             "http://test.com",
		"ENV":                  "test",
		"DYNAMODB_TABLE":       "orders-test",
		"AWS_REGION":           "us-east-2",
		"HTTP_ALLOWED_ORIGINS": "http://test.com",
	})
	defer UnsetMockEnv([]string{
		"HTTP_PORT", "HTTP_URL", "ENV",
		"DYNAMODB_TABLE", "AWS_REGION", "HTTP_ALLOWED_ORIGINS",
	})

	cfg, err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if cfg.Http == nil {
		t.Error("Config.Http not initialized")
	}
	if cfg.DynamoDB == nil {
		t.Error("Config.DynamoDB not initialized")
	}

	if cfg.Http.Port != "9999" {
		t.Errorf("Expected HTTP_PORT='9999', got %q", cfg.Http.Port)
	}
	if cfg.Env != "test" {
		t.Errorf("Expected ENV='test', got %q", cfg.Env)
	}
	if cfg.DynamoDB.TableName != "orders-test" {
		t.Errorf("Expected DYNAMODB_TABLE='orders-test', got %q", cfg.DynamoDB.TableName)
	}
	if cfg.DynamoDB.Region != "us-east-2" {
		t.Errorf("Expected AWS_REGION='us-east-2', got %q", cfg.DynamoDB.Region)
	}
}

func TestConfig_Initi_GetEnvReturnsValueOrDefault(t *testing.T) {
	SetMockEnv(map[string]string{"TEST_KEY": "test_value"})
	val := getEnv("TEST_KEY", "default")
	if val != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", val)
	}
	UnsetMockEnv([]string{"TEST_KEY"})

	val = getEnv("TEST_KEY", "default")
	if val != "default" {
		t.Errorf("Expected 'default', got '%s'", val)
	}
}

func TestConfig_Init_DogStatsD(t *testing.T) {
	tests := []struct {
		name         string
		agentHost    string
		wantAddr     string
		wantDisabled bool
	}{
		{
			name:         "sem DD_AGENT_HOST",
			agentHost:    "",
			wantAddr:     "",
			wantDisabled: true,
		},
		{
			name:         "com hostname do agent",
			agentHost:    "datadog-agent",
			wantAddr:     "datadog-agent:8125",
			wantDisabled: false,
		},
		{
			name:         "com IPv4",
			agentHost:    "127.0.0.1",
			wantAddr:     "127.0.0.1:8125",
			wantDisabled: false,
		},
		{
			name:         "com IPv6",
			agentHost:    "::1",
			wantAddr:     "[::1]:8125",
			wantDisabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DD_AGENT_HOST", tt.agentHost)

			cfg, err := Init()
			if err != nil {
				t.Fatalf("Init() erro: %v", err)
			}
			if cfg.DogStatsD == nil {
				t.Fatal("cfg.DogStatsD é nil")
			}
			if cfg.DogStatsD.Addr != tt.wantAddr {
				t.Errorf("DogStatsD.Addr = %q, want %q", cfg.DogStatsD.Addr, tt.wantAddr)
			}
			if cfg.DogStatsD.Disabled != tt.wantDisabled {
				t.Errorf("DogStatsD.Disabled = %v, want %v", cfg.DogStatsD.Disabled, tt.wantDisabled)
			}
		})
	}
}
