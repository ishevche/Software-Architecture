package main

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"os"
	"time"
)

func getIncrementArgs() (string, string, string) {
	switch len(os.Args) {
	case 2:
		return os.Args[1], "map", "key"
	case 3:
		return os.Args[1], os.Args[2], "key"
	case 4:
		return os.Args[1], os.Args[2], os.Args[3]
	default:
		return "hw2_shevchenko_hazelcast", "map", "key"
	}
}

func main() {
	ctx := context.TODO()
	cfg := hazelcast.Config{}
	clusterName, mapName, keyName := getIncrementArgs()
	cfg.Cluster.Name = clusterName
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	myMap, _ := hzClient.GetMap(ctx, mapName)

	myMap.Put(ctx, keyName, 0)
	performIncrement(cfg, mapName, keyName, myMap, ctx, nonBlocking, "no blocking")
	myMap.Put(ctx, keyName, 0)
	performIncrement(cfg, mapName, keyName, myMap, ctx, pessimisticLocking, "pessimistic blocking")
	myMap.Put(ctx, keyName, 0)
	performIncrement(cfg, mapName, keyName, myMap, ctx, optimisticLocking, "optimistic blocking")
	myMap.Put(ctx, keyName, 0)
}

func performIncrement(cfg hazelcast.Config, mapName string, keyName string, myMap *hazelcast.Map, ctx context.Context, function func(cfg hazelcast.Config, mapName string, keyName string, chanel chan bool), method string) {
	fmt.Printf("\nStarting incrementing with %s...\n", method)
	done := make(chan bool, 3)
	startTime := time.Now()
	go function(cfg, mapName, keyName, done)
	go function(cfg, mapName, keyName, done)
	go function(cfg, mapName, keyName, done)
	<-done
	<-done
	<-done
	nonBlockingTime := time.Since(startTime)
	value, _ := myMap.Get(ctx, keyName)
	fmt.Println("Ended incrementing")
	fmt.Println("Resulting value:", value)
	fmt.Println("Time taken:", nonBlockingTime)
}

func nonBlocking(cfg hazelcast.Config, mapName string, keyName string, chanel chan bool) {
	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	myMap, _ := hzClient.GetMap(ctx, mapName)

	for range 10_000 {
		value, _ := myMap.Get(ctx, keyName)
		myMap.Set(ctx, keyName, value.(int64)+1)
	}

	chanel <- true
}

func pessimisticLocking(cfg hazelcast.Config, mapName string, keyName string, chanel chan bool) {
	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	myMap, _ := hzClient.GetMap(ctx, mapName)

	for range 10_000 {
		myMap.Lock(ctx, keyName)
		defer myMap.Unlock(ctx, keyName)
		value, _ := myMap.Get(ctx, keyName)
		myMap.Set(ctx, keyName, value.(int64)+1)
	}

	chanel <- true
}

func optimisticLocking(cfg hazelcast.Config, mapName string, keyName string, chanel chan bool) {
	ctx := context.TODO()
	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)
	myMap, _ := hzClient.GetMap(ctx, mapName)

	for range 10_000 {
		for {
			value, _ := myMap.Get(ctx, keyName)
			newValue := value.(int64) + 1
			if same, _ := myMap.ReplaceIfSame(ctx, keyName, value, newValue); same {
				break
			}
		}
	}

	chanel <- true
}
