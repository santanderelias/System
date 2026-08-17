package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/pkgmgr"
)

type discoverView struct {
	content    *container.Scroll
	onSelect   func(p pkgmgr.Package)
	onCategory func(c pkgmgr.Category)
	onAct      func(action string, p pkgmgr.Package)
}

func newDiscoverView(
	onSelect func(p pkgmgr.Package),
	onCategory func(c pkgmgr.Category),
	onAct func(action string, p pkgmgr.Package),
) *discoverView {
	dv := &discoverView{
		onSelect:   onSelect,
		onCategory: onCategory,
		onAct:      onAct,
	}
	dv.content = container.NewVScroll(container.NewVBox())
	return dv
}

func (dv *discoverView) object() fyne.CanvasObject {
	return dv.content
}

func (dv *discoverView) refresh(featured []pkgmgr.Package, counts map[pkgmgr.Category]int) {
	// Hero Header
	heroTitle := canvas.NewText("Discover Software", theme.Color(theme.ColorNameForeground))
	heroTitle.TextStyle = fyne.TextStyle{Bold: true}
	heroTitle.TextSize = 24

	heroSub := canvas.NewText("Find, install and manage applications, developer tools, and system libraries.", mutedColor())
	heroSub.TextSize = theme.TextSize()

	heroBg := canvas.NewRectangle(rgb(0x13, 0x19, 0x24))
	heroBg.CornerRadius = 12

	heroBox := container.NewStack(
		heroBg,
		container.NewPadded(container.NewVBox(heroTitle, heroSub)),
	)

	// Quick Categories Grid
	catTitle := widget.NewLabelWithStyle("Browse by Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	catGrid := container.NewGridWithColumns(3)

	for _, c := range pkgmgr.AllCategories {
		cat := c // capture
		count := counts[cat.ID]
		btn := widget.NewButtonWithIcon(
			cat.Name,
			IconForCategory(cat.ID),
			func() {
				if dv.onCategory != nil {
					dv.onCategory(cat.ID)
				}
			},
		)
		if count > 0 {
			btn.SetText(cat.Name + " (" + formatInt(count) + ")")
		}
		catGrid.Add(btn)
	}

	// Featured Spotlight Section
	spotlightTitle := widget.NewLabelWithStyle("Featured & Popular Applications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	spotlightGrid := container.NewGridWithColumns(2)

	for _, p := range featured {
		pkg := p // capture
		card := dv.makeFeaturedCard(pkg)
		spotlightGrid.Add(card)
	}

	mainLayout := container.NewPadded(container.NewVBox(
		heroBox,
		widget.NewSeparator(),
		catTitle,
		catGrid,
		widget.NewSeparator(),
		spotlightTitle,
		spotlightGrid,
	))

	dv.content.Content = mainLayout
	dv.content.Refresh()
}

func (dv *discoverView) makeFeaturedCard(p pkgmgr.Package) fyne.CanvasObject {
	// Vector Icon Box
	iconBox := canvas.NewRectangle(rgb(0x1e, 0x26, 0x36))
	iconBox.CornerRadius = 10
	iconWidget := widget.NewIcon(IconForPackage(p))
	iconStack := container.NewStack(iconBox, container.NewPadded(iconWidget))
	iconStack.Resize(fyne.NewSize(48, 48))

	// Name & Category
	name := canvas.NewText(p.Name, theme.Color(theme.ColorNameForeground))
	name.TextStyle = fyne.TextStyle{Bold: true}
	name.TextSize = theme.TextSize() + 1

	catInfo := pkgmgr.GetCategoryInfo(p.Category)
	catLabel := canvas.NewText(catInfo.Name, mutedColor())
	catLabel.TextSize = theme.CaptionTextSize()

	descText := p.Description
	if len(descText) > 60 {
		descText = descText[:57] + "..."
	}
	desc := canvas.NewText(descText, mutedColor())
	desc.TextSize = theme.CaptionTextSize()

	// Action button / status
	var actionBtn *widget.Button
	if p.Installed {
		if p.UpdateAvailable {
			actionBtn = widget.NewButtonWithIcon("Update", theme.ViewRefreshIcon(), func() {
				if dv.onAct != nil {
					dv.onAct("update", p)
				}
			})
			actionBtn.Importance = widget.HighImportance
		} else {
			actionBtn = widget.NewButton("Installed", func() {
				if dv.onSelect != nil {
					dv.onSelect(p)
				}
			})
		}
	} else {
		actionBtn = widget.NewButtonWithIcon("Install", theme.DownloadIcon(), func() {
			if dv.onAct != nil {
				dv.onAct("install", p)
			}
		})
		actionBtn.Importance = widget.HighImportance
	}

	detailsBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		if dv.onSelect != nil {
			dv.onSelect(p)
		}
	})

	btnRow := container.NewHBox(actionBtn, detailsBtn)

	infoColumn := container.NewVBox(name, catLabel, desc)
	cardLayout := container.NewBorder(nil, nil, iconStack, btnRow, container.NewPadded(infoColumn))

	cardBg := canvas.NewRectangle(cardBgColor())
	cardBg.CornerRadius = 10

	return container.NewStack(cardBg, container.NewPadded(cardLayout))
}
