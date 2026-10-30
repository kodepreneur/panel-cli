//go:build linux

package metrics

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type linuxCollector struct{}

func newPlatformCollector() Collector {
	return &linuxCollector{}
}

func (c *linuxCollector) GetOverview() (*protocol.SystemOverview, error) {
	return BuildOverview(c)
}

func (c *linuxCollector) GetOSInfo() (*protocol.OSInfo, error) {
	hostname, _ := os.Hostname()

	name := "Linux"
	version := "Ubuntu"
	if f, err := os.Open("/etc/os-release"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				version = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			}
			if strings.HasPrefix(line, "NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			}
		}
	}

	kernel := "Linux"
	if data, err := os.ReadFile("/proc/version"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			kernel = fields[2]
		}
	}

	return &protocol.OSInfo{
		Name:     name,
		Version:  version,
		Kernel:   kernel,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}, nil
}

func (c *linuxCollector) GetCPUUsage() (*protocol.CPUInfo, error) {
	cores := runtime.NumCPU()
	model := "Generic CPU"

	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Calculate CPU usage percentage from /proc/stat
	usage := 0.0
	if data, err := os.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu ") {
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					var user, nice, system, idle, iowait, irq, softirq, steal float64
					user, _ = strconv.ParseFloat(fields[1], 64)
					nice, _ = strconv.ParseFloat(fields[2], 64)
					system, _ = strconv.ParseFloat(fields[3], 64)
					idle, _ = strconv.ParseFloat(fields[4], 64)
					if len(fields) >= 6 {
						iowait, _ = strconv.ParseFloat(fields[5], 64)
					}
					if len(fields) >= 7 {
						irq, _ = strconv.ParseFloat(fields[6], 64)
					}
					if len(fields) >= 8 {
						softirq, _ = strconv.ParseFloat(fields[7], 64)
					}
					if len(fields) >= 9 {
						steal, _ = strconv.ParseFloat(fields[8], 64)
					}

					busy := user + nice + system + irq + softirq + steal
					total := busy + idle + iowait
					if total > 0 {
						usage = (busy / total) * 100.0
					}
				}
				break
			}
		}
	}

	return &protocol.CPUInfo{
		Model:        model,
		Cores:        cores,
		UsagePercent: usage,
	}, nil
}

func (c *linuxCollector) GetMemoryUsage() (*protocol.MemoryInfo, error) {
	var memTotal, memFree, memAvailable, buffers, cached uint64
	var swapTotal, swapFree uint64

	if f, err := os.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			valBytes := val * 1024 // convert kB to bytes

			switch fields[0] {
			case "MemTotal:":
				memTotal = valBytes
			case "MemFree:":
				memFree = valBytes
			case "MemAvailable:":
				memAvailable = valBytes
			case "Buffers:":
				buffers = valBytes
			case "Cached:":
				cached = valBytes
			case "SwapTotal:":
				swapTotal = valBytes
			case "SwapFree:":
				swapFree = valBytes
			}
		}
	}

	used := memTotal - memAvailable
	if memAvailable == 0 {
		used = memTotal - memFree - buffers - cached
	}

	usagePct := 0.0
	if memTotal > 0 {
		usagePct = (float64(used) / float64(memTotal)) * 100.0
	}

	swapUsed := swapTotal - swapFree

	return &protocol.MemoryInfo{
		TotalBytes:     memTotal,
		UsedBytes:      used,
		FreeBytes:      memTotal - used,
		UsagePercent:   usagePct,
		SwapTotalBytes: swapTotal,
		SwapUsedBytes:  swapUsed,
	}, nil
}

func (c *linuxCollector) GetDiskUsage() ([]protocol.DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return []protocol.DiskInfo{}, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	usagePct := 0.0
	if total > 0 {
		usagePct = float64(used) / float64(total) * 100.0
	}

	return []protocol.DiskInfo{
		{
			Mount:        "/",
			Filesystem:   "ext4",
			TotalBytes:   total,
			UsedBytes:    used,
			FreeBytes:    free,
			UsagePercent: usagePct,
		},
	}, nil
}

func (c *linuxCollector) GetNetworkUsage() ([]protocol.NetworkInfo, error) {
	var result []protocol.NetworkInfo

	if f, err := os.Open("/proc/net/dev"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.Contains(line, ":") {
				continue
			}
			parts := strings.Split(line, ":")
			iface := strings.TrimSpace(parts[0])
			if iface == "lo" {
				continue
			}
			fields := strings.Fields(parts[1])
			if len(fields) >= 9 {
				rxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
				txBytes, _ := strconv.ParseUint(fields[8], 10, 64)
				result = append(result, protocol.NetworkInfo{
					Interface: iface,
					RxBytes:   rxBytes,
					TxBytes:   txBytes,
					RxRateSec: 0,
					TxRateSec: 0,
				})
			}
		}
	}

	if len(result) == 0 {
		result = append(result, protocol.NetworkInfo{
			Interface: "eth0",
			RxBytes:   0,
			TxBytes:   0,
		})
	}

	return result, nil
}

func (c *linuxCollector) GetLoadAvg() ([3]float64, error) {
	var loads [3]float64
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		for i := 0; i < 3 && i < len(fields); i++ {
			loads[i], _ = strconv.ParseFloat(fields[i], 64)
		}
	}
	return loads, nil
}

func (c *linuxCollector) GetUptime() (uint64, error) {
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			upFloat, _ := strconv.ParseFloat(fields[0], 64)
			return uint64(upFloat), nil
		}
	}
	return 0, nil
}
