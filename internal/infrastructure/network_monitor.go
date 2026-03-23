package infrastructure

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"superhue/internal/domain"
)

type NetworkMonitor struct {
	interval      time.Duration
	failTolerance int
	repo          domain.DeviceRepository
}

func NewNetworkMonitor(repo domain.DeviceRepository, interval time.Duration, failTolerance int) *NetworkMonitor {
	return &NetworkMonitor{interval: interval, failTolerance: failTolerance, repo: repo}
}

func (m *NetworkMonitor) Start(ctx context.Context, sink chan<- domain.RuleEvent) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	m.poll(ctx, sink)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.poll(ctx, sink)
		}
	}
}

func (m *NetworkMonitor) poll(ctx context.Context, sink chan<- domain.RuleEvent) {
	devices, err := m.repo.List(ctx)
	if err != nil {
		return
	}
	for _, device := range devices {
		ok := pingDevice(ctx, device.IP)
		device.LastCheckedAt = time.Now().UTC()
		prev := device.Present
		if ok {
			device.Present = true
			device.FailureCount = 0
			device.ConsecutiveOKs++
			device.LastSeenAt = time.Now().UTC()
		} else {
			device.FailureCount++
			device.ConsecutiveOKs = 0
			if device.FailureCount >= m.failTolerance {
				device.Present = false
			}
		}
		if err := m.repo.UpdateStatus(ctx, device); err != nil {
			continue
		}
		if prev != device.Present {
			value := "absent"
			if device.Present {
				value = "present"
			}
			sink <- domain.RuleEvent{Trigger: domain.TriggerNetworkPresent, Name: device.Name, Value: value}
		}
	}
}

func pingDevice(ctx context.Context, ip string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", "1000", ip)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", ip)
	}
	return cmd.Run() == nil
}
