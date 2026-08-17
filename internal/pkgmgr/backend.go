package pkgmgr

import (
	"context"
	"fmt"
	"sync"
)

// Backend is a package manager implementation (pacman now; apt/dnf/flatpak later).
type Backend interface {
	// ID is a stable machine name such as "pacman".
	ID() string
	// Name is a human-readable label such as "Pacman".
	Name() string
	// IsAvailable reports whether this backend is available on the current host.
	IsAvailable() bool
	// List returns every known package (repos + installed + foreign).
	List(ctx context.Context) ([]Package, error)
	// Refresh updates the on-disk package databases. Needs privileges.
	Refresh(ctx context.Context, password string, progress ProgressFunc) error
	// Install installs one package and its dependencies. Needs privileges.
	Install(ctx context.Context, password string, name string, progress ProgressFunc) error
	// Reinstall forces a reinstall of an already-installed package.
	Reinstall(ctx context.Context, password string, name string, progress ProgressFunc) error
	// Remove uninstalls one package. Needs privileges.
	Remove(ctx context.Context, password string, name string, progress ProgressFunc) error
	// Upgrade upgrades one installed package to the repo version.
	Upgrade(ctx context.Context, password string, name string, progress ProgressFunc) error
	// UpgradeAll performs a full system upgrade.
	UpgradeAll(ctx context.Context, password string, progress ProgressFunc) error
}

var (
	registryMu sync.RWMutex
	backends   = make(map[string]Backend)
)

// Register adds a package backend to the registry.
func Register(b Backend) {
	registryMu.Lock()
	defer registryMu.Unlock()
	backends[b.ID()] = b
}

// GetBackend retrieves a backend by ID.
func GetBackend(id string) (Backend, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	b, ok := backends[id]
	return b, ok
}

// AvailableBackends returns all registered backends that are available on this system.
func AvailableBackends() []Backend {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var list []Backend
	for _, b := range backends {
		if b.IsAvailable() {
			list = append(list, b)
		}
	}
	return list
}

// DetectDefaultBackend returns the first available backend.
func DetectDefaultBackend() (Backend, error) {
	avail := AvailableBackends()
	if len(avail) == 0 {
		return nil, fmt.Errorf("no supported package manager backend found on this system")
	}
	return avail[0], nil
}
