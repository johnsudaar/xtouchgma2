package gui

import (
	"fmt"
	"net"
	"strconv"

	"fyne.io/fyne/widget"
	"github.com/johnsudaar/xtouchgma2/link"
)

type ConfigurationTab struct {
	gMAIP        *widget.Entry
	gMAUser      *widget.Entry
	gMAPassword  *widget.Entry
	sACNUniverse *widget.Entry
	formErrors   *widget.Label
	status       *widget.Label
	logs         *widget.Entry
	start        *widget.Button
	stop         *widget.Button

	gui *GUI
}

func NewConfigurationTab(g *GUI) *ConfigurationTab {
	tab := &ConfigurationTab{
		gMAIP:        widget.NewEntry(),
		gMAUser:      widget.NewEntry(),
		gMAPassword:  widget.NewEntry(),
		sACNUniverse: widget.NewEntry(),
		formErrors:   widget.NewLabel(""),
		status:       widget.NewLabel("Status: stopped"),
		logs:         widget.NewMultiLineEntry(),
		gui:          g,
	}

	tab.start = widget.NewButton("Start", tab.onStart)
	tab.stop = widget.NewButton("Stop", tab.onStop)

	return tab
}

func (g *ConfigurationTab) getTabItem() *widget.TabItem {
	return widget.NewTabItem(
		"Configuration",
		widget.NewVBox(
			widget.NewForm(
				g.formItemFor("GrandMA IP: ", g.gMAIP),
				g.formItemFor("GrandMA User: ", g.gMAUser),
				g.formItemFor("GrandMA Password: ", g.gMAPassword),
				g.formItemFor("sACN Universe: ", g.sACNUniverse),
			),
			g.formErrors,
			g.start,
			g.stop,
			g.status,
			g.logs,
		),
	)
}

func (g *ConfigurationTab) onWindowsInit() {
	g.gMAIP.SetPlaceHolder("192.168.1.21")
	g.gMAUser.SetPlaceHolder("user")
	g.gMAPassword.SetPlaceHolder("password")
	g.sACNUniverse.SetPlaceHolder("10")
	g.logs.SetReadOnly(true)
	g.stop.Disable()
}

func (g *ConfigurationTab) formItemFor(label string, entry *widget.Entry) *widget.FormItem {
	return &widget.FormItem{
		Text:   label,
		Widget: entry,
	}
}

func (g *ConfigurationTab) onStart() {
	g.resetFormErrors()
	valid := g.validInputs()
	g.formErrors.Refresh()
	if !valid {
		return
	}

	g.gui.startRequested()
}

func (g *ConfigurationTab) onStop() {
	g.gui.stopRequested()
}

func (g *ConfigurationTab) resetFormErrors() {
	g.formErrors.Text = ""
}

func (g *ConfigurationTab) disableInputs() {
	g.gMAIP.Disable()
	g.gMAUser.Disable()
	g.gMAPassword.Disable()
	g.sACNUniverse.Disable()
}

func (g *ConfigurationTab) enableInputs() {
	g.gMAIP.Enable()
	g.gMAUser.Enable()
	g.gMAPassword.Enable()
	g.sACNUniverse.Enable()
}

func (g *ConfigurationTab) disableStart() {
	g.start.Disable()
}

func (g *ConfigurationTab) disableStop() {
	g.stop.Disable()
}

func (g *ConfigurationTab) enableStart() {
	g.start.Enable()
}

func (g *ConfigurationTab) enableStop() {
	g.stop.Enable()
}

func (g *ConfigurationTab) linkParams() link.NewLinkParams {
	u, _ := strconv.Atoi(g.sACNUniverse.Text)

	return link.NewLinkParams{
		GMAHost:      g.gMAIP.Text,
		GMAUser:      g.gMAUser.Text,
		GMAPassword:  g.gMAPassword.Text,
		SACNUniverse: uint16(u),
	}
}

func (g *ConfigurationTab) validInputs() bool {
	if g.gMAIP.Text == "" {
		g.formErrors.Text = "GrandMA IP can't be empty"
		return false
	}
	ip := net.ParseIP(g.gMAIP.Text)
	if ip == nil {
		g.formErrors.Text = fmt.Sprintf("GrandMA IP: %s is not a valid IP", g.gMAIP.Text)
		return false
	}

	if g.gMAUser.Text == "" {
		g.formErrors.Text = "GrandMA User can't be empty"
		return false
	}

	if g.sACNUniverse.Text == "" {
		g.formErrors.Text = "sACN Universe can't be empty"
		return false
	}

	_, err := strconv.Atoi(g.sACNUniverse.Text)
	if err != nil {
		g.formErrors.Text = fmt.Sprintf("sACNUniverse: %s is not a number", g.sACNUniverse.Text)
		return false
	}

	return true
}

func (g *ConfigurationTab) SetStatus(status string) {
	g.status.Text = "Status : " + status
	g.status.Refresh()
}
