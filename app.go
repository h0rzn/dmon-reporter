package main

import (
	"fmt"

	"github.com/h0rzn/dmon-reporter/config"
)

type App struct {
	monitor   *Monitor
	publisher *Publisher
	config    *config.Config
}

func NewApp() *App {
	return &App{
		monitor: NewMonitor(),
	}
}

func (a *App) Run() {
	cfg, err := config.Load("config.toml")
	if err != nil {
		log.Warn("failed to load `config.toml`, loading defaults")
		// load default config
		a.config = config.NewDefaultConfig()
	}
	a.config = cfg
	fmt.Printf("config: %+v\n", a.config)

	err = a.monitor.Init()
	if err != nil {
		log.Fatal("monitor init failed, exit")
		return
	}
	err = a.monitor.Start()
	if err != nil {
		log.Fatalf("failed to start monitor: %s, exit", err.Error())
		return
	}
	metricsC := a.monitor.Out()

	a.publisher = NewPublisher(cfg)
	err = a.publisher.Init()
	if err != nil {
		log.Fatal(err)
		return
	}
	a.publisher.Run(metricsC)

}
