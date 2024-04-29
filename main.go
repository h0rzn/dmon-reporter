package main

import (
	"fmt"
	"time"

	"github.com/h0rzn/dmon-reporter/store"
)

func main() {
	sqlite := store.SqliteProvider{}
	sqlConfig := map[string]string{
		"db_path": "./data.db",
	}
	err := sqlite.Init(sqlConfig)
	if err != nil {
		panic(err)
	}

	monitor := NewMonitor()
	err = monitor.Start()
	if err != nil {
		panic(err)
	}

	time.AfterFunc(30*time.Second, func() {
		fmt.Printf("hello after 20s")
		monitor.Stop()
		data, err := sqlite.Fetch()
		if err != nil {
			panic(err)
		}

		fmt.Printf("fetched %d items from cache:\n", len(data))
		for i, set := range data {
			fmt.Printf("%d -> %+v\n", i, set)
		}
		fmt.Println("dropping cache...")
		err = sqlite.Drop()
		fmt.Println("dropped")
		if err != nil {
			panic(err)
		}

		fmt.Printf("all done")

	})

	// var data store.CacheData
	for metrics := range monitor.Metrics() {
		//	fmt.Println("main: rcv metrics", metrics)
		fmt.Println("rcvd metrics")
		data := store.NewData(metrics.ID, time.Now(), metrics)
		sqlite.Push(data)
	}

}

// https://hackernoon.com/how-to-create-a-dynamic-pipeline-route-in-go
// https://golang.cafe/blog/golang-functional-options-pattern.html
