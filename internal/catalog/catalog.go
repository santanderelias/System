package catalog

import (
	"context"
	"sync"

	"store/internal/pkgmgr"
)

// FeaturedPackageNames lists well-known, popular Linux applications for the Discover view spotlight.
var FeaturedPackageNames = []string{
	"firefox", "chromium", "vlc", "mpv", "gimp", "inkscape", "blender", "obs-studio",
	"code", "neovim", "git", "docker", "htop", "btop", "kitty", "alacritty",
	"steam", "lutris", "retroarch", "libreoffice-fresh", "thunderbird", "discord",
	"telegram-desktop", "spotify-launcher", "audacity", "kdenlive", "wireshark-qt",
	"fastfetch", "zellij", "tmux", "ffmpeg",
}

// Catalog caches the last successful backend listing and serves filtered queries.
type Catalog struct {
	backend pkgmgr.Backend

	mu   sync.RWMutex
	pkgs []pkgmgr.Package
}

// New creates a new Catalog for a backend.
func New(backend pkgmgr.Backend) *Catalog {
	return &Catalog{backend: backend}
}

// Backend returns the current package manager backend.
func (c *Catalog) Backend() pkgmgr.Backend {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.backend
}

// SetBackend changes the backend and clears old packages.
func (c *Catalog) SetBackend(b pkgmgr.Backend) {
	c.mu.Lock()
	c.backend = b
	c.pkgs = nil
	c.mu.Unlock()
}

// Reload replaces the in-memory catalog from the backend.
func (c *Catalog) Reload(ctx context.Context) error {
	c.mu.RLock()
	b := c.backend
	c.mu.RUnlock()

	pkgs, err := b.List(ctx)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.pkgs = pkgs
	c.mu.Unlock()
	return nil
}

// Snapshot returns a copy of the current catalog.
func (c *Catalog) Snapshot() []pkgmgr.Package {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]pkgmgr.Package, len(c.pkgs))
	copy(out, c.pkgs)
	return out
}

// Query filters the current catalog by string query and view.
func (c *Catalog) Query(query string, view pkgmgr.View) []pkgmgr.Package {
	return pkgmgr.Filter(c.Snapshot(), query, view)
}

// QueryWithOptions filters and sorts the current catalog.
func (c *Catalog) QueryWithOptions(opts pkgmgr.QueryOptions) []pkgmgr.Package {
	return pkgmgr.FilterWithOptions(c.Snapshot(), opts)
}

// Lookup finds a package by exact name.
func (c *Catalog) Lookup(name string) (pkgmgr.Package, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.pkgs {
		if p.Name == name {
			return p, true
		}
	}
	return pkgmgr.Package{}, false
}

// Stats computes snapshot stats.
func (c *Catalog) Stats() pkgmgr.Stats {
	return pkgmgr.CountStats(c.Snapshot())
}

// CategoryCounts returns the count of packages per category.
func (c *Catalog) CategoryCounts() map[pkgmgr.Category]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	counts := make(map[pkgmgr.Category]int)
	for _, p := range c.pkgs {
		counts[p.Category]++
	}
	return counts
}

// Featured returns packages that match the featured showcase list and exist in catalog.
func (c *Catalog) Featured() []pkgmgr.Package {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []pkgmgr.Package
	pkgMap := make(map[string]pkgmgr.Package, len(c.pkgs))
	for _, p := range c.pkgs {
		pkgMap[p.Name] = p
	}
	for _, name := range FeaturedPackageNames {
		if p, ok := pkgMap[name]; ok {
			out = append(out, p)
		}
	}
	return out
}
