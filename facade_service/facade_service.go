package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"log"
	"net/http"
)

var loggingService = "http://log:8081/log"
var messagingService = "http://msg:8082/msg"
var client = req.C()

func getHandles(c *gin.Context) {
	var cached []string
	var messages string

	if err := client.Get(loggingService).Do().Into(&cached); err != nil {
		log.Fatal(err)
		return
	}
	if err := client.Get(messagingService).Do().Into(&messages); err != nil {
		log.Fatal(err)
		return
	}

	c.String(http.StatusOK, "%s: %s", cached, messages)
}

func postHandler(c *gin.Context) {
	var msg string
	if err := c.BindJSON(&msg); err != nil {
		log.Fatal(err)
		return
	}
	UUID := uuid.New()
	data := map[string]string{UUID.String(): msg}
	client.Post(loggingService).SetBodyJsonMarshal(data).Do()
	c.Status(http.StatusOK)
}

func main() {
	router := gin.Default()

	router.GET("/facade_service", getHandles)
	router.POST("/facade_service", postHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
