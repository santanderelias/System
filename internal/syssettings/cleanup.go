package syssettings

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"store/internal/privilege"
)

// GetCacheSize computes the total bytes occupied by pacman package cache.
func GetCacheSize() int64 {
	cacheDir := "/var/cache/pacman/pkg"
	var total int64
	_ = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

// GetOrphans returns the list of orphaned package dependencies.
func GetOrphans(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "pacman", "-Qtdq")
	out, err := cmd.Output()
	if err != nil {
		// Exit status 1 means no orphans found
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var orphans []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			orphans = append(orphans, trimmed)
		}
	}
	return orphans, nil
}

// CleanCache cleans old downloaded pacman package cache tarballs.
func CleanCache(ctx context.Context, password string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "pacman", "-Sc", "--noconfirm")
}

// RemoveOrphans removes unneeded orphan dependencies.
func RemoveOrphans(ctx context.Context, password string, orphans []string, progress privilege.ProgressFunc) error {
	if len(orphans) == 0 {
		if progress != nil {
			progress("No orphan packages to remove.")
		}
		return nil
	}
	args := append([]string{"-Rns", "--noconfirm", "--"}, orphans...)
	return privilege.Execute(ctx, password, progress, "pacman", args...)
}

// VacuumJournal cleans systemd journal logs to retain at most 200MB.
func VacuumJournal(ctx context.Context, password string, progress privilege.ProgressFunc) error {
	return privilege.Execute(ctx, password, progress, "journalctl", "--vacuum-size=200M")
}

// GetJournalSize gets the disk usage of journal logs.
func GetJournalSize(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "journalctl", "--disk-usage")
	out, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(bytes.TrimPrefix(out, []byte("Archived and active journals take up "))))
}
