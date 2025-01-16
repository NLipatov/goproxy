package eventhandlers

import (
	"context"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/services"
	"log"
)

type UserPasswordChangedEventProcessor[T any] struct {
	boundedContext domain.BoundedContexts
	cache          application.Cache[T]
}

func NewUserPasswordChangedEventProcessor[T any](boundedContext domain.BoundedContexts, cache application.Cache[T]) *UserPasswordChangedEventProcessor[T] {
	return &UserPasswordChangedEventProcessor[T]{
		boundedContext,
		cache,
	}
}

func (c *UserPasswordChangedEventProcessor[T]) ProcessEvents() error {
	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(c.boundedContext)
	if kafkaConfigErr != nil {
		return kafkaConfigErr
	}

	kafkaConf := config.KafkaConfig{
		BootstrapServers: kafkaConfig.BootstrapServers,
		GroupID:          "UserPasswordChangedEventProcessor",
		AutoOffsetReset:  kafkaConfig.AutoOffsetReset,
		Topic:            kafkaConfig.Topic,
	}

	kafka, kafkaErr := services.NewKafkaService(kafkaConf)
	if kafkaErr != nil {
		log.Fatal(kafkaErr)
	}

	repoEventHandler := NewUserPasswordChangedEventHandler(c.cache)
	repoEventProcessor := application.NewEventProcessor(kafka).
		RegisterTopic(fmt.Sprintf("%s", c.boundedContext)).
		RegisterHandler("UserPasswordChangedEvent", repoEventHandler)

	if repoEventProcessorErr := repoEventProcessor.Build(); repoEventProcessorErr != nil {
		log.Fatal(repoEventProcessorErr)
	}

	go func() {
		processingErr := repoEventProcessor.Start(context.Background())
		if processingErr != nil {
			log.Fatal(processingErr)
		}
	}()

	return nil
}
