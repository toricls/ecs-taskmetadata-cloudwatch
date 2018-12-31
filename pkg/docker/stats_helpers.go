// https://github.com/docker/docker-ce/blob/6ac495401ad144386a089d483539aa8889fa56cc/components/cli/cli/command/container/stats_helpers.go

package docker

import "github.com/docker/docker/api/types"

func CalculateMemUtilization(stats *types.Stats) float64 {
	if stats.MemoryStats.Limit != 0 {
		return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return 0.0
}
