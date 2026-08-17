package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/catalog"
	"store/internal/pkgmgr"
	"store/internal/privilege"
)

// App manages the application UI state and lifecycle.
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	catalog *catalog.Catalog

	activeView     pkgmgr.View
	activeCategory pkgmgr.Category
	activeSort     pkgmgr.SortMode
	searchQuery    string

	filteredPkgs []pkgmgr.Package
	mu           sync.Mutex
	initialized  bool

	// UI Components
	sidebarContainer *container.Scroll
	searchEntry      *widget.Entry
	sortSelect       *widget.Select
	viewTitle        *canvas.Text
	refreshBtn       *widget.Button
	upgradeAllBtn    *widget.Button
	statusLabel      *widget.Label
	backendPill      *canvas.Text

	installedBadge *canvas.Text
	updatesBadge   *canvas.Text

	pkgList        *widget.List
	detailPane     *detailPane
	discoverView   *discoverView
	actionModal    *actionModal
	passwordDialog *passwordDialog

	mainStack *fyne.Container
	listSplit *container.Split

	loadingOverlay *fyne.Container
	loadingText    *widget.Label
	loadingSpinner *widget.ProgressBarInfinite
}

// NewApp initializes the Store application.
func NewApp(cat *catalog.Catalog) *App {
	a := app.NewWithID("dev.store.app")
	a.Settings().SetTheme(&storeTheme{})

	w := a.NewWindow("Store — Universal Package Manager")
	w.Resize(fyne.NewSize(1100, 720))
	w.SetMaster()

	ui := &App{
		fyneApp:        a,
		window:         w,
		catalog:        cat,
		activeView:     pkgmgr.ViewDiscover,
		activeCategory: pkgmgr.CategoryAll,
		activeSort:     pkgmgr.SortRelevance,
	}

	ui.initUI()
	return ui
}

func (ui *App) initUI() {
	// Status & Badge Labels initialized first
	ui.statusLabel = widget.NewLabel("Initializing...")
	ui.statusLabel.TextStyle = fyne.TextStyle{Italic: true}
	ui.statusLabel.Truncation = fyne.TextTruncateEllipsis

	ui.installedBadge = canvas.NewText("", installedColor())
	ui.installedBadge.TextSize = theme.CaptionTextSize()

	ui.updatesBadge = canvas.NewText("", updateColor())
	ui.updatesBadge.TextSize = theme.CaptionTextSize()

	ui.viewTitle = canvas.NewText("Discover", theme.Color(theme.ColorNameForeground))
	ui.viewTitle.TextStyle = fyne.TextStyle{Bold: true}
	ui.viewTitle.TextSize = 18

	backendName := "Pacman"
	if b := ui.catalog.Backend(); b != nil {
		backendName = b.Name()
	}
	ui.backendPill = canvas.NewText(backendName+" Engine", brandAccent())
	ui.backendPill.TextSize = theme.CaptionTextSize()

	// Modals
	ui.actionModal = newActionModal(ui.window)
	ui.passwordDialog = newPasswordDialog(ui.window)

	// Detail Pane
	ui.detailPane = newDetailPane(func(action string, p pkgmgr.Package) {
		ui.promptAction(action, p)
	})

	// Discover View
	ui.discoverView = newDiscoverView(
		func(p pkgmgr.Package) {
			ui.showPackageDetail(p)
		},
		func(c pkgmgr.Category) {
			ui.setCategory(c)
		},
		func(action string, p pkgmgr.Package) {
			ui.promptAction(action, p)
		},
	)

	// Package List Widget
	ui.pkgList = widget.NewList(
		func() int {
			ui.mu.Lock()
			defer ui.mu.Unlock()
			return len(ui.filteredPkgs)
		},
		func() fyne.CanvasObject {
			return newPackageRow(
				func(action string, p pkgmgr.Package) {
					ui.promptAction(action, p)
				},
				func(p pkgmgr.Package) {
					ui.showPackageDetail(p)
				},
			)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			ui.mu.Lock()
			if id < 0 || id >= len(ui.filteredPkgs) {
				ui.mu.Unlock()
				return
			}
			p := ui.filteredPkgs[id]
			ui.mu.Unlock()

			if row, ok := o.(*packageRow); ok {
				row.set(p)
			}
		},
	)

	ui.pkgList.OnSelected = func(id widget.ListItemID) {
		ui.mu.Lock()
		if id >= 0 && id < len(ui.filteredPkgs) {
			p := ui.filteredPkgs[id]
			ui.mu.Unlock()
			ui.detailPane.show(p)
		} else {
			ui.mu.Unlock()
		}
	}

	// List + Detail Split View
	ui.listSplit = container.NewHSplit(ui.pkgList, ui.detailPane.object())
	ui.listSplit.SetOffset(0.48)

	// Main Stack (Switch between Discover and List)
	ui.mainStack = container.NewStack(ui.discoverView.object(), ui.listSplit)

	// Top Bar
	topBar := ui.buildTopBar()

	// Sidebar
	sidebar := ui.buildSidebar()

	var authTag string
	if privilege.IsRoot() {
		authTag = "Root User (Elevated)"
	} else {
		authTag = "Normal User (Password on action)"
	}
	authLabel := widget.NewLabel(authTag)
	footer := container.NewBorder(nil, nil, ui.statusLabel, authLabel)

	// Loading Overlay
	ui.loadingText = widget.NewLabel("Loading package database...")
	ui.loadingSpinner = widget.NewProgressBarInfinite()
	ui.loadingOverlay = container.NewCenter(container.NewVBox(
		ui.loadingText,
		ui.loadingSpinner,
	))

	contentWithOverlay := container.NewStack(
		ui.mainStack,
		ui.loadingOverlay,
	)

	centerLayout := container.NewBorder(topBar, footer, nil, nil, contentWithOverlay)
	mainSplit := container.NewHSplit(sidebar, centerLayout)
	mainSplit.SetOffset(0.20)

	ui.window.SetContent(mainSplit)
	ui.initialized = true
}

