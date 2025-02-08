package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gocolly/colly"
)

func main() {
	// Kafka configuration
	// broker := "lkc-rg8jm0-ap832wr4.us-east-2.aws.accesspoint.glb.confluent.cloud:9092"
	broker := "lkc-rg8jm0-ap832wr4.us-east-2.aws.accesspoint.glb.confluent.cloud:9092"
	topic := "ams-poc-101"
	// apiKey := os.Getenv("apiKey")
	// apiSecret := os.Getenv("apiSecret")

	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"linger.ms":         10,
		"sasl.username":     os.Getenv("apiKey"),
		"sasl.password":     os.Getenv("apiSecret"),
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Delivery report handler
	go func() {
		for e := range kafkaProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Message delivered to %v\n", ev.TopicPartition)
					startTime := ev.Opaque.(time.Time) // Retrieve the start time from the opaque field
					delay := time.Since(startTime)
					log.Printf("Message delivered to %v in %v\n", ev.TopicPartition, delay)
				}
			}
		}
	}()

	// Colly setup
	c := colly.NewCollector(
		colly.AllowedDomains("quotes.toscrape.com"),
		colly.IgnoreRobotsTxt(),
		colly.AllowURLRevisit(),
	)

	// Scrape quotes
	c.OnHTML(".quote", func(e *colly.HTMLElement) {
		quote := e.ChildText(".text")
		author := e.ChildText(".author")
		tags := e.ChildText(".tags")

		message := fmt.Sprintf("Quote: %s | Author: %s | Tags: %s", quote, author, tags)
		startTime := time.Now()
		// Produce message to Kafka
		err := kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(message),
			Opaque:         startTime,
		}, nil)
		if err != nil {
			log.Printf("Failed to produce message: %v", err)
		}
	})

	// Handle pagination links
	c.OnHTML(".pager a[href]", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		log.Printf("Found next page: %s", nextPage)
		c.Visit(nextPage) // Visit the next page
	})

	// Error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request failed: %v", err)
	})

	// Start scraping
	for {
		fmt.Println("Starting to scrape...")
		err := c.Visit("http://quotes.toscrape.com")
		if err != nil {
			log.Printf("Scraping failed: %v", err)
		}
		fmt.Println("Scraping completed. Waiting for the next iteration...")

		// Sleep for a while to avoid overwhelming the server
		time.Sleep(2 * time.Second)
		kafkaProducer.Flush(15 * 1000)
		log.Println("Scraping and producing completed.")
	}

}
