// https://github.com/docker/docker-ce/blob/6ac495401ad144386a089d483539aa8889fa56cc/components/cli/cli/command/container/stats_helpers.go

package docker

import "github.com/docker/docker/api/types"

func CalculateMemUtilization(stats *types.Stats) float64 {
	if stats.MemoryStats.Limit != 0 {
		return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return 0.0
}

func CalculateCpuUtilization(stats *types.Stats) float64 {
	var (
		prevCPU    = stats.PreCPUStats.CPUUsage.TotalUsage
		prevSystem = stats.PreCPUStats.SystemUsage

		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(prevCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(stats.CPUStats.SystemUsage) - float64(prevSystem)
		onlineCPUs  = float64(stats.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}
