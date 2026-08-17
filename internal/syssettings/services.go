package syssettings

import (
	"bufio"
	"context"
	"os/exec"
	"strings"

	"store/internal/privilege"
)

// ServiceUnit represents a system service / daemon.
type ServiceUnit struct {
	Name        string
	Description string
	LoadState   string
	ActiveState string // "active", "inactive", "failed"
	SubState    string // "running", "dead", "exited"
}

// IsRunning reports whether the service is actively running.
func (s ServiceUnit) IsRunning() bool {
	return s.ActiveState == "active"
}

// ListServices retrieves all known system services.
func ListServices(ctx context.Context) ([]ServiceUnit, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-legend", "--no-pager", "--plain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []ServiceUnit
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		unitName := fields[0]
		load := fields[1]
		active := fields[2]
		sub := fields[3]
		desc := ""
		if len(fields) > 4 {
			desc = strings.Join(fields[4:], " ")
		}

		services = append(services, ServiceUnit{
			Name:        unitName,
			Description: desc,
			LoadState:   load,
			ActiveState: active,
			SubState:    sub,
		})
	}

	return services, nil
}

// StartService starts a system service.
func StartService(ctx context.Context, password, name string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "systemctl", "start", name)
}

// StopService stops a running system service.
func StopService(ctx context.Context, password, name string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "systemctl", "stop", name)
}

// RestartService restarts a system service.
func RestartService(ctx context.Context, password, name string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "systemctl", "restart", name)
}

// EnableService enables a system service at boot.
func EnableService(ctx context.Context, password, name string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "systemctl", "enable", name)
}

// DisableService disables a system service from starting at boot.
func DisableService(ctx context.Context, password, name string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "systemctl", "disable", name)
}
