package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	// Kafka Consumer Configuration
	broker := "lkc-rg8jm0-ap832wr4.us-east-2.aws.accesspoint.glb.confluent.cloud:9092" // Replace with your broker(s)
	group := "ams-poc-101"                                                             // Replace with your consumer group ID
	topic := "ams-poc-101"                                                             // Replace with your topic name

	config := &kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          group,
		"auto.offset.reset": "earliest", // Start from the beginning if no offset is stored
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     os.Getenv("apiKey"),
		"sasl.password":     os.Getenv("apiSecret"),
	}

	// Create a new Kafka Consumer
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Subscribe to the topic
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	log.Printf("Consumer started. Subscribed to topic: %s\n", topic)

	// Graceful Shutdown Handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigChan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			// Poll for messages
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err == nil {
				log.Printf("Message received on %s: %s\n", msg.TopicPartition, string(msg.Value))
			} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
				log.Printf("Error reading message: %v\n", err)
			}
		}
	}

	log.Println("Closing consumer...")
}
