package pacman

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"store/internal/pkgmgr"
)

// parseDesc reads an alpm desc file into field -> values.
func parseDesc(r io.Reader) (map[string][]string, error) {
	fields := make(map[string][]string)
	sc := bufio.NewScanner(r)
	// PGPSIG / long provides lists can exceed the default 64KiB token.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var key string
	for sc.Scan() {
		line := sc.Text()
		if len(line) >= 3 && line[0] == '%' && line[len(line)-1] == '%' {
			key = line[1 : len(line)-1]
			continue
		}
		if key == "" {
			continue
		}
		if line == "" {
			key = ""
			continue
		}
		fields[key] = append(fields[key], line)
	}
	return fields, sc.Err()
}

func first(fields map[string][]string, key string) string {
	if vs := fields[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func packageFromFields(fields map[string][]string, repo string, installed bool) pkgmgr.Package {
	p := pkgmgr.Package{
		Name:        first(fields, "NAME"),
		Version:     first(fields, "VERSION"),
		Description: first(fields, "DESC"),
		Repo:        repo,
		Arch:        first(fields, "ARCH"),
		URL:         first(fields, "URL"),
		Packager:    first(fields, "PACKAGER"),
		Licenses:    fields["LICENSE"],
		Depends:     fields["DEPENDS"],
		OptDepends:  fields["OPTDEPENDS"],
		Provides:    fields["PROVIDES"],
		Conflicts:   fields["CONFLICTS"],
		Replaces:    fields["REPLACES"],
		Groups:      fields["GROUPS"],
		Installed:   installed,
		Explicit:    first(fields, "REASON") != "1",
	}
	if installed {
		p.InstalledVersion = p.Version
	}
	p.DownloadSize = parseInt64(first(fields, "CSIZE"))
	if n := parseInt64(first(fields, "ISIZE")); n > 0 {
		p.InstalledSize = n
	} else {
		p.InstalledSize = parseInt64(first(fields, "SIZE"))
	}
	if n := parseInt64(first(fields, "BUILDDATE")); n > 0 {
		p.BuildDate = time.Unix(n, 0)
	}
	if n := parseInt64(first(fields, "INSTALLDATE")); n > 0 {
		p.InstallDate = time.Unix(n, 0)
	}
	p.Category = pkgmgr.ClassifyPackage(p)
	return p
}

// parseUpgrades parses `pacman -Qu` output: "name oldver -> newver".
func parseUpgrades(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "error:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[2] == "->" && ValidName(parts[0]) {
			out[parts[0]] = parts[3]
		}
	}
	return out, sc.Err()
}

// ValidName reports whether name is a safe pacman package identifier.
func ValidName(name string) bool {
	if name == "" || name[0] == '-' {
		return false
	}
	for _, r := range name {
		if !validNameRune(r) {
			return false
		}
	}
	return true
}

func validNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		r == '@' || r == '.' || r == '_' || r == '+' || r == '-'
}

func requireName(name string) error {
	if !ValidName(name) {
		return fmt.Errorf("invalid package name %q", name)
	}
	return nil
}
