package main

import (
	"context"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	p     *kafka.Producer
	topic string
	flush int
	err error
}

func (k *KafkaProducer) Connect(conf *KafkaConfig) {
	var err error
	defer func() {
		k.err = err
	}()
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": conf.server})
	if err != nil {
		return
	}
	k.p = p
	k.topic = conf.topic
	k.flush = 1000
	err = k.createTopic(conf.topic)
}

func (k *KafkaProducer) createTopic(topic string) error {
	a, err := kafka.NewAdminClientFromProducer(k.p)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, err := a.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1}},
		kafka.SetAdminOperationTimeout(time.Second*5))
	if err != nil {
		log.Printf("Admin Client request error: %v\n", err)
		return err
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Printf("Failed to create topic: %v\n", result.Error)
			return result.Error
		}
		log.Printf("Creation topic: %v\n", result)
	}
	return nil
}

func (k *KafkaProducer) Close() {
	k.p.Close()
}

// Delivery report handler for produced messages
func (k *KafkaProducer) deliveryHandler() {
	for e := range k.p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				log.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
		}
	}
}

func (p *KafkaProducer) Produce(msg string) {
	log.Printf("Producing message: %v", msg)
	p.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          []byte(msg),
	}, nil)
	p.p.Flush(p.flush)
}
