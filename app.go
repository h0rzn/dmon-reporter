package main

import (
	"fmt"
)

type App struct {
	monitor   *Monitor
	publisher *Publisher
}

func NewApp() *App {
	return &App{
		monitor:   NewMonitor(),
		publisher: NewPublisher(),
	}
}

func (a *App) Run() {
	err := a.monitor.Init()
	if err != nil {
		fmt.Println("monitor init failed, exit")
		return
	}
	err = a.monitor.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	metricsC := a.monitor.Out()
	a.publisher.Run(metricsC)

}
