package pkgmgr

import "strings"

// Category represents a top-level software categorization in the store.
type Category string

const (
	CategoryAll          Category = ""
	CategoryDevelopment  Category = "development"
	CategoryInternet     Category = "internet"
	CategoryMultimedia   Category = "multimedia"
	CategoryGraphics     Category = "graphics"
	CategoryProductivity Category = "productivity"
	CategorySystem       Category = "system"
	CategoryUtilities    Category = "utilities"
	CategoryGames        Category = "games"
	CategoryLibraries    Category = "libraries"
)

// CategoryInfo contains display metadata for a category.
type CategoryInfo struct {
	ID          Category
	Name        string
	Description string
	Icon        string
}

// AllCategories returns the list of all browsing categories in order.
var AllCategories = []CategoryInfo{
	{
		ID:          CategoryDevelopment,
		Name:        "Development",
		Description: "IDEs, compilers, debuggers & programming tools",
		Icon:        "code",
	},
	{
		ID:          CategoryInternet,
		Name:        "Internet & Web",
		Description: "Web browsers, messengers, cloud & network tools",
		Icon:        "globe",
	},
	{
		ID:          CategoryMultimedia,
		Name:        "Audio & Video",
		Description: "Media players, music production, video editors",
		Icon:        "media",
	},
	{
		ID:          CategoryGraphics,
		Name:        "Graphics & Design",
		Description: "Photo editing, 3D modeling, vector art & viewers",
		Icon:        "palette",
	},
	{
		ID:          CategoryProductivity,
		Name:        "Productivity & Office",
		Description: "Document suites, notes, calculators, readers",
		Icon:        "document",
	},
	{
		ID:          CategorySystem,
		Name:        "System & Hardware",
		Description: "Kernel, drivers, monitoring, terminal & shells",
		Icon:        "cpu",
	},
	{
		ID:          CategoryUtilities,
		Name:        "Utilities & Tools",
		Description: "Archivers, file managers, CLI helpers & security",
		Icon:        "wrench",
	},
	{
		ID:          CategoryGames,
		Name:        "Games & Emulation",
		Description: "Action, adventure, emulators, gaming platforms",
		Icon:        "gamepad",
	},
	{
		ID:          CategoryLibraries,
		Name:        "Libraries & Runtimes",
		Description: "Shared libraries, GUI bindings & runtime engines",
		Icon:        "cube",
	},
}

// GetCategoryInfo returns the info for a given category ID.
func GetCategoryInfo(id Category) CategoryInfo {
	for _, c := range AllCategories {
		if c.ID == id {
			return c
		}
	}
	return CategoryInfo{
		ID:          id,
		Name:        string(id),
		Description: "Software in category " + string(id),
		Icon:        "cube",
	}
}

