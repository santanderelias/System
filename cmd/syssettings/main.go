package main

import (
	"fmt"
	"os"

	"store/internal/catalog"
	_ "store/internal/pacman"
	"store/internal/pkgmgr"
	"store/internal/ui"
)

func main() {
	backend, err := pkgmgr.DetectDefaultBackend()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cat := catalog.New(backend)
	launcher := ui.NewLauncherApp(cat, ui.SectionSettings)
	launcher.Run()
}
