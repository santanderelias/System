package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/catalog"
	"store/internal/privilege"
	"store/internal/sysinfo"
)

// Section identifies a primary subsystem in the suite.
type Section string

const (
	SectionHub      Section = "hub"
	SectionStore    Section = "apps"
	SectionInfo     Section = "info"
	SectionSettings Section = "settings"
)

// Launcher is the main unified Control Center orchestrating all subsystems.
type Launcher struct {
	fyneApp fyne.App
	window  fyne.Window
	catalog *catalog.Catalog

	currentSection Section

	// Subsystem Views
	hubContainer *fyne.Container
	storeApp     *App
	sysInfoView  *sysInfoView
	settingsView *settingsView

	// Root container
	mainSplit    *container.Split
	contentStack *fyne.Container

	// Sidebar Nav buttons
	btnHub      *widget.Button
	btnStore    *widget.Button
	btnInfo     *widget.Button
	btnSettings *widget.Button
}

// NewLauncherApp creates the unified System Administration Suite.
func NewLauncherApp(cat *catalog.Catalog, startSection Section) *Launcher {
	a := app.NewWithID("dev.store.systemsuite")
	a.Settings().SetTheme(&storeTheme{})

	w := a.NewWindow("Store — System Administration Suite")
	w.Resize(fyne.NewSize(1180, 760))
	w.SetMaster()

	if startSection == "" {
		startSection = SectionHub
	}

	l := &Launcher{
		fyneApp:        a,
		window:         w,
		catalog:        cat,
		currentSection: startSection,
	}

	l.initUI()
	return l
}

func (l *Launcher) initUI() {
	// Create subsystem views
	l.storeApp = NewApp(l.catalog)

	actionModal := newActionModal(l.window)
	pwdDialog := newPasswordDialog(l.window)

	l.sysInfoView = newSysInfoView()
	l.settingsView = newSettingsView(l.window, actionModal, pwdDialog)

	// Hub Overview Card View
	l.hubContainer = l.buildHub()

	// Content Stack
	l.contentStack = container.NewStack(
		l.hubContainer,
		l.storeApp.window.Content(),
		l.sysInfoView.Object(),
		l.settingsView.Object(),
	)

	// Global Sidebar
	sidebar := l.buildGlobalSidebar()

	l.mainSplit = container.NewHSplit(sidebar, l.contentStack)
	l.mainSplit.SetOffset(0.18)

	l.window.SetContent(l.mainSplit)
	l.SwitchSection(l.currentSection)
}

func (l *Launcher) buildGlobalSidebar() fyne.CanvasObject {
	brandTitle := canvas.NewText("SYSTEM SUITE", theme.Color(theme.ColorNameForeground))
	brandTitle.TextStyle = fyne.TextStyle{Bold: true}
	brandTitle.TextSize = 16

	var authTag string
	if privilege.IsRoot() {
		authTag = "Root Privileges"
	} else {
		authTag = "User Mode"
	}
	brandSub := canvas.NewText(authTag, brandAccent())
	brandSub.TextSize = theme.CaptionTextSize()

	brandHeader := container.NewVBox(brandTitle, brandSub)

	l.btnHub = widget.NewButtonWithIcon("Hub & Overview", theme.HomeIcon(), func() {
		l.SwitchSection(SectionHub)
	})
	l.btnStore = widget.NewButtonWithIcon("Apps & Store", theme.StorageIcon(), func() {
		l.SwitchSection(SectionStore)
	})
	l.btnInfo = widget.NewButtonWithIcon("System Info", theme.InfoIcon(), func() {
		l.SwitchSection(SectionInfo)
	})
	l.btnSettings = widget.NewButtonWithIcon("Administration", theme.SettingsIcon(), func() {
		l.SwitchSection(SectionSettings)
	})

	navButtons := container.NewVBox(
		l.btnHub,
		l.btnStore,
		l.btnInfo,
		l.btnSettings,
	)

	return container.NewPadded(container.NewVBox(
		brandHeader,
		widget.NewSeparator(),
		navButtons,
	))
}

