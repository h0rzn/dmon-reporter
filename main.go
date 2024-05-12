package main

import "github.com/sirupsen/logrus"

var log = logrus.New()

func main() {
	log.SetLevel(logrus.DebugLevel)

	app := NewApp()
	app.Run()
}

// https://hackernoon.com/how-to-create-a-dynamic-pipeline-route-in-go
// https://golang.cafe/blog/golang-functional-options-pattern.html
