package pkgmgr

import "time"

// View is a catalog filter shown in the UI navigation.
type View int

const (
	ViewDiscover View = iota
	ViewAll
	ViewInstalled
	ViewAvailable
	ViewUpdates
)

func (v View) String() string {
	switch v {
	case ViewDiscover:
		return "Discover"
	case ViewInstalled:
		return "Installed"
	case ViewAvailable:
		return "Available"
	case ViewUpdates:
		return "Updates"
	default:
		return "All Packages"
	}
}

// Match reports whether p belongs in this view.
func (v View) Match(p Package) bool {
	switch v {
	case ViewInstalled:
		return p.Installed
	case ViewAvailable:
		return !p.Installed
	case ViewUpdates:
		return p.UpdateAvailable
	default:
		return true
	}
}

// SortMode specifies ordering of search/filter results.
type SortMode int

const (
	SortRelevance SortMode = iota
	SortNameAsc
	SortNameDesc
	SortSizeDesc
	SortStatus
)

func (s SortMode) String() string {
	switch s {
	case SortNameAsc:
		return "Name (A–Z)"
	case SortNameDesc:
		return "Name (Z–A)"
	case SortSizeDesc:
		return "Largest Size"
	case SortStatus:
		return "Status"
	default:
		return "Relevance"
	}
}

// Package is a backend-agnostic software package.
type Package struct {
	Name             string
	Version          string
	InstalledVersion string
	NewVersion       string
	Description      string
	Repo             string
	Arch             string
	URL              string
	Packager         string
	Licenses         []string
	Depends          []string
	OptDepends       []string
	Provides         []string
	Conflicts        []string
	Replaces         []string
	Groups           []string
	DownloadSize     int64
	InstalledSize    int64
	BuildDate        time.Time
	InstallDate      time.Time
	Installed        bool
	Explicit         bool
	Foreign          bool
	UpdateAvailable  bool
	Category         Category
}

// DisplayVersion is the version the list should show as current.
func (p Package) DisplayVersion() string {
	if p.Installed && p.InstalledVersion != "" {
		return p.InstalledVersion
	}
	return p.Version
}

// Source is a short origin label (repo name, or "local" for foreign packages).
func (p Package) Source() string {
	if p.Repo != "" {
		return p.Repo
	}
	if p.Foreign {
		return "local"
	}
	return ""
}

// Stats is a snapshot of catalog counts.
type Stats struct {
	Total     int
	Installed int
	Updates   int
	Foreign   int
}

// ProgressFunc receives one line of privileged-command output.
type ProgressFunc func(line string)
