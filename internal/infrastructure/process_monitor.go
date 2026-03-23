package infrastructure

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"superhue/internal/domain"
)

type ProcessMonitor struct {
	interval time.Duration
	lastSeen map[string]struct{}
}

func NewProcessMonitor(interval time.Duration) *ProcessMonitor {
	return &ProcessMonitor{interval: interval, lastSeen: map[string]struct{}{}}
}

func (m *ProcessMonitor) Start(ctx context.Context, sink chan<- domain.RuleEvent) {
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

func (m *ProcessMonitor) poll(ctx context.Context, sink chan<- domain.RuleEvent) {
	current, err := listProcesses(ctx)
	if err != nil {
		return
	}
	for name := range current {
		if _, exists := m.lastSeen[name]; !exists {
			sink <- domain.RuleEvent{Trigger: domain.TriggerProcessStart, Name: "process", Value: name}
		}
	}
	for name := range m.lastSeen {
		if _, exists := current[name]; !exists {
			sink <- domain.RuleEvent{Trigger: domain.TriggerProcessStop, Name: "process", Value: name}
		}
	}
	m.lastSeen = current
}

func listProcesses(ctx context.Context) (map[string]struct{}, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "tasklist", "/FO", "CSV", "/NH")
	} else {
		cmd = exec.CommandContext(ctx, "ps", "-eo", "comm=")
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}
	lines := strings.Split(string(output), "\n")
	result := map[string]struct{}{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if runtime.GOOS == "windows" {
			parts := strings.Split(line, ",")
			if len(parts) == 0 {
				continue
			}
			line = strings.Trim(parts[0], "\"")
		}
		result[strings.ToLower(line)] = struct{}{}
	}
	return result, nil
}
