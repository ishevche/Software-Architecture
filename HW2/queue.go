package main

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"os"
)

func getQueueArgs() (string, string) {
	switch len(os.Args) {
	case 2:
		return os.Args[1], "queue"
	case 3:
		return os.Args[1], os.Args[2]
	default:
		return "hw2_shevchenko_hazelcast", "queue"
	}
}

func main() {
	cfg := hazelcast.Config{}
	clusterName, qName := getQueueArgs()
	cfg.Cluster.Name = clusterName

	putChannel, getChannel1, getChannel2 := make(chan bool), make(chan bool), make(chan bool)
	go putValues(cfg, qName, putChannel)
	go getValues(cfg, qName, getChannel1)
	go getValues(cfg, qName, getChannel2)
	<-getChannel2
	<-getChannel1
	<-putChannel

	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	queue, _ := hzClient.GetQueue(ctx, qName)
	queue.Clear(ctx)
}

func putValues(cfg hazelcast.Config, qName string, done chan bool) {
	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	queue, _ := hzClient.GetQueue(ctx, qName)

	for i := range 100 {
		queue.Put(ctx, i+1)
	}
	queue.Put(ctx, -1)
	done <- true
}

func getValues(cfg hazelcast.Config, qName string, done chan bool) {
	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	queue, _ := hzClient.GetQueue(ctx, qName)

	var values []int64

	for {
		value, _ := queue.Take(ctx)
		if value.(int64) == -1 {
			queue.Put(ctx, value)
			break
		}
		values = append(values, value.(int64))
	}

	fmt.Println("Values got from the queue:")
	fmt.Println(values)

	done <- true
}
