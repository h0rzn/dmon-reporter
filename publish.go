package main

import (
	"time"

	"github.com/h0rzn/dmon-reporter/store"
)

// sendable() bool

// send usual data set / data group
// send() error

// send all cached data
// sendCached() error

// PIPELINE
// 1. Monitor [monitor.Metrics()]
// 2. Store-Provider
// 3. Publisher e

type Publisher struct {
	monitor         *Monitor
	cache           store.SqliteProvider
	retryTicker     time.Ticker
	remoteReachable bool
}

func (p *Publisher) Run() {
	p.sendLoop(p.monitor.Out())
}

func (p *Publisher) send(data any) error {
	_ = data
	return nil
}

func (p *Publisher) sendLoop(in <-chan any) {
	for data := range in {
		if p.remoteReachable {
			err := p.send(data)
			if err != nil && p.isSendErr(err) {
				p.remoteReachable = false
				go func() {
					<-p.waitForRemote()
					p.remoteReachable = true
				}()
				// push this dataset to cache
			}
		} else {
			// push to cache
		}
	}

	toCache := false
	var awaitRemoteC chan bool
	for {
		select {
		case data := <-in:
			if !toCache {
				if err := p.send(data); err != nil {
					toCache = true
					// send this dataset to cache
					// p.cache.Push(data)
					// start remote listener
					awaitRemoteC = p.subscribeRemoteStatus()
				}
			}
		case remoteStatus := <-awaitRemoteC:
			toCache = remoteStatus
		default:
		}
	}

}

func (p *Publisher) process(in chan any) {
	for set := range in {
		if p.remoteReachable {
			p.send(set)
		} else {
			// TODO use `set`
			p.cache.Push(&store.Data{})
		}
	}
}

func (p *Publisher) isSendErr(err error) bool {
	return err != nil
}

func (p *Publisher) subscribeRemoteStatus() chan bool {
	return make(chan bool)
}

func (p *Publisher) waitForRemote() chan bool {
	remoteReady := make(chan bool)
	go func() {
		for range p.retryTicker.C {
			// check if remote is reachable
			remoteReady <- true
		}
	}()

	return remoteReady
}

func (p *Publisher) sendStaleData() error {
	// fetch stale data from cache
	// send data
	return nil
}
