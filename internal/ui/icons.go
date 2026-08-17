package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"store/internal/pkgmgr"
)

// IconForCategory returns the appropriate vector icon for a category.
func IconForCategory(cat pkgmgr.Category) fyne.Resource {
	switch cat {
	case pkgmgr.CategoryDevelopment:
		return theme.ComputerIcon()
	case pkgmgr.CategoryInternet:
		return theme.NavigateNextIcon()
	case pkgmgr.CategoryMultimedia:
		return theme.MediaPlayIcon()
	case pkgmgr.CategoryGraphics:
		return theme.ColorPaletteIcon()
	case pkgmgr.CategoryProductivity:
		return theme.DocumentIcon()
	case pkgmgr.CategorySystem:
		return theme.SettingsIcon()
	case pkgmgr.CategoryUtilities:
		return theme.ContentCutIcon()
	case pkgmgr.CategoryGames:
		return theme.MediaFastForwardIcon()
	case pkgmgr.CategoryLibraries:
		return theme.StorageIcon()
	default:
		return theme.FolderIcon()
	}
}

// IconForPackage returns a vector icon based on package name or category.
func IconForPackage(p pkgmgr.Package) fyne.Resource {
	n := strings.ToLower(p.Name)
	switch {
	case strings.Contains(n, "browser") || strings.Contains(n, "firefox") || strings.Contains(n, "chromium"):
		return theme.NavigateNextIcon()
	case strings.Contains(n, "media") || strings.Contains(n, "vlc") || strings.Contains(n, "mpv") || strings.Contains(n, "audio") || strings.Contains(n, "music"):
		return theme.MediaPlayIcon()
	case strings.Contains(n, "edit") || strings.Contains(n, "code") || strings.Contains(n, "vim") || strings.Contains(n, "ide"):
		return theme.DocumentCreateIcon()
	case strings.Contains(n, "term") || strings.Contains(n, "shell") || strings.Contains(n, "console"):
		return theme.ComputerIcon()
	case strings.Contains(n, "clean") || strings.Contains(n, "trash") || strings.Contains(n, "remove"):
		return theme.DeleteIcon()
	default:
		return IconForCategory(p.Category)
	}
}

// IconForView returns the navigation icon for a main view.
func IconForView(v pkgmgr.View) fyne.Resource {
	switch v {
	case pkgmgr.ViewDiscover:
		return theme.HomeIcon()
	case pkgmgr.ViewInstalled:
		return theme.ConfirmIcon()
	case pkgmgr.ViewUpdates:
		return theme.ViewRefreshIcon()
	case pkgmgr.ViewAvailable:
		return theme.DownloadIcon()
	default:
		return theme.ListIcon()
	}
}
