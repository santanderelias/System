package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// storeTheme is a dark modern store theme. It gives the application a sleek,
// polished look on all Linux desktop environments (GNOME, KDE Plasma, XFCE, Sway, Hyprland).
type storeTheme struct{}

func (storeTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return rgb(0x0e, 0x11, 0x17)
	case theme.ColorNameForeground:
		return rgb(0xf0, 0xf3, 0xf6)
	case theme.ColorNameDisabled:
		return rgb(0x64, 0x74, 0x8b)
	case theme.ColorNamePlaceHolder:
		return rgb(0x94, 0xa3, 0xb8)
	case theme.ColorNamePrimary:
		return rgb(0x10, 0xb9, 0x81) // Emerald / Mint accent
	case theme.ColorNameForegroundOnPrimary:
		return rgb(0x05, 0x1f, 0x14)
	case theme.ColorNameButton:
		return rgb(0x1e, 0x24, 0x30)
	case theme.ColorNameDisabledButton:
		return rgb(0x15, 0x19, 0x22)
	case theme.ColorNameHover:
		return rgb(0x28, 0x31, 0x41)
	case theme.ColorNamePressed:
		return rgb(0x19, 0x1f, 0x2a)
	case theme.ColorNameFocus:
		return rgb(0x10, 0xb9, 0x81)
	case theme.ColorNameSelection:
		return rgb(0x17, 0x3d, 0x32)
	case theme.ColorNameInputBackground:
		return rgb(0x16, 0x1b, 0x24)
	case theme.ColorNameInputBorder:
		return rgb(0x2a, 0x34, 0x45)
	case theme.ColorNameScrollBar:
		return rgb(0x3e, 0x4c, 0x62)
	case theme.ColorNameScrollBarBackground:
		return rgb(0x0e, 0x11, 0x17)
	case theme.ColorNameSeparator:
		return rgb(0x23, 0x2b, 0x3a)
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0x99}
	case theme.ColorNameOverlayBackground:
		return rgb(0x14, 0x19, 0x22)
	case theme.ColorNameMenuBackground:
		return rgb(0x14, 0x19, 0x22)
	case theme.ColorNameHeaderBackground:
		return rgb(0x11, 0x15, 0x1d)
	case theme.ColorNameHyperlink:
		return rgb(0x38, 0xbd, 0xf8)
	case theme.ColorNameSuccess:
		return rgb(0x34, 0xd3, 0x99)
	case theme.ColorNameWarning:
		return rgb(0xfb, 0xbf, 0x24)
	case theme.ColorNameError:
		return rgb(0xf8, 0x71, 0x71)
	default:
		return theme.DefaultTheme().Color(n, theme.VariantDark)
	}
}

func (storeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (storeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (storeTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameScrollBar:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func rgb(r, g, b uint8) color.Color {
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}
}

func mutedColor() color.Color      { return rgb(0x94, 0xa3, 0xb8) }
func installedColor() color.Color  { return rgb(0x34, 0xd3, 0x99) }
func updateColor() color.Color     { return rgb(0xfb, 0xbf, 0x24) }
func foreignColor() color.Color    { return rgb(0x81, 0x8c, 0xf8) }
func availableColor() color.Color  { return rgb(0x64, 0x74, 0x8b) }
func cardBgColor() color.Color     { return rgb(0x16, 0x1b, 0x24) }
func cardHoverBgColor() color.Color{ return rgb(0x20, 0x27, 0x34) }
func brandAccent() color.Color     { return rgb(0x10, 0xb9, 0x81) }
