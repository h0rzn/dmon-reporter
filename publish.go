package main

import (
	"fmt"
	"net"
	"time"

	"github.com/h0rzn/dmon-reporter/config"
	"github.com/h0rzn/dmon-reporter/store"
	"golang.org/x/net/context"
)

type Publisher struct {
	monitor *Monitor
	cache   store.OfflineCache
	// receive `true` if remote is reachable again
	// `false` if remote started beeing unreachable
	remoteAvailableC chan bool
	// send `true` to start, `false` to stop retrying
	controlRetryC chan bool
	ctx           *context.Context
	cancelFunc    context.CancelFunc
	config        *config.Config
}

func NewPublisher(config *config.Config) *Publisher {
	return &Publisher{
		remoteAvailableC: make(chan bool),
		controlRetryC:    make(chan bool),
		config:           config,
	}
}

func (p *Publisher) Init() error {
	switch p.config.Cache.Provider {
	case "sqlite":
		p.cache = &store.SqliteProvider{}
	}

	err := p.cache.Init(p.config)
	if err != nil {
		return err
	}
	return nil
}

func (p *Publisher) Run(in chan store.Data) {
	go p.sendLoop(in)
	p.handleRetries()
}

func (p *Publisher) send(data any) error {
	log.Debug("send")
	_ = data
	timeout := time.Duration(p.config.Master.Send_Timeout) * time.Second
	_, err := net.DialTimeout(p.config.Master.Protocol, p.config.Master.Addr, timeout)
	return err
}

func (p *Publisher) sendLoop(in <-chan store.Data) {
	remoteAvailable := false
	for {
		select {
		case available := <-p.remoteAvailableC:
			// remote is available again
			if available {
				log.Info("remote is back up ", time.Now())
				if err := p.sendStaleData(); err == nil {
					p.controlRetryC <- false
				} else {
					fmt.Println(err)
				}
			}
			remoteAvailable = available
		case data := <-in:
			if !remoteAvailable {
				err := p.cache.Push(data)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				if err := p.send(data); p.isSendErr(err) {
					p.controlRetryC <- true
					remoteAvailable = false
				}
			}
		}
	}
}

func (p *Publisher) isSendErr(err error) bool {
	return err != nil
}

func (p *Publisher) handleRetries() {
	ticker := time.NewTicker(time.Duration(p.config.Master.Retry_Interval) * time.Second)
	i := 0
	for {
		select {
		case signal := <-p.controlRetryC:
			if signal {
				ticker.Reset(500 * time.Millisecond)
			} else {
				ticker.Stop()
			}
		case <-ticker.C:
			if i > 0 {
				sig := p.retry()
				p.remoteAvailableC <- sig
			}
			// skip first tick
			i = i + 1
		}
	}
}

func (p *Publisher) retry() bool {
	log.Debug("retrying")
	_, err := net.Dial(p.config.Master.Protocol, p.config.Master.Addr)
	return err == nil
}

func (p *Publisher) sendStaleData() error {
	stale, err := p.cache.Fetch()
	// log.Info("sending stale data (", len(stale), ")")
	// for _, set := range stale {
	// fmt.Println(set.ID(), set.When())
	// }
	_ = stale
	if err != nil {
		return err
	}

	return nil
}