func (ui *App) buildTopBar() fyne.CanvasObject {
	ui.searchEntry = widget.NewEntry()
	ui.searchEntry.SetPlaceHolder("Search packages, tools, libraries...")
	ui.searchEntry.OnChanged = func(q string) {
		ui.searchQuery = q
		if q != "" && ui.activeView == pkgmgr.ViewDiscover {
			ui.activeView = pkgmgr.ViewAll
			ui.updateViewTitle()
		}
		if ui.initialized {
			ui.applyFilter()
		}
	}

	ui.sortSelect = widget.NewSelect([]string{
		pkgmgr.SortRelevance.String(),
		pkgmgr.SortNameAsc.String(),
		pkgmgr.SortNameDesc.String(),
		pkgmgr.SortSizeDesc.String(),
		pkgmgr.SortStatus.String(),
	}, func(s string) {
		switch s {
		case pkgmgr.SortNameAsc.String():
			ui.activeSort = pkgmgr.SortNameAsc
		case pkgmgr.SortNameDesc.String():
			ui.activeSort = pkgmgr.SortNameDesc
		case pkgmgr.SortSizeDesc.String():
			ui.activeSort = pkgmgr.SortSizeDesc
		case pkgmgr.SortStatus.String():
			ui.activeSort = pkgmgr.SortStatus
		default:
			ui.activeSort = pkgmgr.SortRelevance
		}
		if ui.initialized {
			ui.applyFilter()
		}
	})
	ui.sortSelect.Selected = pkgmgr.SortRelevance.String()

	ui.refreshBtn = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		ui.promptRefresh()
	})

	ui.upgradeAllBtn = widget.NewButtonWithIcon("Upgrade All", theme.UploadIcon(), func() {
		ui.promptUpgradeAll()
	})
	ui.upgradeAllBtn.Importance = widget.HighImportance

	rightControls := container.NewHBox(
		ui.sortSelect,
		ui.refreshBtn,
		ui.upgradeAllBtn,
	)

	return container.NewPadded(
		container.NewBorder(nil, nil, ui.viewTitle, rightControls, ui.searchEntry),
	)
}

