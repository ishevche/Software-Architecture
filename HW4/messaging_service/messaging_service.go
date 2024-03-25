package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"net/http"
	"time"
)

var topic = "queue"

var data = make(map[string]string)

func requestMessages(c *gin.Context) {
	c.JSON(http.StatusOK, maps.Values(data))
}

func main() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "broker:29092",
		"group.id":          "messaging",
		"auto.offset.reset": "earliest"})
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	err = consumer.Subscribe(topic, nil)
	if err != nil {
		log.Fatal(err)
	}

	println("Subscribed")

	stop := make(chan bool, 1)
	defer func() { stop <- true }()

	go consumeMessages(consumer, stop)

	router := gin.Default()

	router.GET("/msg", requestMessages)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func consumeMessages(consumer *kafka.Consumer, stop chan bool) {
	run := true
	for run {
		select {
		case <-stop:
			run = false
		default:
			ev, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				continue
			}
			UUID, err := uuid.FromBytes(ev.Key)
			if err != nil {
				fmt.Printf("Error occurred while reading from Kafka: %e\n", err)
				continue
			}
			msg := string(ev.Value)
			data[UUID.String()] = msg
			fmt.Printf("Got message: key=%s, value=%s\n", UUID.String(), msg)
		}
	}
}
