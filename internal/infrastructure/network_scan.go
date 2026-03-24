package infrastructure

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

var ipRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

func ScanNetworkIPs(ctx context.Context) ([]string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "arp", "-a")
	} else {
		cmd = exec.CommandContext(ctx, "arp", "-a")
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	matches := ipRegex.FindAllString(string(output), -1)
	seen := map[string]struct{}{}
	ips := make([]string, 0, len(matches))
	for _, ip := range matches {
		if strings.HasPrefix(ip, "0.") {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		ips = append(ips, ip)
	}
	slices.Sort(ips)
	return ips, nil
}
