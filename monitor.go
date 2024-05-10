package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/h0rzn/dmon-reporter/store"
)

const containerPublishInterval = 5 * time.Second

type Monitor struct {
	dockerClient *client.Client
	clientCtx    context.Context
	mut          *sync.Mutex
	containers   map[string]*Container
	addContainer chan *Container
}

func NewMonitor() *Monitor {
	return &Monitor{
		clientCtx:    context.Background(),
		mut:          &sync.Mutex{},
		containers:   map[string]*Container{},
		addContainer: make(chan *Container),
	}
}

func (m *Monitor) Init() error {
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	m.dockerClient = dockerClient

	return nil
}

func (m *Monitor) Start() error {
	err := m.Init()
	if err != nil {
		return err
	}
	err = m.IndexContainers()
	if err != nil {
		return err
	}
	m.HandleContainerEvents()
	return nil
}

func (m *Monitor) IndexContainers() error {
	listOpts := container.ListOptions{}
	containers, err := m.dockerClient.ContainerList(m.clientCtx, listOpts)
	if err != nil {
		return err
	}

	for _, container := range containers {
		m.LinkContainer(container)
	}

	return nil
}

func (m *Monitor) LinkContainer(container types.Container) {
	if _, exists := m.containers[container.ID]; exists {
		log(fmt.Sprintf("container %s already linked\n", container.ID), LOG_MODE_WARNING)
		return
	}

	m.mut.Lock()
	wrappedContainer := NewContainer(container.ID, m.dockerClient, m.clientCtx, containerPublishInterval)
	m.containers[container.ID] = wrappedContainer
	m.mut.Unlock()
	log(fmt.Sprintf("LINKED %s", container.ID), LOG_MODE_INFO)
}

func (m *Monitor) LinkContainerByID(id string) {
	if _, exists := m.containers[id]; exists {
		log(fmt.Sprintf("container %s already linked\n", id), LOG_MODE_WARNING)
		return
	}

	containerJSON, err := m.dockerClient.ContainerInspect(m.clientCtx, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	m.mut.Lock()
	wrappedContainer := NewContainer(containerJSON.ID, m.dockerClient, m.clientCtx, containerPublishInterval)
	m.containers[wrappedContainer.ID] = wrappedContainer
	m.addContainer <- wrappedContainer
	m.mut.Unlock()
}

func (m *Monitor) UnlinkContainer(id string) {
	m.mut.Lock()
	if container, exists := m.containers[id]; exists {
		container.Stop()
		delete(m.containers, id)
	}
	m.mut.Unlock()
}

func (m *Monitor) HandleContainerEvents() {
	ctx := context.Background()
	filter := filters.Args{}
	eventChan, errChan := m.dockerClient.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	go func() {
		for {
			select {
			case event := <-eventChan:
				switch event.Action {
				case "stop":
					m.UnlinkContainer(event.Actor.ID)
				case "start":
					m.LinkContainerByID(event.Actor.ID)
				default:
					// fmt.Println(event.Action)
				}
			case err := <-errChan:
				log(fmt.Sprintf("Failed to read events: %s", err.Error()), LOG_MODE_ERROR)
			}
		}
	}()

}

func (m *Monitor) fetch() chan types.StatsJSON {
	out := make(chan types.StatsJSON)

	startRead := func(container *Container) {
		for stats := range container.ReadStats() {
			out <- stats
		}
	}
	// start reading for all currently indexed containers
	for _, container := range m.containers {
		go startRead(container)
	}
	// start reading for newly indexed containers
	go func() {
		for newContainer := range m.addContainer {
			go startRead(newContainer)
		}
	}()

	return out
}

func (m *Monitor) parse(statsChan chan types.StatsJSON) chan Metrics {
	out := make(chan Metrics)
	go func() {
		defer close(out)
		for stats := range statsChan {
			// cpu
			cpuPerc := parseCPU(stats.PreCPUStats, stats.CPUStats)
			// memory
			memUsage, memUsagePerc := parseMemory(stats.MemoryStats)

			metrics := Metrics{
				ID:              stats.ID,
				CPUUsagePerc:    cpuPerc,
				MemoryUsage:     memUsage,
				MemoryUsagePerc: memUsagePerc,
			}
			out <- metrics
		}
	}()

	return out
}

func (m *Monitor) Metrics() chan store.Data {
	metricsC := m.parse(m.fetch())
	outC := make(chan store.Data)
	go func() {
		for metric := range metricsC {
			_ = metric
			set := store.NewData(metric.ID, time.Now(), metric)
			outC <- *set
		}
	}()
	return outC
}

func (m *Monitor) Out() chan store.Data {
	return m.Metrics()
}

func (m *Monitor) Stop() {
	m.mut.Lock()
	containerStopGroup := &sync.WaitGroup{}
	for _, container := range m.containers {
		containerStopGroup.Add(1)
		container.StopInGroup(containerStopGroup)
	}
	containerStopGroup.Wait()
	log("Stopped all containers", LOG_MODE_INFO)
	m.mut.Unlock()
}
