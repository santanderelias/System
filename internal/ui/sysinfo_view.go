package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"store/internal/sysinfo"
)

type sysInfoView struct {
	content *container.Scroll

	// OS Card
	osTitle    *canvas.Text
	osDetails  *widget.Form

	// Resource Meters
	cpuModel   *widget.Label
	cpuDetails *widget.Label

	memBar     *widget.ProgressBar
	memLabel   *widget.Label

	swapBar    *widget.ProgressBar
	swapLabel  *widget.Label

	// Disks
	disksBox   *fyne.Container
}

func newSysInfoView() *sysInfoView {
	v := &sysInfoView{}

	v.osTitle = canvas.NewText("System Overview", theme.Color(theme.ColorNameForeground))
	v.osTitle.TextStyle = fyne.TextStyle{Bold: true}
	v.osTitle.TextSize = 20

	v.osDetails = widget.NewForm()

	v.cpuModel = widget.NewLabel("")
	v.cpuModel.TextStyle = fyne.TextStyle{Bold: true}
	v.cpuDetails = widget.NewLabel("")

	v.memBar = widget.NewProgressBar()
	v.memLabel = widget.NewLabel("")

	v.swapBar = widget.NewProgressBar()
	v.swapLabel = widget.NewLabel("")

	v.disksBox = container.NewVBox()

	refreshBtn := widget.NewButtonWithIcon("Refresh System Info", theme.ViewRefreshIcon(), func() {
		v.Refresh()
	})
	refreshBtn.Importance = widget.HighImportance

	copyBtn := widget.NewButtonWithIcon("Copy System Summary", theme.ContentCopyIcon(), func() {
		snap := sysinfo.GetSnapshot()
		summary := fmt.Sprintf("OS: %s (%s)\nKernel: %s\nHost: %s\nUptime: %s\nCPU: %s (%d cores)\nMemory: %s / %s (%.1f%%)\n",
			snap.OS.PrettyName, snap.OS.Arch, snap.OS.Kernel, snap.OS.Hostname, snap.OS.UptimeFormatted,
			snap.CPU.ModelName, snap.CPU.Cores,
			formatBytes(int64(snap.Memory.UsedBytes)), formatBytes(int64(snap.Memory.TotalBytes)), snap.Memory.UsagePct,
		)
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(summary)
	})

	topActions := container.NewHBox(refreshBtn, copyBtn)

	osCardBg := canvas.NewRectangle(cardBgColor())
	osCardBg.CornerRadius = 10
	osCard := container.NewStack(osCardBg, container.NewPadded(container.NewVBox(
		v.osTitle,
		v.osDetails,
	)))

	resCardBg := canvas.NewRectangle(cardBgColor())
	resCardBg.CornerRadius = 10
	resTitle := widget.NewLabelWithStyle("Processor & Memory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	resCard := container.NewStack(resCardBg, container.NewPadded(container.NewVBox(
		resTitle,
		v.cpuModel,
		v.cpuDetails,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Memory (RAM)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.memBar,
		v.memLabel,
		widget.NewLabelWithStyle("Swap Space", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.swapBar,
		v.swapLabel,
	)))

	diskCardBg := canvas.NewRectangle(cardBgColor())
	diskCardBg.CornerRadius = 10
	diskTitle := widget.NewLabelWithStyle("Storage & Partitions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	diskCard := container.NewStack(diskCardBg, container.NewPadded(container.NewVBox(
		diskTitle,
		v.disksBox,
	)))

	mainLayout := container.NewPadded(container.NewVBox(
		topActions,
		widget.NewSeparator(),
		osCard,
		widget.NewSeparator(),
		resCard,
		widget.NewSeparator(),
		diskCard,
	))

	v.content = container.NewVScroll(mainLayout)
	v.Refresh()
	return v
}

func (v *sysInfoView) Object() fyne.CanvasObject {
	return v.content
}

func (v *sysInfoView) Refresh() {
	snap := sysinfo.GetSnapshot()

	// OS Info
	v.osTitle.Text = snap.OS.PrettyName
	v.osDetails.Items = nil
	v.osDetails.Append("Hostname", widget.NewLabel(snap.OS.Hostname))
	v.osDetails.Append("Kernel", widget.NewLabel(snap.OS.Kernel))
	v.osDetails.Append("Architecture", widget.NewLabel(snap.OS.Arch))
	v.osDetails.Append("Desktop Environment", widget.NewLabel(snap.OS.DesktopEnv))
	v.osDetails.Append("Session Type", widget.NewLabel(snap.OS.WindowManager))
	v.osDetails.Append("System Uptime", widget.NewLabel(snap.OS.UptimeFormatted))
	v.osDetails.Refresh()

	// CPU Info
	v.cpuModel.SetText(snap.CPU.ModelName)
	freqStr := snap.CPU.Frequency
	if freqStr != "" {
		freqStr = " @ " + freqStr
	}
	v.cpuDetails.SetText(fmt.Sprintf("%d physical cores, %d threads%s", snap.CPU.Cores, snap.CPU.Threads, freqStr))

	// Memory Info
	memUsedStr := formatBytes(int64(snap.Memory.UsedBytes))
	memTotalStr := formatBytes(int64(snap.Memory.TotalBytes))
	memAvailStr := formatBytes(int64(snap.Memory.AvailableBytes))
	v.memBar.SetValue(snap.Memory.UsagePct / 100.0)
	v.memLabel.SetText(fmt.Sprintf("Used: %s / %s (%.1f%%)  ·  Available: %s", memUsedStr, memTotalStr, snap.Memory.UsagePct, memAvailStr))

	// Swap Info
	if snap.Memory.SwapTotalBytes > 0 {
		swapUsedStr := formatBytes(int64(snap.Memory.SwapUsedBytes))
		swapTotalStr := formatBytes(int64(snap.Memory.SwapTotalBytes))
		v.swapBar.SetValue(snap.Memory.SwapUsagePct / 100.0)
		v.swapLabel.SetText(fmt.Sprintf("Used: %s / %s (%.1f%%)", swapUsedStr, swapTotalStr, snap.Memory.SwapUsagePct))
		v.swapBar.Show()
		v.swapLabel.Show()
	} else {
		v.swapBar.Hide()
		v.swapLabel.SetText("No Swap partition configured.")
	}

	// Disks Info
	v.disksBox.Objects = nil
	for _, d := range snap.Disks {
		mount := d.MountPoint
		dev := d.Device
		fstype := d.FSType
		usedStr := formatBytes(int64(d.UsedBytes))
		totalStr := formatBytes(int64(d.TotalBytes))
		freeStr := formatBytes(int64(d.FreeBytes))

		dBar := widget.NewProgressBar()
		dBar.SetValue(d.UsagePct / 100.0)

		dTitle := widget.NewLabelWithStyle(fmt.Sprintf("%s (%s - %s)", mount, dev, fstype), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		dSub := widget.NewLabel(fmt.Sprintf("Used: %s / %s (%.1f%%)  ·  Free: %s", usedStr, totalStr, d.UsagePct, freeStr))

		v.disksBox.Add(container.NewVBox(
			dTitle,
			dBar,
			dSub,
			widget.NewSeparator(),
		))
	}
	v.disksBox.Refresh()
}
