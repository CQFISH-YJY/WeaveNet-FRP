//go:build !linux

package main

import (
	"os/exec"
	"runtime"
	"strings"
)

// readLoadavg 非 Linux 平台不支持。
func readLoadavg() []float64 {
	return nil
}

// readMemory 非 Linux 平台暂不采集。
func readMemory() map[string]any {
	return map[string]any{"error": "unsupported platform: " + runtime.GOOS}
}

// readDisk 非 Linux 平台暂不采集。
func readDisk() map[string]any {
	return map[string]any{"error": "unsupported platform: " + runtime.GOOS}
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
