package pkgmgr

import "testing"

func sample() []Package {
	return []Package{
		{Name: "firefox", Description: "Fast, Private & Safe Web Browser", Repo: "extra", Installed: true, Category: CategoryInternet, InstalledSize: 200000000},
		{Name: "firefox-developer-edition", Description: "Developer Edition of Firefox", Repo: "extra", Category: CategoryInternet, InstalledSize: 220000000},
		{Name: "neovim", Description: "Fork of Vim", Repo: "extra", Installed: true, UpdateAvailable: true, NewVersion: "0.11.0-1", Category: CategoryDevelopment, InstalledSize: 30000000},
		{Name: "vim", Description: "Vi Improved", Repo: "extra", Category: CategoryDevelopment, InstalledSize: 25000000},
		{Name: "pacman", Description: "A library-based package manager", Repo: "core", Installed: true, Provides: []string{"libalpm.so=16-64"}, Category: CategorySystem, InstalledSize: 10000000},
	}
}

func names(pkgs []Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.Name
	}
	return out
}

func TestFilterEmptyQueryKeepsView(t *testing.T) {
	got := Filter(sample(), "", ViewInstalled)
	want := []string{"firefox", "neovim", "pacman"}
	if gotNames := names(got); !equal(gotNames, want) {
		t.Fatalf("installed = %v, want %v", gotNames, want)
	}

	got = Filter(sample(), "", ViewUpdates)
	if gotNames := names(got); !equal(gotNames, []string{"neovim"}) {
		t.Fatalf("updates = %v, want [neovim]", gotNames)
	}

	got = Filter(sample(), "", ViewAvailable)
	if gotNames := names(got); !equal(gotNames, []string{"firefox-developer-edition", "vim"}) {
		t.Fatalf("available = %v, want [firefox-developer-edition vim]", gotNames)
	}
}

func TestFilterCategory(t *testing.T) {
	got := FilterWithOptions(sample(), QueryOptions{
		View:     ViewAll,
		Category: CategoryDevelopment,
	})
	want := []string{"neovim", "vim"}
	if gotNames := names(got); !equal(gotNames, want) {
		t.Fatalf("dev category = %v, want %v", gotNames, want)
	}
}

func TestFilterSortModes(t *testing.T) {
	// Name Desc
	got := FilterWithOptions(sample(), QueryOptions{
		View: ViewAll,
		Sort: SortNameDesc,
	})
	if len(got) == 0 || got[0].Name != "vim" {
		t.Fatalf("sort name desc first = %v, want vim", got[0].Name)
	}

	// Size Desc
	got = FilterWithOptions(sample(), QueryOptions{
		View: ViewAll,
		Sort: SortSizeDesc,
	})
	if len(got) == 0 || got[0].Name != "firefox-developer-edition" {
		t.Fatalf("sort size desc first = %v, want firefox-developer-edition", got[0].Name)
	}
}

func TestFilterRanksNameBeforeDescription(t *testing.T) {
	got := names(Filter(sample(), "vi", ViewAll))
	if len(got) < 2 || got[0] != "vim" {
		t.Fatalf("ranked = %v, want vim first", got)
	}
}

func TestFilterExactNameFirst(t *testing.T) {
	got := names(Filter(sample(), "firefox", ViewAll))
	if len(got) < 2 || got[0] != "firefox" || got[1] != "firefox-developer-edition" {
		t.Fatalf("ranked = %v, want firefox then firefox-developer-edition", got)
	}
}

func TestFilterProvides(t *testing.T) {
	got := names(Filter(sample(), "libalpm.so", ViewAll))
	if !equal(got, []string{"pacman"}) {
		t.Fatalf("provides = %v, want [pacman]", got)
	}
}

func TestFilterNoMatch(t *testing.T) {
	got := Filter(sample(), "does-not-exist-xyz", ViewAll)
	if len(got) != 0 {
		t.Fatalf("got %d hits, want 0", len(got))
	}
}

func TestCountStats(t *testing.T) {
	s := CountStats(sample())
	if s.Total != 5 || s.Installed != 3 || s.Updates != 1 {
		t.Fatalf("stats = %+v", s)
	}
}

func TestClassifyPackage(t *testing.T) {
	p1 := Package{Name: "gcc", Description: "The GNU Compiler Collection - C and C++ frontends"}
	if cat := ClassifyPackage(p1); cat != CategoryDevelopment {
		t.Fatalf("gcc category = %v, want %v", cat, CategoryDevelopment)
	}

	p2 := Package{Name: "firefox", Description: "Fast Web Browser"}
	if cat := ClassifyPackage(p2); cat != CategoryInternet {
		t.Fatalf("firefox category = %v, want %v", cat, CategoryInternet)
	}

	p3 := Package{Name: "vlc", Description: "Multi-platform MPEG, VCD/DVD, and DivX player"}
	if cat := ClassifyPackage(p3); cat != CategoryMultimedia {
		t.Fatalf("vlc category = %v, want %v", cat, CategoryMultimedia)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
