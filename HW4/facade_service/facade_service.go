package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var loggingServices []string
var messagingServices []string
var client = req.C()

var topic = "queue"

func getAddresses() {
	logAddresses := os.Getenv("LOG_ADDRESSES")
	if logAddresses == "" {
		panic("Logging services addresses are not set up")
	}
	msgServices := os.Getenv("MSG_ADDRESSES")
	if msgServices == "" {
		panic("Messaging service address is not set up")
	}
	loggingServices = strings.Split(logAddresses, ",")
	messagingServices = strings.Split(msgServices, ",")
}

func getHandles(c *gin.Context) {
	var cached []string
	var messages []string
	for range 5 {
		if err := client.Get(loggingServices[rand.Intn(len(loggingServices))]).Do().Into(&cached); err == nil {
			break
		} else {
			log.Println("Failed to connect to logging service. Retrying ...")
		}
	}
	if cached == nil {
		log.Println("Unable to reach logging service")
		c.String(http.StatusInternalServerError, "Unable to reach logging service")
		return
	}
	for range 5 {
		if err := client.Get(messagingServices[rand.Intn(len(messagingServices))]).Do().Into(&messages); err == nil {
			break
		} else {
			log.Println("Failed to connect to messaging service. Retrying ...")
		}
	}
	if messages == nil {
		log.Println("Unable to reach messaging service")
		c.String(http.StatusInternalServerError, "Unable to reach messaging service")
		return
	}

	c.String(http.StatusOK, "%s: %s", cached, messages)
}

func postHandler(c *gin.Context, producer *kafka.Producer) {
	var msg string
	if err := c.BindJSON(&msg); err != nil {
		log.Printf("Failed to parse JSON file: %e\n", err)
		c.String(http.StatusBadRequest, "failed to parse JSON file: %e", err)
		return
	}
	UUID := uuid.New()
	err := errors.Join(postLogging(UUID, msg), postMessaging(UUID, msg, producer))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func postLogging(UUID uuid.UUID, msg string) error {
	data := map[string]string{UUID.String(): msg}
	sent := false
	for range 5 {
		if client.Post(loggingServices[rand.Intn(len(loggingServices))]).SetBodyJsonMarshal(data).Do().Err == nil {
			sent = true
			break
		} else {
			log.Println("Failed to connect to logging service. Retrying ...")
		}
	}
	if !sent {
		return errors.New("logging service is not available")
	}
	return nil
}

func postMessaging(UUID uuid.UUID, msg string, producer *kafka.Producer) error {
	return producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            UUID[:],
		Value:          []byte(msg),
	}, nil)
}

func handleKafkaEvents(producer *kafka.Producer) {
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
			} else {
				key, _ := uuid.FromBytes(ev.Key)
				fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
					*ev.TopicPartition.Topic, key.String(), string(ev.Value))
			}
		}
	}
}

func createTopic() {
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "broker:29092"})
	if err != nil {
		log.Fatalf("Failed to create admin: %s", err)
	}
	defer adminClient.Close()
	_, err = adminClient.CreateTopics(context.Background(), []kafka.TopicSpecification{
		{Topic: topic, NumPartitions: 10, ReplicationFactor: 1},
	})
	if err != nil {
		log.Fatalf("Failed to create topic: %s", err)
	}
}

func main() {
	createTopic()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "broker:29092"})
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}
	defer producer.Close()
	go handleKafkaEvents(producer)

	getAddresses()

	router := gin.Default()

	router.GET("/facade_service", getHandles)
	router.POST("/facade_service", func(c *gin.Context) {
		postHandler(c, producer)
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
