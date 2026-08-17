package pacman

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"store/internal/pkgmgr"
)

func readLocalDB(dir string) (map[string]pkgmgr.Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read local db %s: %w", dir, err)
	}
	out := make(map[string]pkgmgr.Package, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		descPath := filepath.Join(dir, e.Name(), "desc")
		f, err := os.Open(descPath)
		if err != nil {
			continue
		}
		fields, err := parseDesc(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", descPath, err)
		}
		p := packageFromFields(fields, "", true)
		if p.Name == "" {
			continue
		}
		out[p.Name] = p
	}
	return out, nil
}

func readSyncDBs(dir string) (map[string]pkgmgr.Package, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.db"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no sync databases in %s (refresh the package databases first)", dir)
	}
	out := make(map[string]pkgmgr.Package)
	for _, path := range matches {
		repo := strings.TrimSuffix(filepath.Base(path), ".db")
		if err := readSyncDB(path, repo, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func readSyncDB(path, repo string, into map[string]pkgmgr.Package) error {
	tr, closer, err := openDB(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer closer()

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != "desc" {
			continue
		}
		fields, err := parseDesc(tr)
		if err != nil {
			return fmt.Errorf("parse %s:%s: %w", path, hdr.Name, err)
		}
		p := packageFromFields(fields, repo, false)
		if p.Name == "" {
			continue
		}
		into[p.Name] = p
	}
}

func openDB(path string) (*tar.Reader, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	br := bufio.NewReader(f)
	magic, err := br.Peek(2)
	if err != nil && err != io.EOF {
		f.Close()
		return nil, nil, err
	}
	var r io.Reader = br
	var gz *gzip.Reader
	if len(magic) >= 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err = gzip.NewReader(br)
		if err != nil {
			f.Close()
			return nil, nil, err
		}
		r = gz
	}
	closer := func() {
		if gz != nil {
			_ = gz.Close()
		}
		_ = f.Close()
	}
	return tar.NewReader(r), closer, nil
}

func merge(sync, local map[string]pkgmgr.Package, updates map[string]string) []pkgmgr.Package {
	out := make([]pkgmgr.Package, 0, len(sync)+16)
	for name, p := range sync {
		if l, ok := local[name]; ok {
			p.Installed = true
			p.InstalledVersion = l.Version
			p.InstallDate = l.InstallDate
			p.Explicit = l.Explicit
			if l.InstalledSize > 0 {
				p.InstalledSize = l.InstalledSize
			}
		}
		if nv, ok := updates[name]; ok {
			p.UpdateAvailable = true
			p.NewVersion = nv
		}
		out = append(out, p)
	}
	for name, l := range local {
		if _, ok := sync[name]; ok {
			continue
		}
		l.Foreign = true
		if l.Repo == "" {
			l.Repo = "local"
		}
		if nv, ok := updates[name]; ok {
			l.UpdateAvailable = true
			l.NewVersion = nv
		}
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}
