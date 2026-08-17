package ui

import (
	"context"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type actionModal struct {
	parent   fyne.Window
	dialog   dialog.Dialog
	title    *widget.Label
	status   *widget.Label
	progress *widget.ProgressBarInfinite
	logArea  *widget.Entry
	logScroll *container.Scroll
	closeBtn *widget.Button
	cancelBtn *widget.Button
	cancelFn context.CancelFunc

	mu   sync.Mutex
	logs []string
}

func newActionModal(parent fyne.Window) *actionModal {
	m := &actionModal{parent: parent}

	m.title = widget.NewLabelWithStyle("Running Action", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	m.status = widget.NewLabel("Executing package manager command...")
	m.progress = widget.NewProgressBarInfinite()

	m.logArea = widget.NewMultiLineEntry()
	m.logArea.TextStyle = fyne.TextStyle{Monospace: true}
	m.logArea.Wrapping = fyne.TextWrapWord
	m.logArea.Disable() // Read-only

	m.logScroll = container.NewVScroll(m.logArea)

	m.cancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		m.mu.Lock()
		if m.cancelFn != nil {
			m.cancelFn()
			m.status.SetText("Cancelling...")
		}
		m.mu.Unlock()
	})

	m.closeBtn = widget.NewButtonWithIcon("Close", theme.ConfirmIcon(), func() {
		if m.dialog != nil {
			m.dialog.Hide()
		}
	})
	m.closeBtn.Importance = widget.HighImportance

	btnRow := container.NewHBox(m.cancelBtn, m.closeBtn)

	content := container.NewBorder(
		container.NewVBox(m.title, m.status, m.progress),
		btnRow,
		nil,
		nil,
		m.logScroll,
	)

	m.dialog = dialog.NewCustomWithoutButtons("Package Operation", content, parent)
	m.dialog.Resize(fyne.NewSize(640, 440))
	return m
}

func (m *actionModal) Start(title string, cancelFn context.CancelFunc) {
	m.mu.Lock()
	m.cancelFn = cancelFn
	m.logs = nil
	m.mu.Unlock()

	m.title.SetText(title)
	m.status.SetText("Working...")
	m.progress.Start()
	m.progress.Show()
	m.logArea.SetText("")
	m.cancelBtn.Enable()
	m.cancelBtn.Show()
	m.closeBtn.Hide()

	m.dialog.Show()
}

func (m *actionModal) AppendLog(line string) {
	m.mu.Lock()
	m.logs = append(m.logs, line)
	allText := strings.Join(m.logs, "\n")
	m.mu.Unlock()

	// Update UI on main thread
	m.logArea.SetText(allText)
	m.logScroll.ScrollToBottom()
}

func (m *actionModal) Finish(err error) {
	m.progress.Stop()
	m.progress.Hide()
	m.cancelBtn.Hide()
	m.closeBtn.Show()

	if err != nil {
		m.status.SetText("Failed: " + err.Error())
		m.AppendLog("\n❌ Error: " + err.Error())
	} else {
		m.status.SetText("Operation completed successfully!")
		m.AppendLog("\n✅ Complete.")
	}
}
