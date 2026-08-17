package pacman

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"store/internal/pkgmgr"
	"store/internal/privilege"
)

const (
	defaultLocalDB = "/var/lib/pacman/local"
	defaultSyncDB  = "/var/lib/pacman/sync"
	lockFile       = "/var/lib/pacman/db.lck"
)

// Backend talks to pacman: databases for listing, the CLI for mutations.
type Backend struct {
	Pacman  string
	LocalDB string
	SyncDB  string
	Lock    string
}

// New returns a backend pointed at the system pacman databases.
func New() *Backend {
	bin, err := exec.LookPath("pacman")
	if err != nil {
		bin = "pacman"
	}
	return &Backend{
		Pacman:  bin,
		LocalDB: defaultLocalDB,
		SyncDB:  defaultSyncDB,
		Lock:    lockFile,
	}
}

func init() {
	pkgmgr.Register(New())
}

// Available reports whether the pacman binary is on PATH.
func Available() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (b *Backend) ID() string        { return "pacman" }
func (b *Backend) Name() string      { return "Pacman" }
func (b *Backend) IsAvailable() bool { return Available() }

func (b *Backend) List(ctx context.Context) ([]pkgmgr.Package, error) {
	sync, err := readSyncDBs(b.SyncDB)
	if err != nil {
		return nil, err
	}
	local, err := readLocalDB(b.LocalDB)
	if err != nil {
		return nil, err
	}
	updates, err := b.upgrades(ctx)
	if err != nil {
		// Listing still works if the upgrade query fails (offline, empty dbs, …).
		updates = map[string]string{}
	}
	return merge(sync, local, updates), nil
}

func (b *Backend) Refresh(ctx context.Context, password string, progress pkgmgr.ProgressFunc) error {
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-Sy", "--noconfirm", "--color", "never")
}

func (b *Backend) Install(ctx context.Context, password string, name string, progress pkgmgr.ProgressFunc) error {
	if err := requireName(name); err != nil {
		return err
	}
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-S", "--noconfirm", "--needed", "--color", "never", "--", name)
}

func (b *Backend) Reinstall(ctx context.Context, password string, name string, progress pkgmgr.ProgressFunc) error {
	if err := requireName(name); err != nil {
		return err
	}
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-S", "--noconfirm", "--color", "never", "--", name)
}

func (b *Backend) Remove(ctx context.Context, password string, name string, progress pkgmgr.ProgressFunc) error {
	if err := requireName(name); err != nil {
		return err
	}
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-R", "--noconfirm", "--color", "never", "--", name)
}

func (b *Backend) Upgrade(ctx context.Context, password string, name string, progress pkgmgr.ProgressFunc) error {
	if err := requireName(name); err != nil {
		return err
	}
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-S", "--noconfirm", "--color", "never", "--", name)
}

func (b *Backend) UpgradeAll(ctx context.Context, password string, progress pkgmgr.ProgressFunc) error {
	if err := b.checkLock(); err != nil {
		return err
	}
	return privilege.Execute(ctx, password, progress, b.Pacman, "-Syu", "--noconfirm", "--color", "never")
}

func (b *Backend) upgrades(ctx context.Context) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, b.Pacman, "-Qu", "--color", "never")
	cmd.Env = sanitizedEnv()
	out, err := cmd.Output()
	if err != nil {
		// pacman -Qu exits 1 when there is nothing to upgrade on some versions.
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) == 0 && len(out) == 0 {
			return map[string]string{}, nil
		}
		if ee, ok := err.(*exec.ExitError); ok {
			combined := append(append([]byte{}, out...), ee.Stderr...)
			if !bytes.Contains(combined, []byte("error:")) {
				return parseUpgrades(bytes.NewReader(out))
			}
			return nil, fmt.Errorf("pacman -Qu: %s", strings.TrimSpace(string(combined)))
		}
		return nil, err
	}
	return parseUpgrades(bytes.NewReader(out))
}

func (b *Backend) checkLock() error {
	if b.Lock == "" {
		return nil
	}
	if _, err := os.Stat(b.Lock); err == nil {
		return fmt.Errorf("pacman database is locked (%s); another package manager is running", b.Lock)
	}
	return nil
}

func sanitizedEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+2)
	for _, e := range env {
		if strings.HasPrefix(e, "LANG=") || strings.HasPrefix(e, "LC_ALL=") || strings.HasPrefix(e, "LC_MESSAGES=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, "LANG=C", "LC_ALL=C")
	return out
}
