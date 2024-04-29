package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Container struct {
	ID             string
	dockerClient   *client.Client
	ctx            context.Context
	ctxCancelFunc  context.CancelFunc
	publishTicker  *time.Ticker
	prunableSig    chan bool
	metricsRunning bool
}

func NewContainer(ID string, dockerClient *client.Client, ctx context.Context, publishInterval time.Duration) *Container {
	metricsCtx, metricsCancelFunc := context.WithCancel(context.Background())
	return &Container{
		ID,
		dockerClient,
		metricsCtx,
		metricsCancelFunc,
		time.NewTicker(publishInterval),
		make(chan bool, 1),
		false,
	}
}

func (c *Container) StatsReader() (io.ReadCloser, error) {
	ctx := context.Background()
	r, err := c.dockerClient.ContainerStats(ctx, c.ID, true)
	return r.Body, err
}

func (c *Container) ReadStats() chan types.StatsJSON {
	out := make(chan types.StatsJSON)
	statsReader, err := c.StatsReader()
	if err != nil {
		panic(err)
	}
	statsDecoder := json.NewDecoder(statsReader)
	c.metricsRunning = true

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				statsReader.Close()
				close(out)
				c.prunableSig <- true
				return
			case <-c.publishTicker.C:
				var stats types.StatsJSON
				statsDecoder.Decode(&stats)
				out <- stats
			}
		}

	}()
	return out
}

func (c *Container) Stop() {
	c.ctxCancelFunc()
	select {
	case <-c.prunableSig:
		break
	case <-time.After(5 * time.Second):
		log(fmt.Sprintf("Stop timeout: %s", c.ID), LOG_MODE_WARNING)
	}
	close(c.prunableSig)
	log(fmt.Sprintf("Stopped: %s", c.ID), LOG_MODE_INFO)
}

func (c *Container) StopInGroup(wg *sync.WaitGroup) {
	c.Stop()
	wg.Done()
}
