package main

import (
	"context"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	kafkaTopic = "user_creation"
)

type KafkaMessage struct {
	msg string
	ts  time.Time
}

type KafkaConsumer struct {
	c           *kafka.Consumer
	topic       string
	ch          *CH
	batch       []KafkaMessage
	batchMaxLen int
	err         error
}

func (c *KafkaConsumer) Destroy() {
	c.c.Close()
}

func (k *KafkaConsumer) Connect(server, kafkagroup string, ch *CH, topic string) {

	var err error
	defer func() {
		k.err = err
	}()

	k.topic = topic
	k.ch = ch

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		// https://kafka.apache.org/documentation.html#consumerconfigs
		"bootstrap.servers": server,
		"group.id":          kafkagroup,
		// We will commit after writing to the Clickhouse
		"enable.auto.commit": "false",
		// After at least 500 ms, the consumer will receive messages
		"fetch.wait.max.ms": "500",
		// The broker will hold on to the fetch until enough data is available (at least 50 bytes)
		"fetch.min.bytes": "50",
		// Receive all old messages if it's possible
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("ERROR: [Logger] Failed Kafka connection: %s", err)
		return
	}

	k.c = c

	err = k.createTopic(topic)
	if err != nil {
		log.Fatalf("ERROR: [Logger] Failed topic creation: %s", err)
		return
	}

	err = k.c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("ERROR: [Logger] Failed topic subcription: %s", err)
		return
	}

}

func (k *KafkaConsumer) createTopic(topic string) error {
	a, err := kafka.NewAdminClientFromConsumer(k.c)
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

func (k *KafkaConsumer) Listen() {
	for {
		msg, err := k.c.ReadMessage(-1)
		if err == nil {
			k.handleMsg(msg)
		} else {
			// The client will automatically try to recover from all errors.
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (k *KafkaConsumer) handleMsg(msg *kafka.Message) {
	log.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
	k.batch = append(k.batch, KafkaMessage{string(msg.Value), time.Now()})
	if len(k.batch) >= k.batchMaxLen {
		err := k.ch.Write(k.batch)
		if err == nil {
			k.batch = make([]KafkaMessage, 0)
			log.Printf("Commit the Message")
			k.c.CommitMessage(msg)
		}
	}
}

func (k *KafkaConsumer) Close() {
	k.c.Close()
}

func NewKafkaConsumer(batchMaxLen int) *KafkaConsumer {
	k := &KafkaConsumer{
		batchMaxLen: batchMaxLen,
	}
	return k
}
