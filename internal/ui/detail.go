package ui

import (
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/pkgmgr"
)

type detailPane struct {
	empty      *fyne.Container
	filled     *container.Scroll
	stack      *fyne.Container
	iconWidget *widget.Icon
	title      *canvas.Text
	repoPill   *canvas.Text
	catPill    *canvas.Text
	desc       *widget.Label
	installBtn *widget.Button
	updateBtn  *widget.Button
	reinstBtn  *widget.Button
	removeBtn  *widget.Button
	urlBtn     *widget.Button
	fields     *widget.Form
	depsText   *widget.Label
	optText    *widget.Label
	provText   *widget.Label
	confText   *widget.Label
	onAct      func(action string, p pkgmgr.Package)
	current    pkgmgr.Package
	has        bool
}

func newDetailPane(onAct func(action string, p pkgmgr.Package)) *detailPane {
	d := &detailPane{onAct: onAct}

	// Hero Icon
	d.iconWidget = widget.NewIcon(theme.FolderIcon())
	iconBox := canvas.NewRectangle(rgb(0x1e, 0x26, 0x36))
	iconBox.CornerRadius = 14
	iconStack := container.NewStack(iconBox, container.NewPadded(d.iconWidget))
	iconStack.Resize(fyne.NewSize(64, 64))

	// Title & Tags
	d.title = canvas.NewText("Package Details", theme.Color(theme.ColorNameForeground))
	d.title.TextStyle = fyne.TextStyle{Bold: true}
	d.title.TextSize = 22

	d.repoPill = canvas.NewText("", brandAccent())
	d.repoPill.TextSize = theme.CaptionTextSize()

	d.catPill = canvas.NewText("", mutedColor())
	d.catPill.TextSize = theme.CaptionTextSize()

	titleRow := container.NewHBox(d.title, d.repoPill, d.catPill)

	d.desc = widget.NewLabel("")
	d.desc.Wrapping = fyne.TextWrapWord

	// Action buttons
	d.installBtn = widget.NewButtonWithIcon("Install", theme.DownloadIcon(), func() { d.fire("install") })
	d.installBtn.Importance = widget.HighImportance

	d.updateBtn = widget.NewButtonWithIcon("Update", theme.ViewRefreshIcon(), func() { d.fire("update") })
	d.updateBtn.Importance = widget.HighImportance

	d.reinstBtn = widget.NewButtonWithIcon("Reinstall", theme.ViewRefreshIcon(), func() { d.fire("reinstall") })

	d.removeBtn = widget.NewButtonWithIcon("Uninstall", theme.DeleteIcon(), func() { d.fire("remove") })
	d.removeBtn.Importance = widget.DangerImportance

	d.urlBtn = widget.NewButtonWithIcon("Homepage", theme.NavigateNextIcon(), func() {
		if d.current.URL != "" {
			if u, err := url.Parse(d.current.URL); err == nil {
				_ = fyne.CurrentApp().OpenURL(u)
			}
		}
	})

	actions := container.NewHBox(d.installBtn, d.updateBtn, d.reinstBtn, d.removeBtn, d.urlBtn)

	headerCard := container.NewBorder(nil, nil, iconStack, nil,
		container.NewPadded(container.NewVBox(titleRow, d.desc, actions)),
	)

	d.fields = widget.NewForm()

	d.depsText = widget.NewLabel("")
	d.depsText.Wrapping = fyne.TextWrapWord

	d.optText = widget.NewLabel("")
	d.optText.Wrapping = fyne.TextWrapWord

	d.provText = widget.NewLabel("")
	d.provText.Wrapping = fyne.TextWrapWord

	d.confText = widget.NewLabel("")
	d.confText.Wrapping = fyne.TextWrapWord

	depsSection := widget.NewAccordion(
		widget.NewAccordionItem("Dependencies", d.depsText),
		widget.NewAccordionItem("Optional Dependencies", d.optText),
		widget.NewAccordionItem("Provides & Conflicts", container.NewVBox(d.provText, d.confText)),
	)

	bodyLayout := container.NewVBox(
		headerCard,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Specifications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		d.fields,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Package Relations", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		depsSection,
	)

	d.filled = container.NewVScroll(container.NewPadded(bodyLayout))

	hintIcon := widget.NewIcon(theme.FolderOpenIcon())
	hintTitle := widget.NewLabelWithStyle("Select a package to view details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	hintSub := widget.NewLabel("Choose any app from the list or search above to inspect, install or manage it.")
	hintSub.Alignment = fyne.TextAlignCenter
	hintSub.Wrapping = fyne.TextWrapWord

	d.empty = container.NewCenter(container.NewVBox(
		container.NewCenter(hintIcon),
		hintTitle,
		hintSub,
	))

	d.stack = container.NewStack(d.empty, d.filled)
	d.filled.Hide()
	return d
}

func (d *detailPane) object() fyne.CanvasObject { return d.stack }

func (d *detailPane) fire(action string) {
	if d.has && d.onAct != nil {
		d.onAct(action, d.current)
	}
}

func (d *detailPane) clear() {
	d.has = false
	d.current = pkgmgr.Package{}
	d.filled.Hide()
	d.empty.Show()
}

func (d *detailPane) show(p pkgmgr.Package) {
	d.has = true
	d.current = p
	d.empty.Hide()
	d.filled.Show()

	d.iconWidget.SetResource(IconForPackage(p))
	d.title.Text = p.Name
	if p.Repo != "" {
		d.repoPill.Text = "[" + p.Repo + "]"
	} else if p.Foreign {
		d.repoPill.Text = "[local]"
	} else {
		d.repoPill.Text = ""
	}

	catInfo := pkgmgr.GetCategoryInfo(p.Category)
	d.catPill.Text = "· " + catInfo.Name

	d.desc.SetText(p.Description)

	d.installBtn.Hide()
	d.updateBtn.Hide()
	d.reinstBtn.Hide()
	d.removeBtn.Hide()

	switch {
	case p.UpdateAvailable:
		d.updateBtn.Show()
		d.removeBtn.Show()
	case p.Installed:
		d.reinstBtn.Show()
		d.removeBtn.Show()
	default:
		d.installBtn.Show()
	}

	if p.URL != "" {
		d.urlBtn.Show()
	} else {
		d.urlBtn.Hide()
	}

	d.fields.Items = nil
	add := func(k, v string) {
		if v == "" || v == "—" {
			return
		}
		lbl := widget.NewLabel(v)
		lbl.Wrapping = fyne.TextWrapWord
		d.fields.Append(k, lbl)
	}

	if p.Installed && p.InstalledVersion != "" && p.Version != "" && p.InstalledVersion != p.Version {
		add("Installed Version", p.InstalledVersion)
		add("Repository Version", p.Version)
	} else {
		add("Version", p.DisplayVersion())
	}
	if p.UpdateAvailable && p.NewVersion != "" {
		add("Update Target", p.NewVersion)
	}
	add("Repository", p.Source())
	add("Architecture", p.Arch)
	if p.Installed {
		if p.Explicit {
			add("Install Reason", "Explicitly installed by user")
		} else {
			add("Install Reason", "Installed as dependency")
		}
	}
	add("Download Size", formatBytes(p.DownloadSize))
	add("Installed Size", formatBytes(p.InstalledSize))
	add("Licenses", joinOrDash(p.Licenses))
	add("Groups", joinOrDash(p.Groups))
	add("Build Date", formatTime(p.BuildDate))
	add("Install Date", formatTime(p.InstallDate))
	add("Packager", p.Packager)

	// Relations
	d.depsText.SetText(joinOrDash(p.Depends))
	d.optText.SetText(joinOrDash(p.OptDepends))
	d.provText.SetText("Provides: " + joinOrDash(p.Provides))
	d.confText.SetText("Conflicts: " + joinOrDash(p.Conflicts))

	d.title.Refresh()
	d.repoPill.Refresh()
	d.catPill.Refresh()
	d.fields.Refresh()
	d.filled.Refresh()
}

func confirmMessage(action string, p pkgmgr.Package) (title, body string) {
	switch action {
	case "install":
		return "Install " + p.Name + "?",
			"Install " + p.Name + " (" + p.Version + ") and its required dependencies using " + p.Source() + "."
	case "reinstall":
		return "Reinstall " + p.Name + "?",
			"Reinstall " + p.Name + " (" + p.DisplayVersion() + ") from the repositories."
	case "remove":
		return "Remove " + p.Name + "?",
			"Uninstall " + p.Name + " (" + p.DisplayVersion() + ") from your system. Packages depending on it might also be affected."
	case "update":
		dest := p.NewVersion
		if dest == "" {
			dest = p.Version
		}
		return "Update " + p.Name + "?",
			"Upgrade " + p.Name + " from " + p.DisplayVersion() + " to " + dest + "."
	case "update-all":
		return "Perform System Upgrade?",
			"Refresh package databases and upgrade all outdated packages."
	case "refresh":
		return "Refresh Package Databases?",
			"Synchronize and update package repositories to fetch the latest index."
	default:
		return strings.Title(action), "Execute this action on " + p.Name + "?"
	}
}