func (ui *App) buildSidebar() fyne.CanvasObject {
	brandTitle := canvas.NewText("STORE", theme.Color(theme.ColorNameForeground))
	brandTitle.TextStyle = fyne.TextStyle{Bold: true}
	brandTitle.TextSize = 20

	brandBox := container.NewVBox(brandTitle, ui.backendPill)

	// Navigation Items
	navDiscover := widget.NewButtonWithIcon("Discover", theme.HomeIcon(), func() {
		ui.setView(pkgmgr.ViewDiscover)
	})
	navAll := widget.NewButtonWithIcon("All Packages", theme.ListIcon(), func() {
		ui.setView(pkgmgr.ViewAll)
	})
	navInstalled := widget.NewButtonWithIcon("Installed", theme.ConfirmIcon(), func() {
		ui.setView(pkgmgr.ViewInstalled)
	})
	navUpdates := widget.NewButtonWithIcon("Updates", theme.ViewRefreshIcon(), func() {
		ui.setView(pkgmgr.ViewUpdates)
	})

	installedRow := container.NewBorder(nil, nil, nil, ui.installedBadge, navInstalled)
	updatesRow := container.NewBorder(nil, nil, nil, ui.updatesBadge, navUpdates)

	navSection := container.NewVBox(
		navDiscover,
		navAll,
		installedRow,
		updatesRow,
	)

	// Categories Section
	catHeader := widget.NewLabelWithStyle("CATEGORIES", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	catVBox := container.NewVBox(catHeader)

	for _, c := range pkgmgr.AllCategories {
		cat := c // capture
		btn := widget.NewButtonWithIcon(cat.Name, IconForCategory(cat.ID), func() {
			ui.setCategory(cat.ID)
		})
		catVBox.Add(btn)
	}

	sidebarContent := container.NewVBox(
		brandBox,
		widget.NewSeparator(),
		navSection,
		widget.NewSeparator(),
		catVBox,
	)

	ui.sidebarContainer = container.NewVScroll(container.NewPadded(sidebarContent))
	return ui.sidebarContainer
}

func (ui *App) setView(v pkgmgr.View) {
	ui.activeView = v
	ui.activeCategory = pkgmgr.CategoryAll
	ui.updateViewTitle()
	ui.applyFilter()
}

func (ui *App) setCategory(c pkgmgr.Category) {
	ui.activeCategory = c
	if ui.activeView == pkgmgr.ViewDiscover {
		ui.activeView = pkgmgr.ViewAll
	}
	ui.updateViewTitle()
	ui.applyFilter()
}

func (ui *App) updateViewTitle() {
	if ui.viewTitle == nil {
		return
	}
	if ui.activeCategory != pkgmgr.CategoryAll {
		catInfo := pkgmgr.GetCategoryInfo(ui.activeCategory)
		ui.viewTitle.Text = catInfo.Name
	} else {
		ui.viewTitle.Text = ui.activeView.String()
	}
	ui.viewTitle.Refresh()
}

func (ui *App) showPackageDetail(p pkgmgr.Package) {
	if ui.activeView == pkgmgr.ViewDiscover {
		ui.activeView = pkgmgr.ViewAll
		ui.updateViewTitle()
		ui.applyFilter()
	}
	if ui.detailPane != nil {
		ui.detailPane.show(p)
	}
}

func (ui *App) applyFilter() {
	if !ui.initialized {
		return
	}

	if ui.activeView == pkgmgr.ViewDiscover && ui.searchQuery == "" && ui.activeCategory == pkgmgr.CategoryAll {
		if ui.listSplit != nil {
			ui.listSplit.Hide()
		}
		if ui.discoverView != nil {
			ui.discoverView.object().Show()
		}
		if ui.sortSelect != nil {
			ui.sortSelect.Hide()
		}
		ui.refreshDiscover()
		return
	}

	if ui.discoverView != nil {
		ui.discoverView.object().Hide()
	}
	if ui.listSplit != nil {
		ui.listSplit.Show()
	}
	if ui.sortSelect != nil {
		ui.sortSelect.Show()
	}

	opts := pkgmgr.QueryOptions{
		Query:    ui.searchQuery,
		View:     ui.activeView,
		Category: ui.activeCategory,
		Sort:     ui.activeSort,
	}

	pkgs := ui.catalog.QueryWithOptions(opts)

	ui.mu.Lock()
	ui.filteredPkgs = pkgs
	ui.mu.Unlock()

	if ui.pkgList != nil {
		ui.pkgList.Refresh()
	}

	stats := ui.catalog.Stats()
	if ui.statusLabel != nil {
		ui.statusLabel.SetText(countsLine(stats, len(pkgs), ui.searchQuery, ui.activeView))
	}

	ui.updateBadges(stats)
}

func (ui *App) refreshDiscover() {
	if ui.discoverView != nil {
		featured := ui.catalog.Featured()
		counts := ui.catalog.CategoryCounts()
		ui.discoverView.refresh(featured, counts)
	}

	stats := ui.catalog.Stats()
	if ui.statusLabel != nil {
		ui.statusLabel.SetText(countsLine(stats, stats.Total, "", pkgmgr.ViewDiscover))
	}
	ui.updateBadges(stats)
}

func (ui *App) updateBadges(stats pkgmgr.Stats) {
	if ui.installedBadge != nil {
		if stats.Installed > 0 {
			ui.installedBadge.Text = formatInt(stats.Installed)
		} else {
			ui.installedBadge.Text = ""
		}
		ui.installedBadge.Refresh()
	}

	if ui.updatesBadge != nil {
		if stats.Updates > 0 {
			ui.updatesBadge.Text = fmt.Sprintf("%d updates", stats.Updates)
		} else {
			ui.updatesBadge.Text = ""
		}
		ui.updatesBadge.Refresh()
	}

	if ui.upgradeAllBtn != nil {
		if stats.Updates > 0 {
			ui.upgradeAllBtn.SetText(fmt.Sprintf("Upgrade All (%d)", stats.Updates))
			ui.upgradeAllBtn.Show()
		} else {
			ui.upgradeAllBtn.SetText("Upgrade All")
			ui.upgradeAllBtn.Hide()
		}
	}
}

func (ui *App) promptAction(action string, p pkgmgr.Package) {
	title, body := confirmMessage(action, p)
	d := dialog.NewConfirm(title, body, func(confirm bool) {
		if !confirm {
			return
		}
		if privilege.IsRoot() {
			ui.executeAction(action, p, "")
		} else {
			ui.passwordDialog.Show(
				fmt.Sprintf("Enter password to %s %s:", action, p.Name),
				"",
				func(pwd string) {
					ui.executeAction(action, p, pwd)
				},
				nil,
			)
		}
	}, ui.window)
	d.Show()
}

func (ui *App) promptRefresh() {
	title, body := confirmMessage("refresh", pkgmgr.Package{})
	d := dialog.NewConfirm(title, body, func(confirm bool) {
		if !confirm {
			return
		}
		if privilege.IsRoot() {
			ui.executeRefresh("")
		} else {
			ui.passwordDialog.Show(
				"Enter password to refresh package databases:",
				"",
				func(pwd string) {
					ui.executeRefresh(pwd)
				},
				nil,
			)
		}
	}, ui.window)
	d.Show()
}

func (ui *App) promptUpgradeAll() {
	title, body := confirmMessage("update-all", pkgmgr.Package{})
	d := dialog.NewConfirm(title, body, func(confirm bool) {
		if !confirm {
			return
		}
		if privilege.IsRoot() {
			ui.executeUpgradeAll("")
		} else {
			ui.passwordDialog.Show(
				"Enter password to perform system upgrade:",
				"",
				func(pwd string) {
					ui.executeUpgradeAll(pwd)
				},
				nil,
			)
		}
	}, ui.window)
	d.Show()
}

func (ui *App) executeAction(action string, p pkgmgr.Package, password string) {
	ctx, cancel := context.WithCancel(context.Background())
	actionTitle := fmt.Sprintf("%s %s...", strings.Title(action), p.Name)
	ui.actionModal.Start(actionTitle, cancel)

	go func() {
		defer cancel()
		var err error
		b := ui.catalog.Backend()
		progress := func(line string) {
			ui.actionModal.AppendLog(line)
		}

		switch action {
		case "install":
			err = b.Install(ctx, password, p.Name, progress)
		case "reinstall":
			err = b.Reinstall(ctx, password, p.Name, progress)
		case "remove":
			err = b.Remove(ctx, password, p.Name, progress)
		case "update":
			err = b.Upgrade(ctx, password, p.Name, progress)
		}

		ui.actionModal.Finish(err)

		if err != nil && strings.Contains(err.Error(), "authentication failed") {
			// Offer retry password prompt
			ui.passwordDialog.Show(
				fmt.Sprintf("Enter password to %s %s:", action, p.Name),
				"Authentication failed: Incorrect password. Please try again.",
				func(newPwd string) {
					ui.executeAction(action, p, newPwd)
				},
				nil,
			)
			return
		}

		if err == nil {
			// Reload catalog to update package states
			_ = ui.catalog.Reload(context.Background())
			ui.applyFilter()
			if updatedPkg, ok := ui.catalog.Lookup(p.Name); ok {
				ui.detailPane.show(updatedPkg)
			}
		}
	}()
}

func (ui *App) executeRefresh(password string) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.actionModal.Start("Refreshing Package Databases...", cancel)

	go func() {
		defer cancel()
		b := ui.catalog.Backend()
		err := b.Refresh(ctx, password, func(line string) {
			ui.actionModal.AppendLog(line)
		})
		ui.actionModal.Finish(err)

		if err != nil && strings.Contains(err.Error(), "authentication failed") {
			ui.passwordDialog.Show(
				"Enter password to refresh package databases:",
				"Authentication failed: Incorrect password. Please try again.",
				func(newPwd string) {
					ui.executeRefresh(newPwd)
				},
				nil,
			)
			return
		}

		if err == nil {
			_ = ui.catalog.Reload(context.Background())
			ui.applyFilter()
		}
	}()
}

