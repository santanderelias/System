package sysinfo

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// OSInfo holds operating system details.
type OSInfo struct {
	Name            string
	PrettyName      string
	Version         string
	Kernel          string
	Arch            string
	Hostname        string
	DesktopEnv      string
	WindowManager   string
	Uptime          time.Duration
	UptimeFormatted string
}

// CPUInfo holds processor details.
type CPUInfo struct {
	ModelName string
	Cores     int
	Threads   int
	Frequency string
	UsagePct  float64
}

// MemInfo holds memory and swap statistics.
type MemInfo struct {
	TotalBytes     uint64
	AvailableBytes uint64
	UsedBytes      uint64
	UsagePct       float64
	SwapTotalBytes uint64
	SwapFreeBytes  uint64
	SwapUsedBytes  uint64
	SwapUsagePct   float64
}

// DiskInfo holds partition storage details.
type DiskInfo struct {
	MountPoint string
	Device     string
	FSType     string
	TotalBytes uint64
	UsedBytes  uint64
	FreeBytes  uint64
	UsagePct   float64
}

// SystemSnapshot aggregates all system information.
type SystemSnapshot struct {
	OS     OSInfo
	CPU    CPUInfo
	Memory MemInfo
	Disks  []DiskInfo
}

// GetOSInfo collects operating system metadata.
func GetOSInfo() OSInfo {
	var info OSInfo
	info.Arch = runtime.GOARCH

	// Hostname
	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}

	// /etc/os-release
	if f, err := os.Open("/etc/os-release"); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.PrettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
			} else if strings.HasPrefix(line, "NAME=") && info.Name == "" {
				info.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
			} else if strings.HasPrefix(line, "VERSION=") {
				info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION="), `"`)
			}
		}
	}
	if info.PrettyName == "" {
		info.PrettyName = info.Name
	}
	if info.PrettyName == "" {
		info.PrettyName = "Linux"
	}

	// Kernel version
	if data, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		info.Kernel = strings.TrimSpace(string(data))
	} else {
		info.Kernel = runtime.GOOS
	}

	// Desktop Environment / Window Manager
	info.DesktopEnv = os.Getenv("XDG_CURRENT_DESKTOP")
	if info.DesktopEnv == "" {
		info.DesktopEnv = os.Getenv("DESKTOP_SESSION")
	}
	if info.DesktopEnv == "" {
		info.DesktopEnv = "Headless / Unknown"
	}

	if wm := os.Getenv("XDG_SESSION_TYPE"); wm != "" {
		info.WindowManager = wm
	} else {
		info.WindowManager = "X11/Wayland"
	}

	// Uptime from /proc/uptime
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if secs, err := strconv.ParseFloat(fields[0], 64); err == nil {
				info.Uptime = time.Duration(secs) * time.Second
				info.UptimeFormatted = formatDuration(info.Uptime)
			}
		}
	}

	return info
}

// GetCPUInfo collects CPU hardware specifications.
func GetCPUInfo() CPUInfo {
	var info CPUInfo
	info.Threads = runtime.NumCPU()

	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "model name") && info.ModelName == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.ModelName = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "cpu MHz") && info.Frequency == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if mhz, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						info.Frequency = fmt.Sprintf("%.2f GHz", mhz/1000.0)
					}
				}
			} else if strings.HasPrefix(line, "cpu cores") && info.Cores == 0 {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if c, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						info.Cores = c
					}
				}
			}
		}
	}

	if info.Cores == 0 {
		info.Cores = info.Threads
	}
	if info.ModelName == "" {
		info.ModelName = fmt.Sprintf("Generic %s Processor (%d Threads)", runtime.GOARCH, info.Threads)
	}

	return info
}

// GetMemInfo collects RAM and Swap status from /proc/meminfo.
func GetMemInfo() MemInfo {
	var info MemInfo
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return info
	}
	defer f.Close()

	var memTotal, memAvailable, swapTotal, swapFree uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		valBytes := val * 1024 // /proc/meminfo is in kB

		switch parts[0] {
		case "MemTotal:":
			memTotal = valBytes
		case "MemAvailable:":
			memAvailable = valBytes
		case "SwapTotal:":
			swapTotal = valBytes
		case "SwapFree:":
			swapFree = valBytes
		}
	}

	info.TotalBytes = memTotal
	info.AvailableBytes = memAvailable
	if memTotal > memAvailable {
		info.UsedBytes = memTotal - memAvailable
	}
	if memTotal > 0 {
		info.UsagePct = (float64(info.UsedBytes) / float64(memTotal)) * 100.0
	}

	info.SwapTotalBytes = swapTotal
	info.SwapFreeBytes = swapFree
	if swapTotal > swapFree {
		info.SwapUsedBytes = swapTotal - swapFree
	}
	if swapTotal > 0 {
		info.SwapUsagePct = (float64(info.SwapUsedBytes) / float64(swapTotal)) * 100.0
	}

	return info
}

// GetDisksInfo collects mounted disk partitions and usage.
func GetDisksInfo() []DiskInfo {
	var disks []DiskInfo
	seen := make(map[string]bool)

	f, err := os.Open("/proc/mounts")
	if err != nil {
		return disks
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 {
			continue
		}
		device, mount, fstype := fields[0], fields[1], fields[2]

		// Filter for real physical / storage filesystems
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		if strings.HasPrefix(mount, "/var/lib/docker") || strings.HasPrefix(mount, "/var/lib/flatpak") {
			continue
		}
		if seen[mount] {
			continue
		}
		seen[mount] = true

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		if total == 0 {
			continue
		}
		used := total - free
		pct := (float64(used) / float64(total)) * 100.0

		disks = append(disks, DiskInfo{
			MountPoint: mount,
			Device:     device,
			FSType:     fstype,
			TotalBytes: total,
			UsedBytes:  used,
			FreeBytes:  free,
			UsagePct:   pct,
		})
	}

	return disks
}

// GetSnapshot captures an instantaneous snapshot of system state.
func GetSnapshot() SystemSnapshot {
	return SystemSnapshot{
		OS:     GetOSInfo(),
		CPU:    GetCPUInfo(),
		Memory: GetMemInfo(),
		Disks:  GetDisksInfo(),
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d days, %d hrs, %d mins", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hrs, %d mins", hours, mins)
	}
	return fmt.Sprintf("%d mins", mins)
}
