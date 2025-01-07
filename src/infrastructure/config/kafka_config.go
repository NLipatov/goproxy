package config

import (
	"fmt"
	"os"
)

const (
	bootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	groupIdKey          = "KAFKA_GROUP_ID"
	offsetKey           = "KAFKA_AUTO_OFFSET_RESET"
	topicKey            = "KAFKA_TOPIC"
)

type KafkaConfig struct {
	BootstrapServers string
	GroupID          string
	AutoOffsetReset  string
	Topic            string
}

func NewKafkaConfig(context BoundedContext) (KafkaConfig, error) {
	if context == UNSET {
		return KafkaConfig{}, fmt.Errorf("unset context")
	}

	bootstrapServers, err := getEnv(context, bootstrapServersKey)
	if err != nil {
		return KafkaConfig{}, err
	}

	groupID, err := getEnv(context, groupIdKey)
	if err != nil {
		return KafkaConfig{}, err
	}

	autoOffsetReset, err := getEnv(context, offsetKey)
	if err != nil {
		return KafkaConfig{}, err
	}

	topic, err := getEnv(context, topicKey)
	if err != nil {
		return KafkaConfig{}, err
	}

	return KafkaConfig{
		BootstrapServers: bootstrapServers,
		GroupID:          groupID,
		AutoOffsetReset:  autoOffsetReset,
		Topic:            topic,
	}, nil
}

func getEnv(context BoundedContext, envVarName string) (string, error) {
	envVarKey := fmt.Sprintf("%s_%s", context, envVarName)
	value := os.Getenv(envVarKey)
	if value == "" {
		return "", NewEnvVarNotSetError(envVarKey)
	}
	return value, nil
}
