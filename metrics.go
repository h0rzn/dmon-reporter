package main

import "github.com/docker/docker/api/types"

type Metrics struct {
	ID              string
	CPUUsagePerc    float64
	MemoryUsage     float64
	MemoryUsagePerc float64
}

func parseCPU(preCPU, sysCPU types.CPUStats) float64 {
	var cpuPerc = 0.0
	prevCPUU := preCPU.CPUUsage.TotalUsage
	prevSysU := preCPU.SystemUsage

	cpuDelta := float64(sysCPU.CPUUsage.TotalUsage) - float64(prevCPUU)
	systemDelta := float64(sysCPU.SystemUsage) - float64(prevSysU)

	online := float64(sysCPU.OnlineCPUs)

	if online == 0.0 {
		online = float64(len(sysCPU.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 {
		cpuPerc = (cpuDelta / systemDelta) * online * 100.0
	}

	return cpuPerc
}

func parseMemory(stats types.MemoryStats) (memUsage float64, memUsagePerc float64) {
	var limit = float64(stats.Limit)

	// add support for cgroup v1
	// cgroup v2
	if v := stats.Stats["inactive_file"]; v < stats.Usage {
		memUsage = float64(stats.Usage - v)
	}

	// in percent
	if limit != 0 { // memLimit, memU
		memUsagePerc = memUsage / float64(limit) * 100
	}

	return memUsage, memUsagePerc
}
