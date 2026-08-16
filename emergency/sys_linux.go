//go:build linux

package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// readLoadavg 读取系统负载。
func readLoadavg() []float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return nil
	}
	var out []float64
	for _, f := range fields[:3] {
		if v, err := strconv.ParseFloat(f, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// readMemory 读取内存信息（/proc/meminfo 前 4 项）。
func readMemory() map[string]any {
	mem := map[string]any{}
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		mem["error"] = err.Error()
		return mem
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valFields := strings.Fields(parts[1])
		if len(valFields) == 0 {
			continue
		}
		if v, err := strconv.ParseInt(valFields[0], 10, 64); err == nil {
			mem[key] = v
		} else {
			mem[key] = valFields[0]
		}
		count++
		if count >= 4 {
			break
		}
	}
	return mem
}

// readDisk 读取磁盘使用量（/ 目录）。
func readDisk() map[string]any {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return map[string]any{"error": err.Error()}
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	used := total - st.Bfree*uint64(st.Bsize)
	return map[string]any{
		"total": total,
		"used":  used,
		"free":  free,
	}
}

// readContainers 读取容器状态。
func readContainers() map[string]any {
	containers := map[string]any{}
	if _, err := exec.LookPath("docker"); err != nil {
		return containers
	}
	for _, name := range []string{"weave-panel", "weave-redis", "weave-frps", "weave-emergency"} {
		out, err := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", name).Output()
		if err != nil {
			containers[name] = "unknown"
			continue
		}
		status := strings.TrimSpace(string(out))
		if status == "" {
			status = "not found"
		}
		containers[name] = status
	}
	return containers
}
