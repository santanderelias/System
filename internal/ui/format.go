package ui

import (
	"fmt"
	"strings"
	"time"

	"store/internal/pkgmgr"
)

func formatBytes(n int64) string {
	if n <= 0 {
		return "—"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Local().Format("2 Jan 2006")
}

func joinOrDash(vals []string) string {
	if len(vals) == 0 {
		return "—"
	}
	return strings.Join(vals, "  ·  ")
}

func statusLabel(p pkgmgr.Package) string {
	switch {
	case p.UpdateAvailable:
		return "update"
	case p.Foreign:
		return "local"
	case p.Installed:
		return "installed"
	default:
		return p.Source()
	}
}

func statusColor(p pkgmgr.Package) interface{ /* color via helper */ } {
	return nil
}

func metaLine(p pkgmgr.Package) string {
	src := p.Source()
	ver := p.DisplayVersion()
	if p.UpdateAvailable && p.NewVersion != "" {
		if src != "" {
			return src + "  ·  " + ver + "  →  " + p.NewVersion
		}
		return ver + "  →  " + p.NewVersion
	}
	if src != "" && ver != "" {
		return src + "  ·  " + ver
	}
	if ver != "" {
		return ver
	}
	return src
}

func countsLine(s pkgmgr.Stats, shown int, query string, view pkgmgr.View) string {
	base := fmt.Sprintf("%s  ·  %s installed  ·  %s updates",
		plural(s.Total, "package"),
		formatInt(s.Installed),
		formatInt(s.Updates),
	)
	if query != "" || view != pkgmgr.ViewAll {
		base += fmt.Sprintf("  ·  showing %s", formatInt(shown))
	}
	return base
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return formatInt(n) + " " + word + "s"
}

func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
