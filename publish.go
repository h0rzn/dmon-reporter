package main

import (
	"fmt"
	"net"
	"time"

	"github.com/h0rzn/dmon-reporter/store"
	"golang.org/x/net/context"
)

const (
	RETRY_INTERVAL = 5 * time.Second
	SEND_TIMEOUT   = 2 * time.Second
)

type Publisher struct {
	monitor *Monitor
	cache   store.SqliteProvider
	// receive `true` if remote is reachable again
	// `false` if remote started beeing unreachable
	remoteAvailableC chan bool
	// send `true` to start, `false` to stop retrying
	controlRetryC chan bool
	ctx           *context.Context
	cancelFunc    context.CancelFunc
}

func NewPublisher() *Publisher {
	return &Publisher{
		remoteAvailableC: make(chan bool),
		controlRetryC:    make(chan bool),
	}
}

func (p *Publisher) Run(in chan store.Data) {
	p.cache.Init(map[string]string{})
	go p.sendLoop(in)
	p.handleRetries()
}

func (p *Publisher) send(data any) error {
	_ = data
	_, err := net.DialTimeout("tcp", "127.0.0.1:8080", SEND_TIMEOUT)
	return err
}

func (p *Publisher) sendLoop(in <-chan store.Data) {
	remoteAvailable := false
	for {
		select {
		case available := <-p.remoteAvailableC:
			// remote is available again
			if available {
				if err := p.sendStaleData(); err != nil {
					fmt.Println(err)
				}
			}
			remoteAvailable = available
		case data := <-in:
			fmt.Println("in", data.ID())
			if !remoteAvailable {
				err := p.cache.Push(data)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				if err := p.send(data); p.isSendErr(err) {
					fmt.Println("sending to remote failed")
					p.controlRetryC <- true
				}
			}
		}
	}
}

func (p *Publisher) isSendErr(err error) bool {
	return err != nil
}

func (p *Publisher) handleRetries() {
	ticker := time.NewTicker(RETRY_INTERVAL)
	i := 0
	for {
		select {
		case signal := <-p.controlRetryC:
			if signal {
				fmt.Printf("handle: start retry (sig: %t)\n", signal)
				ticker.Reset(500 * time.Millisecond)
			} else {
				fmt.Printf("handle: stop retry (sig: %t)\n", signal)
				ticker.Stop()
			}
		case <-ticker.C:
			if i > 0 {
				sig := p.retry()
				fmt.Println("handle retries: send: ", sig)
				p.remoteAvailableC <- sig
				i = i + 1
			}
		}
	}
}

func (p *Publisher) retry() bool {
	fmt.Println("RETRY")
	_, err := net.Dial("tcp", "127.0.0.1:8080")
	//fmt.Println(err)
	return err == nil
}

func (p *Publisher) sendStaleData() error {
	// fetch stale data from cache
	// send data
	fmt.Println("sending stale data...")
	return nil
}