func (l *Launcher) buildHub() *fyne.Container {
	snap := sysinfo.GetSnapshot()

	// Hero Title
	heroTitle := canvas.NewText("System Control Center", theme.Color(theme.ColorNameForeground))
	heroTitle.TextStyle = fyne.TextStyle{Bold: true}
	heroTitle.TextSize = 24

	heroSub := canvas.NewText(fmt.Sprintf("%s · Kernel %s · Host: %s", snap.OS.PrettyName, snap.OS.Kernel, snap.OS.Hostname), mutedColor())
	heroSub.TextSize = theme.TextSize()

	heroBg := canvas.NewRectangle(rgb(0x13, 0x19, 0x24))
	heroBg.CornerRadius = 12
	heroBox := container.NewStack(heroBg, container.NewPadded(container.NewVBox(heroTitle, heroSub)))

	// Section Hub Cards
	grid := container.NewGridWithColumns(3)

	// Card 1: Apps & Store
	card1Bg := canvas.NewRectangle(cardBgColor())
	card1Bg.CornerRadius = 10
	c1Title := widget.NewLabelWithStyle("Apps & Store", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	c1Desc := widget.NewLabel("Search, install, update and manage packages across your distribution.")
	c1Desc.Wrapping = fyne.TextWrapWord
	c1Btn := widget.NewButtonWithIcon("Open Store", theme.NavigateNextIcon(), func() {
		l.SwitchSection(SectionStore)
	})
	c1Btn.Importance = widget.HighImportance
	card1 := container.NewStack(card1Bg, container.NewPadded(container.NewVBox(c1Title, c1Desc, c1Btn)))

	// Card 2: System Info
	card2Bg := canvas.NewRectangle(cardBgColor())
	card2Bg.CornerRadius = 10
	c2Title := widget.NewLabelWithStyle("System Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	c2Desc := widget.NewLabel("Detailed hardware specs, memory meters, disk space, and OS diagnostics.")
	c2Desc.Wrapping = fyne.TextWrapWord
	c2Btn := widget.NewButtonWithIcon("Open System Info", theme.NavigateNextIcon(), func() {
		l.SwitchSection(SectionInfo)
	})
	card2 := container.NewStack(card2Bg, container.NewPadded(container.NewVBox(c2Title, c2Desc, c2Btn)))

	// Card 3: System Administration
	card3Bg := canvas.NewRectangle(cardBgColor())
	card3Bg.CornerRadius = 10
	c3Title := widget.NewLabelWithStyle("Administration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	c3Desc := widget.NewLabel("Manage systemd services, clean package caches, remove orphans & vacuum logs.")
	c3Desc.Wrapping = fyne.TextWrapWord
	c3Btn := widget.NewButtonWithIcon("Open Settings", theme.NavigateNextIcon(), func() {
		l.SwitchSection(SectionSettings)
	})
	card3 := container.NewStack(card3Bg, container.NewPadded(container.NewVBox(c3Title, c3Desc, c3Btn)))

	grid.Add(card1)
	grid.Add(card2)
	grid.Add(card3)

	hubContent := container.NewVBox(
		heroBox,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Quick Modules", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		grid,
	)

	return container.NewPadded(container.NewVScroll(hubContent))
}

// SwitchSection switches the main visible subsystem.
func (l *Launcher) SwitchSection(s Section) {
	l.currentSection = s

	l.btnHub.Importance = widget.LowImportance
	l.btnStore.Importance = widget.LowImportance
	l.btnInfo.Importance = widget.LowImportance
	l.btnSettings.Importance = widget.LowImportance

	l.hubContainer.Hide()
	l.storeApp.window.Content().Hide()
	l.sysInfoView.Object().Hide()
	l.settingsView.Object().Hide()

	switch s {
	case SectionStore:
		l.btnStore.Importance = widget.HighImportance
		l.storeApp.window.Content().Show()
	case SectionInfo:
		l.btnInfo.Importance = widget.HighImportance
		l.sysInfoView.Refresh()
		l.sysInfoView.Object().Show()
	case SectionSettings:
		l.btnSettings.Importance = widget.HighImportance
		l.settingsView.Object().Show()
	default: // Hub
		l.btnHub.Importance = widget.HighImportance
		l.hubContainer.Show()
	}

	l.btnHub.Refresh()
	l.btnStore.Refresh()
	l.btnInfo.Refresh()
	l.btnSettings.Refresh()
}

// Run starts the launcher and runs the app.
func (l *Launcher) Run() {
	l.storeApp.RunAsync()
	l.window.ShowAndRun()
}
