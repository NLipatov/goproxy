package builders

import (
	"context"
	"database/sql"
	"fmt"
	"goproxy/application"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/eventhandlers"
	"goproxy/infrastructure/services"
	"log"
)

type UserRepositoryBuilder struct {
}

func NewUserRepositoryBuilder() *UserRepositoryBuilder {
	return &UserRepositoryBuilder{}
}

func (u *UserRepositoryBuilder) Build(kafkaGroupId string, boundedContext domain.BoundedContexts,
	cache repositories.BigCacheUserRepositoryCache, db *sql.DB) (application.UserRepository, error) {
	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(boundedContext)
	if kafkaConfigErr != nil {
		return nil, kafkaConfigErr
	}

	userRepoKafkaConf := config.KafkaConfig{
		BootstrapServers: kafkaConfig.BootstrapServers,
		GroupID:          kafkaGroupId,
		AutoOffsetReset:  kafkaConfig.AutoOffsetReset,
		Topic:            kafkaConfig.Topic,
	}

	userRepoKafka, userRepoKafkaErr := services.NewKafkaService(userRepoKafkaConf)
	if userRepoKafkaErr != nil {
		log.Fatal(userRepoKafkaErr)
	}
	repoEventHandler := eventhandlers.NewUserPasswordChangedEventHandler(cache)
	repoEventProcessor := application.NewEventProcessor(userRepoKafka).
		RegisterTopic(fmt.Sprintf("%s", domain.PROXY)).
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

	return repositories.NewUserRepository(db, cache), nil
}
