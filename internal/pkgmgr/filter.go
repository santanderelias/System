package pkgmgr

import (
	"sort"
	"strings"
)

// QueryOptions contains parameters for searching and filtering the package catalog.
type QueryOptions struct {
	Query    string
	View     View
	Category Category
	Sort     SortMode
}

// FilterWithOptions returns packages matching view, category, and query, sorted according to Sort.
func FilterWithOptions(pkgs []Package, opts QueryOptions) []Package {
	q := strings.ToLower(strings.TrimSpace(opts.Query))
	isQueryEmpty := q == ""

	type scored struct {
		pkg   Package
		score int
	}

	items := make([]scored, 0, 128)

	for _, p := range pkgs {
		// View filter
		if !opts.View.Match(p) {
			continue
		}

		// Category filter
		if opts.Category != CategoryAll && p.Category != opts.Category {
			continue
		}

		// Search Query filter
		if !isQueryEmpty {
			s := score(p, q)
			if s < 0 {
				continue
			}
			items = append(items, scored{pkg: p, score: s})
		} else {
			items = append(items, scored{pkg: p, score: 0})
		}
	}

	// Sorting
	switch opts.Sort {
	case SortNameAsc:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].pkg.Name) < strings.ToLower(items[j].pkg.Name)
		})
	case SortNameDesc:
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].pkg.Name) > strings.ToLower(items[j].pkg.Name)
		})
	case SortSizeDesc:
		sort.SliceStable(items, func(i, j int) bool {
			si := items[i].pkg.InstalledSize
			if si == 0 {
				si = items[i].pkg.DownloadSize
			}
			sj := items[j].pkg.InstalledSize
			if sj == 0 {
				sj = items[j].pkg.DownloadSize
			}
			return si > sj
		})
	case SortStatus:
		sort.SliceStable(items, func(i, j int) bool {
			return statusRank(items[i].pkg) < statusRank(items[j].pkg)
		})
	default: // SortRelevance
		if !isQueryEmpty {
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].score != items[j].score {
					return items[i].score < items[j].score
				}
				return strings.ToLower(items[i].pkg.Name) < strings.ToLower(items[j].pkg.Name)
			})
		} else {
			// Default to A-Z when no query
			sort.SliceStable(items, func(i, j int) bool {
				return strings.ToLower(items[i].pkg.Name) < strings.ToLower(items[j].pkg.Name)
			})
		}
	}

	out := make([]Package, len(items))
	for i, it := range items {
		out[i] = it.pkg
	}
	return out
}

func statusRank(p Package) int {
	if p.UpdateAvailable {
		return 0
	}
	if p.Installed {
		return 1
	}
	return 2
}

// Filter returns packages matching view and query (retaining backward compatibility).
func Filter(pkgs []Package, query string, view View) []Package {
	return FilterWithOptions(pkgs, QueryOptions{
		Query: query,
		View:  view,
		Sort:  SortRelevance,
	})
}

// CountStats tallies a full (unfiltered) catalog.
func CountStats(pkgs []Package) Stats {
	var s Stats
	s.Total = len(pkgs)
	for _, p := range pkgs {
		if p.Installed {
			s.Installed++
		}
		if p.UpdateAvailable {
			s.Updates++
		}
		if p.Foreign {
			s.Foreign++
		}
	}
	return s
}

// score ranks a hit. Lower is better. -1 means no match.
func score(p Package, q string) int {
	name := strings.ToLower(p.Name)
	switch {
	case name == q:
		return 0
	case strings.HasPrefix(name, q):
		return 1
	case strings.Contains(name, q):
		return 2
	}

	best := -1
	for _, prov := range p.Provides {
		pn := strings.ToLower(provideName(prov))
		switch {
		case pn == q || strings.HasPrefix(pn, q):
			return 2
		case strings.Contains(pn, q):
			best = 3
		}
	}
	if strings.Contains(strings.ToLower(p.Description), q) {
		return 3
	}
	if strings.Contains(strings.ToLower(p.Repo), q) {
		if best < 0 {
			return 4
		}
	}
	return best
}

func provideName(prov string) string {
	if i := strings.IndexAny(prov, "=<"); i >= 0 {
		return prov[:i]
	}
	return prov
}
