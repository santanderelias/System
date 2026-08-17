package ui

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/privilege"
	"store/internal/syssettings"
)

type settingsView struct {
	window         fyne.Window
	actionModal    *actionModal
	passwordDialog *passwordDialog

	content *container.Scroll

	// Services
	servicesList    *widget.List
	allServices     []syssettings.ServiceUnit
	filteredServices []syssettings.ServiceUnit
	serviceSearch   *widget.Entry

	// Maintenance
	cacheSizeLabel  *widget.Label
	orphanSizeLabel *widget.Label
	journalLabel    *widget.Label
}

func newSettingsView(window fyne.Window, am *actionModal, pd *passwordDialog) *settingsView {
	v := &settingsView{
		window:         window,
		actionModal:    am,
		passwordDialog: pd,
	}

	// 1. Services Section
	v.serviceSearch = widget.NewEntry()
	v.serviceSearch.SetPlaceHolder("Filter services (e.g. docker, sshd, bluetooth)...")
	v.serviceSearch.OnChanged = func(q string) {
		v.filterServices(q)
	}

	refreshServicesBtn := widget.NewButtonWithIcon("Refresh Services", theme.ViewRefreshIcon(), func() {
		v.loadServices()
	})

	v.servicesList = widget.NewList(
		func() int {
			return len(v.filteredServices)
		},
		func() fyne.CanvasObject {
			name := canvas.NewText("", theme.Color(theme.ColorNameForeground))
			name.TextStyle = fyne.TextStyle{Bold: true}
			name.TextSize = theme.TextSize()

			desc := canvas.NewText("", mutedColor())
			desc.TextSize = theme.CaptionTextSize()

			badge := canvas.NewText("", installedColor())
			badge.TextSize = theme.CaptionTextSize()

			startBtn := widget.NewButton("Start", nil)
			startBtn.Importance = widget.HighImportance
			stopBtn := widget.NewButton("Stop", nil)
			stopBtn.Importance = widget.DangerImportance
			restartBtn := widget.NewButton("Restart", nil)

			actions := container.NewHBox(startBtn, stopBtn, restartBtn)

			left := container.NewVBox(container.NewHBox(name, badge), desc)
			return container.NewBorder(nil, nil, nil, actions, left)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			if id < 0 || id >= len(v.filteredServices) {
				return
			}
			s := v.filteredServices[id]
			border := o.(*fyne.Container)
			left := border.Objects[0].(*fyne.Container)
			topRow := left.Objects[0].(*fyne.Container)
			name := topRow.Objects[0].(*canvas.Text)
			badge := topRow.Objects[1].(*canvas.Text)
			desc := left.Objects[1].(*canvas.Text)
			actions := border.Objects[1].(*fyne.Container)
			startBtn := actions.Objects[0].(*widget.Button)
			stopBtn := actions.Objects[1].(*widget.Button)
			restartBtn := actions.Objects[2].(*widget.Button)

			name.Text = s.Name
			desc.Text = s.Description
			if s.IsRunning() {
				badge.Text = "[Active]"
				badge.Color = installedColor()
				startBtn.Hide()
				stopBtn.Show()
			} else {
				badge.Text = "[Inactive]"
				badge.Color = mutedColor()
				startBtn.Show()
				stopBtn.Hide()
			}

			startBtn.OnTapped = func() { v.runServiceAction("start", s.Name) }
			stopBtn.OnTapped = func() { v.runServiceAction("stop", s.Name) }
			restartBtn.OnTapped = func() { v.runServiceAction("restart", s.Name) }

			name.Refresh()
			badge.Refresh()
			desc.Refresh()
		},
	)

	servicesSection := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("System Services & Daemons", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewBorder(nil, nil, nil, refreshServicesBtn, v.serviceSearch),
		),
		nil, nil, nil,
		v.servicesList,
	)
	servicesSection.Resize(fyne.NewSize(600, 350))

	// 2. Maintenance Section
	v.cacheSizeLabel = widget.NewLabel("Cache: calculating...")
	v.orphanSizeLabel = widget.NewLabel("Orphans: scanning...")
	v.journalLabel = widget.NewLabel("Journal: calculating...")

	cleanCacheBtn := widget.NewButtonWithIcon("Clean Package Cache", theme.DeleteIcon(), func() {
		v.runMaintenanceAction("Clean Package Cache", func(ctx context.Context, pwd string, p privilege.ProgressFunc) error {
			return syssettings.CleanCache(ctx, pwd, p)
		})
	})
	cleanCacheBtn.Importance = widget.HighImportance

	removeOrphansBtn := widget.NewButtonWithIcon("Remove Orphan Packages", theme.DeleteIcon(), func() {
		v.runMaintenanceAction("Remove Orphan Packages", func(ctx context.Context, pwd string, p privilege.ProgressFunc) error {
			orphans, _ := syssettings.GetOrphans(ctx)
			return syssettings.RemoveOrphans(ctx, pwd, orphans, p)
		})
	})
	removeOrphansBtn.Importance = widget.HighImportance

	vacuumJournalBtn := widget.NewButtonWithIcon("Vacuum System Logs (<200M)", theme.DeleteIcon(), func() {
		v.runMaintenanceAction("Vacuum System Logs", func(ctx context.Context, pwd string, p privilege.ProgressFunc) error {
			return syssettings.VacuumJournal(ctx, pwd, p)
		})
	})

	maintCardBg := canvas.NewRectangle(cardBgColor())
	maintCardBg.CornerRadius = 10
	maintTitle := widget.NewLabelWithStyle("System Maintenance & Disk Cleanup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	maintCard := container.NewStack(maintCardBg, container.NewPadded(container.NewVBox(
		maintTitle,
		container.NewBorder(nil, nil, v.cacheSizeLabel, cleanCacheBtn),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, v.orphanSizeLabel, removeOrphansBtn),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, v.journalLabel, vacuumJournalBtn),
	)))

	mainLayout := container.NewPadded(container.NewVBox(
		maintCard,
		widget.NewSeparator(),
		servicesSection,
	))

	v.content = container.NewVScroll(mainLayout)
	v.loadServices()
	v.refreshMaintenanceStats()
	return v
}

