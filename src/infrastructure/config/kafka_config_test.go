package config_test

import (
	"goproxy/domain"
	"goproxy/infrastructure/config"
	"os"
	"testing"
)

func TestNewKafkaConfig_Success(t *testing.T) {
	if err := os.Setenv("PROXY_KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}
	if err := os.Setenv("PROXY_KAFKA_GROUP_ID", "traffic-group"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}
	if err := os.Setenv("PROXY_KAFKA_AUTO_OFFSET_RESET", "earliest"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}
	if err := os.Setenv("PROXY_KAFKA_TOPIC", "traffic-topic"); err != nil {
		t.Fatalf("failed to set environment variable: %v", err)
	}

	defer func() {
		if err := os.Unsetenv("PROXY_KAFKA_BOOTSTRAP_SERVERS"); err != nil {
			t.Errorf("failed to unset environment variable: %v", err)
		}
		if err := os.Unsetenv("PROXY_KAFKA_GROUP_ID"); err != nil {
			t.Errorf("failed to unset environment variable: %v", err)
		}
		if err := os.Unsetenv("PROXY_KAFKA_AUTO_OFFSET_RESET"); err != nil {
			t.Errorf("failed to unset environment variable: %v", err)
		}
		if err := os.Unsetenv("PROXY_KAFKA_TOPIC"); err != nil {
			t.Errorf("failed to unset environment variable: %v", err)
		}
	}()

	context := domain.BoundedContexts("PROXY")
	kafkaConfig, err := config.NewKafkaConfig(context)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if kafkaConfig.BootstrapServers != "localhost:9092" {
		t.Errorf("expected BootstrapServers to be 'localhost:9092', got: %s", kafkaConfig.BootstrapServers)
	}
	if kafkaConfig.GroupID != "traffic-group" {
		t.Errorf("expected GroupID to be 'traffic-group', got: %s", kafkaConfig.GroupID)
	}
	if kafkaConfig.AutoOffsetReset != "earliest" {
		t.Errorf("expected AutoOffsetReset to be 'earliest', got: %s", kafkaConfig.AutoOffsetReset)
	}
	if kafkaConfig.Topic != "traffic-topic" {
		t.Errorf("expected Topic to be 'traffic-topic', got: %s", kafkaConfig.Topic)
	}
}

func TestNewKafkaConfig_MissingEnvVar(t *testing.T) {
	if err := os.Unsetenv("PROXY_KAFKA_BOOTSTRAP_SERVERS"); err != nil {
		t.Fatalf("failed to unset environment variable: %v", err)
	}

	context := domain.PROXY
	_, err := config.NewKafkaConfig(context)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expectedError := "PROXY_KAFKA_BOOTSTRAP_SERVERS env var not set"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got: %s", expectedError, err.Error())
	}
}

func TestNewKafkaConfig_UnsetContext(t *testing.T) {
	context := domain.UNSET
	_, err := config.NewKafkaConfig(context)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expectedError := "unset context"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got: %s", expectedError, err.Error())
	}
}
