package metrics

import (
	"time"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

// Collector defines the interface for gathering host system metrics
type Collector interface {
	GetOverview() (*protocol.SystemOverview, error)
	GetCPUUsage() (*protocol.CPUInfo, error)
	GetMemoryUsage() (*protocol.MemoryInfo, error)
	GetDiskUsage() ([]protocol.DiskInfo, error)
	GetNetworkUsage() ([]protocol.NetworkInfo, error)
	GetLoadAvg() ([3]float64, error)
	GetUptime() (uint64, error)
	GetOSInfo() (*protocol.OSInfo, error)
}

// NewCollector returns the platform-appropriate metrics collector
func NewCollector() Collector {
	return newPlatformCollector()
}

// BuildOverview compiles metrics from a collector
func BuildOverview(c Collector) (*protocol.SystemOverview, error) {
	osInfo, err := c.GetOSInfo()
	if err != nil {
		osInfo = &protocol.OSInfo{Name: "Unknown", Arch: "unknown"}
	}

	cpuInfo, err := c.GetCPUUsage()
	if err != nil {
		cpuInfo = &protocol.CPUInfo{Cores: 1, UsagePercent: 0}
	}

	memInfo, err := c.GetMemoryUsage()
	if err != nil {
		memInfo = &protocol.MemoryInfo{}
	}

	disks, err := c.GetDiskUsage()
	if err != nil {
		disks = []protocol.DiskInfo{}
	}

	net, err := c.GetNetworkUsage()
	if err != nil {
		net = []protocol.NetworkInfo{}
	}

	load, err := c.GetLoadAvg()
	if err != nil {
		load = [3]float64{0, 0, 0}
	}

	uptime, err := c.GetUptime()
	if err != nil {
		uptime = 0
	}

	return &protocol.SystemOverview{
		UptimeSeconds: uptime,
		OS:            *osInfo,
		CPU:           *cpuInfo,
		Memory:        *memInfo,
		Disks:         disks,
		Network:       net,
		LoadAvg:       load,
		Timestamp:     time.Now().UTC(),
	}, nil
}
