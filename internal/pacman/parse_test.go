package pacman

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"store/internal/pkgmgr"
)

const sampleLocalDesc = `%NAME%
pacman

%VERSION%
7.1.0-2

%DESC%
A library-based package manager with dependency support

%URL%
https://www.archlinux.org/pacman/

%ARCH%
x86_64

%BUILDDATE%
1778057192

%INSTALLDATE%
1782865963

%PACKAGER%
Christian Hesse <eworm@archlinux.org>

%SIZE%
5283285

%REASON%
1

%LICENSE%
GPL-2.0-or-later

%DEPENDS%
bash
glibc

%PROVIDES%
libalpm.so=16-64
`

const sampleSyncDesc = `%FILENAME%
firefox-153.0.4-1-x86_64.pkg.tar.zst

%NAME%
firefox

%VERSION%
153.0.4-1

%DESC%
Fast, Private & Safe Web Browser

%CSIZE%
72000

%ISIZE%
250000000

%URL%
https://www.mozilla.org/firefox/

%LICENSE%
MPL-2.0

%ARCH%
x86_64

%DEPENDS%
gtk3
nss
`

func TestParseDescLocal(t *testing.T) {
	fields, err := parseDesc(strings.NewReader(sampleLocalDesc))
	if err != nil {
		t.Fatal(err)
	}
	p := packageFromFields(fields, "", true)
	if p.Name != "pacman" || p.Version != "7.1.0-2" {
		t.Fatalf("got %s %s", p.Name, p.Version)
	}
	if p.Explicit {
		t.Fatal("REASON 1 should be a dependency")
	}
	if p.InstalledSize != 5283285 {
		t.Fatalf("size = %d", p.InstalledSize)
	}
	if !p.InstallDate.Equal(time.Unix(1782865963, 0)) {
		t.Fatalf("install date = %v", p.InstallDate)
	}
	if len(p.Depends) != 2 || p.Depends[0] != "bash" {
		t.Fatalf("depends = %v", p.Depends)
	}
	if p.InstalledVersion != "7.1.0-2" {
		t.Fatalf("installed version = %s", p.InstalledVersion)
	}
}

func TestParseDescSync(t *testing.T) {
	fields, err := parseDesc(strings.NewReader(sampleSyncDesc))
	if err != nil {
		t.Fatal(err)
	}
	p := packageFromFields(fields, "extra", false)
	if p.Repo != "extra" || p.Installed {
		t.Fatalf("repo/installed = %s %v", p.Repo, p.Installed)
	}
	if p.DownloadSize != 72000 || p.InstalledSize != 250000000 {
		t.Fatalf("sizes %d %d", p.DownloadSize, p.InstalledSize)
	}
	if p.Explicit != true {
		t.Fatal("missing REASON should count as explicit")
	}
}

func TestParseUpgrades(t *testing.T) {
	in := "linux 6.10.1-1 -> 6.10.2-1\nfirefox 153.0.4-1 -> 153.0.5-1\n"
	got, err := parseUpgrades(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if got["linux"] != "6.10.2-1" || got["firefox"] != "153.0.5-1" {
		t.Fatalf("got %#v", got)
	}
}

func TestValidName(t *testing.T) {
	ok := []string{"firefox", "lib32-mesa", "go", "abc@def", "foo.bar", "g++"}
	for _, n := range ok {
		if !ValidName(n) {
			t.Fatalf("%q should be valid", n)
		}
	}
	bad := []string{"", "-evil", "foo bar", "foo/bar", "foo;rm", "../etc"}
	for _, n := range bad {
		if ValidName(n) {
			t.Fatalf("%q should be invalid", n)
		}
	}
}

func TestReadLocalDB(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pacman-7.1.0-2")
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "desc"), []byte(sampleLocalDesc), 0o644); err != nil {
		t.Fatal(err)
	}
	// ALPM_DB_VERSION is a file, not a package dir with desc — should be skipped.
	if err := os.WriteFile(filepath.Join(dir, "ALPM_DB_VERSION"), []byte("9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readLocalDB(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, ok := got["pacman"]
	if !ok {
		t.Fatalf("missing pacman, have %v", got)
	}
	if p.Version != "7.1.0-2" || !p.Installed {
		t.Fatalf("pkg = %+v", p)
	}
}

func TestMergeMarksInstalledAndForeign(t *testing.T) {
	sync := map[string]pkgmgr.Package{
		"firefox": {Name: "firefox", Repo: "extra", Version: "153.0.4-1"},
	}
	local := map[string]pkgmgr.Package{
		"firefox": {Name: "firefox", Version: "153.0.3-1", Installed: true},
		"yay":     {Name: "yay", Version: "12.0-1", Installed: true},
	}

	out := merge(sync, local, map[string]string{"firefox": "153.0.4-1"})
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	var ff, yay bool
	for _, p := range out {
		switch p.Name {
		case "firefox":
			ff = true
			if !p.Installed || p.InstalledVersion != "153.0.3-1" || !p.UpdateAvailable || p.NewVersion != "153.0.4-1" {
				t.Fatalf("firefox = %+v", p)
			}
		case "yay":
			yay = true
			if !p.Foreign || p.Repo != "local" {
				t.Fatalf("yay = %+v", p)
			}
		}
	}
	if !ff || !yay {
		t.Fatal("missing packages after merge")
	}
}

func TestParseDescEmpty(t *testing.T) {
	fields, err := parseDesc(bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 0 {
		t.Fatalf("fields = %#v", fields)
	}
}


