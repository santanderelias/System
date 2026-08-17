package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/pkgmgr"
)

type packageRow struct {
	widget.BaseWidget
	pkg      pkgmgr.Package
	onAct    func(action string, p pkgmgr.Package)
	onSelect func(p pkgmgr.Package)
}

func newPackageRow(onAct func(action string, p pkgmgr.Package), onSelect func(p pkgmgr.Package)) *packageRow {
	r := &packageRow{
		onAct:    onAct,
		onSelect: onSelect,
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *packageRow) set(p pkgmgr.Package) {
	r.pkg = p
	r.Refresh()
}

func (r *packageRow) CreateRenderer() fyne.WidgetRenderer {
	// Avatar / Vector Icon
	iconBox := canvas.NewRectangle(rgb(0x1e, 0x24, 0x30))
	iconBox.CornerRadius = 8
	iconWidget := widget.NewIcon(theme.FolderIcon())
	iconStack := container.NewStack(iconBox, container.NewPadded(iconWidget))
	iconStack.Resize(fyne.NewSize(42, 42))

	// Package Name
	name := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	name.TextStyle = fyne.TextStyle{Bold: true}
	name.TextSize = theme.TextSize() + 1

	// Repo / Source Tag
	repoTag := canvas.NewText("", brandAccent())
	repoTag.TextSize = theme.CaptionTextSize() - 1

	// Description
	desc := canvas.NewText("", mutedColor())
	desc.TextSize = theme.CaptionTextSize()

	// Version & Size
	verSize := canvas.NewText("", mutedColor())
	verSize.TextSize = theme.CaptionTextSize() - 1

	// Status badge text
	badgeText := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	badgeText.TextSize = theme.CaptionTextSize()
	badgeBox := canvas.NewRectangle(rgb(0x1a, 0x22, 0x30))
	badgeBox.CornerRadius = 6
	badgeStack := container.NewStack(badgeBox, container.NewPadded(badgeText))

	topLine := container.NewHBox(name, repoTag)
	textColumn := container.NewVBox(topLine, desc, verSize)

	leftSection := container.NewBorder(nil, nil, iconStack, nil, container.NewPadded(textColumn))
	fullRow := container.NewBorder(nil, nil, nil, container.NewCenter(badgeStack), leftSection)

	cardBackground := canvas.NewRectangle(cardBgColor())
	cardBackground.CornerRadius = 8

	mainContent := container.NewStack(cardBackground, container.NewPadded(fullRow))

	return &packageRowRenderer{
		row:            r,
		name:           name,
		repoTag:        repoTag,
		desc:           desc,
		verSize:        verSize,
		badgeText:      badgeText,
		badgeBox:       badgeBox,
		iconWidget:     iconWidget,
		cardBackground: cardBackground,
		content:        mainContent,
	}
}

type packageRowRenderer struct {
	row            *packageRow
	name           *canvas.Text
	repoTag        *canvas.Text
	desc           *canvas.Text
	verSize        *canvas.Text
	badgeText      *canvas.Text
	badgeBox       *canvas.Rectangle
	iconWidget     *widget.Icon
	cardBackground *canvas.Rectangle
	content        *fyne.Container
}

func (r *packageRowRenderer) Destroy() {}

func (r *packageRowRenderer) Layout(size fyne.Size) {
	r.content.Resize(size)
}

func (r *packageRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(320, 68)
}

func (r *packageRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content}
}

func (r *packageRowRenderer) Refresh() {
	p := r.row.pkg
	r.name.Text = p.Name
	r.name.Color = theme.Color(theme.ColorNameForeground)

	if p.Repo != "" {
		r.repoTag.Text = "[" + p.Repo + "]"
	} else if p.Foreign {
		r.repoTag.Text = "[local]"
	} else {
		r.repoTag.Text = ""
	}

	// Truncate long descriptions
	d := p.Description
	if len(d) > 80 {
		d = d[:77] + "..."
	}
	r.desc.Text = d

	// Version & Size
	sz := p.InstalledSize
	if sz == 0 {
		sz = p.DownloadSize
	}
	szStr := formatBytes(sz)

	if p.UpdateAvailable && p.NewVersion != "" {
		r.verSize.Text = p.DisplayVersion() + " -> " + p.NewVersion + " · " + szStr
	} else {
		r.verSize.Text = "v" + p.DisplayVersion() + " · " + szStr
	}

	// Status badge
	switch {
	case p.UpdateAvailable:
		r.badgeText.Text = "Update"
		r.badgeText.Color = rgb(0x0e, 0x11, 0x17)
		r.badgeBox.FillColor = updateColor()
	case p.Installed:
		r.badgeText.Text = "Installed"
		r.badgeText.Color = rgb(0x0e, 0x11, 0x17)
		r.badgeBox.FillColor = installedColor()
	case p.Foreign:
		r.badgeText.Text = "Local"
		r.badgeText.Color = rgb(0x0e, 0x11, 0x17)
		r.badgeBox.FillColor = foreignColor()
	default:
		r.badgeText.Text = "Get"
		r.badgeText.Color = theme.Color(theme.ColorNameForeground)
		r.badgeBox.FillColor = rgb(0x22, 0x2b, 0x3a)
	}

	// Vector icon
	r.iconWidget.SetResource(IconForPackage(p))

	r.name.Refresh()
	r.repoTag.Refresh()
	r.desc.Refresh()
	r.verSize.Refresh()
	r.badgeText.Refresh()
	r.badgeBox.Refresh()
	r.content.Refresh()
}
