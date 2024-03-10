package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var loggingServices []string
var messagingService string
var client = req.C()

func getAddresses() {
	logAddresses := os.Getenv("LOG_ADDRESSES")
	if logAddresses == "" {
		panic("Logging services addresses are not set up")
	}
	messagingService = os.Getenv("MSG_ADDRESSES")
	if messagingService == "" {
		panic("Messaging service address is not set up")
	}
	loggingServices = strings.Split(logAddresses, ",")
}

func getHandles(c *gin.Context) {
	var cached []string
	var messages string
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
	if err := client.Get(messagingService).Do().Into(&messages); err != nil {
		log.Println(err)
		c.String(http.StatusInternalServerError, "Unable to reach messaging service")
		return
	}

	c.String(http.StatusOK, "%s: %s", cached, messages)
}

func postHandler(c *gin.Context) {
	var msg string
	if err := c.BindJSON(&msg); err != nil {
		log.Printf("Failed to parse JSON file: %e\n", err)
		c.String(http.StatusBadRequest, "Failed to parse JSON file: %e", err)
		return
	}
	UUID := uuid.New()
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
	if sent {
		c.Status(http.StatusOK)
	} else {
		c.String(http.StatusServiceUnavailable, "Logging service is not available")
	}
}

func main() {
	getAddresses()

	router := gin.Default()

	router.GET("/facade_service", getHandles)
	router.POST("/facade_service", postHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