// ClassifyPackage determines the most appropriate category for a package
// by inspecting its name, description, groups, and provides.
func ClassifyPackage(p Package) Category {
	name := strings.ToLower(p.Name)
	desc := strings.ToLower(p.Description)
	groups := make([]string, len(p.Groups))
	for i, g := range p.Groups {
		groups[i] = strings.ToLower(g)
	}

	// 1. Check groups first
	for _, g := range groups {
		switch {
		case strings.Contains(g, "devel") || strings.Contains(g, "compiler") || strings.Contains(g, "ide"):
			return CategoryDevelopment
		case strings.Contains(g, "media") || strings.Contains(g, "audio") || strings.Contains(g, "video") || strings.Contains(g, "sound"):
			return CategoryMultimedia
		case strings.Contains(g, "graphics") || strings.Contains(g, "font") || strings.Contains(g, "icon") || strings.Contains(g, "theme"):
			return CategoryGraphics
		case strings.Contains(g, "game") || strings.Contains(g, "emulator"):
			return CategoryGames
		case strings.Contains(g, "system") || strings.Contains(g, "base") || strings.Contains(g, "plasma") || strings.Contains(g, "gnome") || strings.Contains(g, "xfce"):
			return CategorySystem
		}
	}

	// 2. Name prefixes / patterns for libraries
	if strings.HasPrefix(name, "lib32-") || strings.HasPrefix(name, "lib64-") || strings.HasPrefix(name, "lib") && !strings.Contains(desc, "browser") && !strings.Contains(desc, "player") && !strings.Contains(desc, "editor") {
		return CategoryLibraries
	}
	if strings.HasPrefix(name, "python-") || strings.HasPrefix(name, "perl-") || strings.HasPrefix(name, "ruby-") || strings.HasPrefix(name, "rust-") || strings.HasPrefix(name, "go-") {
		if strings.Contains(desc, "library") || strings.Contains(desc, "module") || strings.Contains(desc, "bindings") || strings.Contains(desc, "package") {
			return CategoryLibraries
		}
	}

	// 3. Keyword matching in name + desc
	full := name + " " + desc

	// Development
	if strings.Contains(full, "compiler") || strings.Contains(full, "debugger") ||
		strings.Contains(full, "ide ") || strings.HasSuffix(full, " ide") ||
		strings.Contains(full, "code editor") || strings.Contains(full, "sdk") ||
		strings.Contains(full, "git ") || strings.Contains(full, "linter") ||
		strings.Contains(full, "parser") || strings.Contains(full, "toolchain") ||
		strings.Contains(full, "language") || strings.Contains(full, "golang") ||
		strings.Contains(full, "python") || strings.Contains(full, "rust") ||
		strings.Contains(full, "cmake") || strings.Contains(full, "meson") ||
		strings.Contains(full, "llvm") || strings.Contains(full, "clang") ||
		strings.Contains(full, "neovim") || strings.Contains(full, "emacs") ||
		strings.Contains(full, "docker") || strings.Contains(full, "kubernetes") {
		return CategoryDevelopment
	}

	// Internet
	if strings.Contains(full, "web browser") || strings.Contains(full, "browser") ||
		strings.Contains(full, "http") || strings.Contains(full, "ftp") ||
		strings.Contains(full, "torrent") || strings.Contains(full, "chat") ||
		strings.Contains(full, "matrix") || strings.Contains(full, "irc") ||
		strings.Contains(full, "vpn") || strings.Contains(full, "dns") ||
		strings.Contains(full, "email") || strings.Contains(full, "mail client") ||
		strings.Contains(full, "ssh") || strings.Contains(full, "download manager") ||
		strings.Contains(full, "messenger") || strings.Contains(full, "discord") ||
		strings.Contains(full, "telegram") || strings.Contains(full, "firefox") ||
		strings.Contains(full, "chromium") {
		return CategoryInternet
	}

	// Games
	if strings.Contains(full, "game") || strings.Contains(full, "emulator") ||
		strings.Contains(full, "arcade") || strings.Contains(full, "puzzle") ||
		strings.Contains(full, "rpg") || strings.Contains(full, "steam") ||
		strings.Contains(full, "lutris") || strings.Contains(full, "retroarch") ||
		strings.Contains(full, "chess") || strings.Contains(full, "simulation") {
		return CategoryGames
	}

	// Multimedia
	if strings.Contains(full, "media player") || strings.Contains(full, "audio player") ||
		strings.Contains(full, "video player") || strings.Contains(full, "sound") ||
		strings.Contains(full, "music") || strings.Contains(full, "codec") ||
		strings.Contains(full, "mp3") || strings.Contains(full, "mp4") ||
		strings.Contains(full, "ffmpeg") || strings.Contains(full, "vlc") ||
		strings.Contains(full, "mpv") || strings.Contains(full, "spotify") ||
		strings.Contains(full, "recording") || strings.Contains(full, "synthesizer") ||
		strings.Contains(full, "streaming") || strings.Contains(full, "obs-studio") ||
		strings.Contains(full, "pipewire") || strings.Contains(full, "pulseaudio") {
		return CategoryMultimedia
	}

	// Graphics & Design
	if strings.Contains(full, "image editor") || strings.Contains(full, "photo") ||
		strings.Contains(full, "svg") || strings.Contains(full, "vector") ||
		strings.Contains(full, "3d model") || strings.Contains(full, "rendering") ||
		strings.Contains(full, "gimp") || strings.Contains(full, "inkscape") ||
		strings.Contains(full, "blender") || strings.Contains(full, "drawing") ||
		strings.Contains(full, "paint") || strings.Contains(full, "icon theme") ||
		strings.Contains(full, "font") || strings.Contains(full, "wallpaper") {
		return CategoryGraphics
	}

	// Productivity & Office
	if strings.Contains(full, "office") || strings.Contains(full, "spreadsheet") ||
		strings.Contains(full, "word processor") || strings.Contains(full, "pdf reader") ||
		strings.Contains(full, "document") || strings.Contains(full, "calculator") ||
		strings.Contains(full, "calendar") || strings.Contains(full, "tasks") ||
		strings.Contains(full, "todo") || strings.Contains(full, "libreoffice") ||
		strings.Contains(full, "notes") || strings.Contains(full, "ebook") {
		return CategoryProductivity
	}

	// System
	if strings.Contains(full, "kernel") || strings.Contains(full, "driver") ||
		strings.Contains(full, "firmware") || strings.Contains(full, "bootloader") ||
		strings.Contains(full, "grub") || strings.Contains(full, "systemd") ||
		strings.Contains(full, "file system") || strings.Contains(full, "partition") ||
		strings.Contains(full, "terminal emulator") || strings.Contains(full, "shell") ||
		strings.Contains(full, "display manager") || strings.Contains(full, "window manager") ||
		strings.Contains(full, "wayland") || strings.Contains(full, "xorg") ||
		strings.Contains(full, "benchmark") || strings.Contains(full, "monitoring") ||
		strings.Contains(full, "htop") || strings.Contains(full, "btop") ||
		strings.Contains(full, "package manager") || strings.Contains(full, "pacman") {
		return CategorySystem
	}

	// Utilities
	if strings.Contains(full, "archive") || strings.Contains(full, "compress") ||
		strings.Contains(full, "zip") || strings.Contains(full, "tar") ||
		strings.Contains(full, "backup") || strings.Contains(full, "converter") ||
		strings.Contains(full, "file manager") || strings.Contains(full, "search") ||
		strings.Contains(full, "clipboard") || strings.Contains(full, "screenshot") ||
		strings.Contains(full, "utility") || strings.Contains(full, "tools") ||
		strings.Contains(full, "cli") {
		return CategoryUtilities
	}

	// Default fallback for libraries
	if strings.Contains(full, "library") || strings.Contains(full, "bindings") ||
		strings.Contains(full, "shared object") || strings.Contains(full, "header files") ||
		strings.Contains(full, "api ") || strings.Contains(full, "interface") {
		return CategoryLibraries
	}

	return CategoryUtilities
}