func (v *settingsView) Object() fyne.CanvasObject {
	return v.content
}

func (v *settingsView) loadServices() {
	go func() {
		units, err := syssettings.ListServices(context.Background())
		if err == nil {
			v.allServices = units
			v.filterServices(v.serviceSearch.Text)
		}
	}()
}

func (v *settingsView) filterServices(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		v.filteredServices = v.allServices
	} else {
		var filtered []syssettings.ServiceUnit
		for _, s := range v.allServices {
			if strings.Contains(strings.ToLower(s.Name), q) || strings.Contains(strings.ToLower(s.Description), q) {
				filtered = append(filtered, s)
			}
		}
		v.filteredServices = filtered
	}
	v.servicesList.Refresh()
}

func (v *settingsView) refreshMaintenanceStats() {
	go func() {
		cacheBytes := syssettings.GetCacheSize()
		v.cacheSizeLabel.SetText(fmt.Sprintf("Pacman Cache: %s", formatBytes(cacheBytes)))

		orphans, _ := syssettings.GetOrphans(context.Background())
		if len(orphans) == 0 {
			v.orphanSizeLabel.SetText("Orphan Packages: 0 found (Clean)")
		} else {
			v.orphanSizeLabel.SetText(fmt.Sprintf("Orphan Packages: %d found (%s)", len(orphans), strings.Join(orphans, ", ")))
		}

		jSize := syssettings.GetJournalSize(context.Background())
		v.journalLabel.SetText(fmt.Sprintf("Journal Logs: %s", jSize))
	}()
}

func (v *settingsView) runServiceAction(action, service string) {
	desc := fmt.Sprintf("Execute '%s' on service %s?", action, service)
	d := dialog.NewConfirm("Service Management", desc, func(confirm bool) {
		if !confirm {
			return
		}
		if privilege.IsRoot() {
			v.executeService(action, service, "")
		} else {
			v.passwordDialog.Show(
				fmt.Sprintf("Enter password to %s %s:", action, service),
				"",
				func(pwd string) {
					v.executeService(action, service, pwd)
				},
				nil,
			)
		}
	}, v.window)
	d.Show()
}

func (v *settingsView) executeService(action, service, password string) {
	ctx, cancel := context.WithCancel(context.Background())
	v.actionModal.Start(fmt.Sprintf("%s service %s...", strings.Title(action), service), cancel)

	go func() {
		defer cancel()
		var err error
		progress := func(l string) { v.actionModal.AppendLog(l) }

		switch action {
		case "start":
			err = syssettings.StartService(ctx, password, service, progress)
		case "stop":
			err = syssettings.StopService(ctx, password, service, progress)
		case "restart":
			err = syssettings.RestartService(ctx, password, service, progress)
		}

		v.actionModal.Finish(err)
		if err == nil {
			v.loadServices()
		}
	}()
}

func (v *settingsView) runMaintenanceAction(title string, fn func(ctx context.Context, pwd string, p privilege.ProgressFunc) error) {
	d := dialog.NewConfirm("Maintenance Operation", fmt.Sprintf("Run '%s'?", title), func(confirm bool) {
		if !confirm {
			return
		}
		if privilege.IsRoot() {
			v.executeMaintenance(title, "", fn)
		} else {
			v.passwordDialog.Show(
				fmt.Sprintf("Enter password for %s:", title),
				"",
				func(pwd string) {
					v.executeMaintenance(title, pwd, fn)
				},
				nil,
			)
		}
	}, v.window)
	d.Show()
}

func (v *settingsView) executeMaintenance(title, password string, fn func(ctx context.Context, pwd string, p privilege.ProgressFunc) error) {
	ctx, cancel := context.WithCancel(context.Background())
	v.actionModal.Start(title, cancel)

	go func() {
		defer cancel()
		progress := func(l string) { v.actionModal.AppendLog(l) }
		err := fn(ctx, password, progress)
		v.actionModal.Finish(err)
		if err == nil {
			v.refreshMaintenanceStats()
		}
	}()
}
