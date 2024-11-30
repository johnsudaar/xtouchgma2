package gui

import (
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
)

type SettingsTab struct {
	debugEnabled         *widget.Check
	logToFile            *widget.Check
	eventLoopRefreshRate *widget.Entry
	inputInhibitTime     *widget.Entry
	xtouchMinRefreshRate *widget.Entry
	xtouchMaxRefreshRate *widget.Entry

	gui *GUI
}

func NewSettingsTag(gui *GUI) *SettingsTab {
	tag := &SettingsTab{
		debugEnabled:         widget.NewCheck("Enable Debug Logs", nil),
		logToFile:            widget.NewCheck("Log to file", nil),
		eventLoopRefreshRate: widget.NewEntry(),
		inputInhibitTime:     widget.NewEntry(),
		xtouchMinRefreshRate: widget.NewEntry(),
		xtouchMaxRefreshRate: widget.NewEntry(),
		gui:                  gui,
	}

	return tag
}

func (g *SettingsTab) getTabItem() *container.TabItem {
	return container.NewTabItem(
		"Settings",
		container.NewVBox(
			widget.NewForm(
				widget.NewFormItem("Enable Debug Logs", g.debugEnabled),
				widget.NewFormItem("Log to file", g.logToFile),
				widget.NewFormItem("Event Loop Refresh Rate", g.eventLoopRefreshRate),
				widget.NewFormItem("Input Inhibit Time", g.inputInhibitTime),
				widget.NewFormItem("X-Touch Min Refresh Rate", g.xtouchMinRefreshRate),
				widget.NewFormItem("X-Touch Max Refresh Rate", g.xtouchMaxRefreshRate),
			),
		),
	)
}
