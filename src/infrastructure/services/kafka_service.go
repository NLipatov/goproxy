package services

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"goproxy/application"
	"goproxy/domain/events"
)

type KafkaService struct {
	consumer *kafka.Consumer
	producer *kafka.Producer
	topics   []string
}

func NewKafkaService(bootstrapServers, groupId, autoOffsetReset string) (application.MessageBusService, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           groupId,
		"auto.offset.reset":  autoOffsetReset,
		"session.timeout.ms": 10000,
		"fetch.min.bytes":    1,
		"fetch.wait.max.ms":  10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %v", err)
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		_ = consumer.Close()
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaService{
		consumer: consumer,
		producer: producer,
	}, nil
}

func (k KafkaService) Subscribe(topics []string) error {
	if len(topics) == 0 {
		return fmt.Errorf("no topics provided to subscribe")
	}
	k.topics = topics
	err := k.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics %v: %v", topics, err)
	}
	return nil
}

func (k KafkaService) Consume() (*events.OutboxEvent, error) {
	msg, err := k.consumer.ReadMessage(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to consume message: %v", err)
	}

	var event events.OutboxEvent
	if unmarshalErr := json.Unmarshal(msg.Value, &event); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to deserialize message: %v", unmarshalErr)
	}
	return &event, nil
}

func (k KafkaService) Produce(topic string, event events.OutboxEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %v", err)
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("failed to deliver message: %v", m.TopicPartition.Error)
	}

	return nil
}

func (k KafkaService) Close() error {
	if k.consumer != nil {
		if err := k.consumer.Close(); err != nil {
			return fmt.Errorf("failed to close consumer: %v", err)
		}
	}

	if k.producer != nil {
		k.producer.Flush(10000)
		k.producer.Close()
	}

	return nil
}