func (ui *App) executeUpgradeAll(password string) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.actionModal.Start("Performing Full System Upgrade...", cancel)

	go func() {
		defer cancel()
		b := ui.catalog.Backend()
		err := b.UpgradeAll(ctx, password, func(line string) {
			ui.actionModal.AppendLog(line)
		})
		ui.actionModal.Finish(err)

		if err != nil && strings.Contains(err.Error(), "authentication failed") {
			ui.passwordDialog.Show(
				"Enter password to perform system upgrade:",
				"Authentication failed: Incorrect password. Please try again.",
				func(newPwd string) {
					ui.executeUpgradeAll(newPwd)
				},
				nil,
			)
			return
		}

		if err == nil {
			_ = ui.catalog.Reload(context.Background())
			ui.applyFilter()
		}
	}()
}

// RunAsync initializes the database loading in the background.
func (ui *App) RunAsync() {
	ui.loadingOverlay.Show()

	go func() {
		start := time.Now()
		err := ui.catalog.Reload(context.Background())
		ui.loadingOverlay.Hide()

		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load package catalog: %w", err), ui.window)
		} else {
			ui.statusLabel.SetText(fmt.Sprintf("Loaded in %v", time.Since(start).Round(time.Millisecond)))
			ui.applyFilter()
		}
	}()
}

// Run starts the application window and runs the event loop.
func (ui *App) Run() {
	ui.RunAsync()
	ui.window.ShowAndRun()
}
