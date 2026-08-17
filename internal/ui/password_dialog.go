package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type passwordDialog struct {
	parent   fyne.Window
	dialog   dialog.Dialog
	desc     *widget.Label
	entry    *widget.Entry
	errLabel *canvas.Text
	onSubmit func(password string)
	onCancel func()
}

func newPasswordDialog(parent fyne.Window) *passwordDialog {
	pd := &passwordDialog{parent: parent}

	title := widget.NewLabelWithStyle("Authentication Required", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	pd.desc = widget.NewLabel("Administrator privileges are required to perform this action.")
	pd.desc.Wrapping = fyne.TextWrapWord

	pd.entry = widget.NewPasswordEntry()
	pd.entry.SetPlaceHolder("Enter password for sudo")

	pd.errLabel = canvas.NewText("", rgb(0xf8, 0x71, 0x71))
	pd.errLabel.TextSize = theme.CaptionTextSize()

	pd.entry.OnSubmitted = func(_ string) {
		pd.submit()
	}

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		pd.cancel()
	})

	authBtn := widget.NewButtonWithIcon("Authenticate", theme.ConfirmIcon(), func() {
		pd.submit()
	})
	authBtn.Importance = widget.HighImportance

	btnRow := container.NewHBox(cancelBtn, authBtn)

	form := container.NewVBox(
		title,
		pd.desc,
		pd.entry,
		pd.errLabel,
		btnRow,
	)

	content := container.NewPadded(form)
	pd.dialog = dialog.NewCustomWithoutButtons("Authentication", content, parent)
	pd.dialog.Resize(fyne.NewSize(420, 220))
	return pd
}

func (pd *passwordDialog) Show(description string, errorMsg string, onSubmit func(password string), onCancel func()) {
	pd.onSubmit = onSubmit
	pd.onCancel = onCancel

	if description != "" {
		pd.desc.SetText(description)
	} else {
		pd.desc.SetText("Administrator privileges are required to perform this action.")
	}

	pd.entry.SetText("")
	pd.errLabel.Text = errorMsg
	pd.errLabel.Refresh()

	pd.dialog.Show()
	pd.parent.Canvas().Focus(pd.entry)
}

func (pd *passwordDialog) Hide() {
	if pd.dialog != nil {
		pd.dialog.Hide()
	}
}

func (pd *passwordDialog) submit() {
	pwd := pd.entry.Text
	pd.Hide()
	if pd.onSubmit != nil {
		pd.onSubmit(pwd)
	}
}

func (pd *passwordDialog) cancel() {
	pd.Hide()
	if pd.onCancel != nil {
		pd.onCancel()
	}
}
