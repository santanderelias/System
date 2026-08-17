package main

import (
	"fmt"
	"os"
	"strings"

	"store/internal/catalog"
	_ "store/internal/pacman" // register pacman backend
	"store/internal/pkgmgr"
	"store/internal/ui"
)

func main() {
	arg := ""
	if len(os.Args) > 1 {
		arg = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	if arg == "-h" || arg == "--help" || arg == "help" {
		printUsage()
		return
	}

	backend, err := pkgmgr.DetectDefaultBackend()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cat := catalog.New(backend)

	switch arg {
	case "apps", "store", "pkg":
		app := ui.NewApp(cat)
		app.Run()
	case "info", "sysinfo", "specs":
		launcher := ui.NewLauncherApp(cat, ui.SectionInfo)
		launcher.Run()
	case "settings", "admin", "services", "maint":
		launcher := ui.NewLauncherApp(cat, ui.SectionSettings)
		launcher.Run()
	default:
		launcher := ui.NewLauncherApp(cat, ui.SectionHub)
		launcher.Run()
	}
}

func printUsage() {
	fmt.Println(`Store — System Administration & Software Management Suite

Usage:
  store [command]

Available Commands:
  (no args)   Launch the full System Administration Suite & Control Center Hub
  apps        Launch the Software Store / Package Manager directly
  info        Launch System Information & Resource Monitor directly
  settings    Launch System Settings, Services & Maintenance directly
  help        Show this help message

Privileges:
  Can be run as a regular user (prompts for password on privileged tasks)
  or directly via sudo (sudo ./store) for elevated execution.`)
}
