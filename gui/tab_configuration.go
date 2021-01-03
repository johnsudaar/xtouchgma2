package gui

import (
	"fmt"
	"net"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/johnsudaar/xtouchgma2/link"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

type ConfigurationTab struct {
	gMAIP        *widget.Entry
	gMAUser      *widget.Entry
	gMAPassword  *widget.Entry
	sACNUniverse *widget.Entry
	xtouchesBox  *fyne.Container
	formErrors   *widget.Label
	status       *widget.Label
	logs         *widget.Entry
	start        *widget.Button
	stop         *widget.Button
	addXTouchBtn *widget.Button

	xTouches []*XTouchConfig

	gui *GUI
}

func NewConfigurationTab(g *GUI) *ConfigurationTab {
	tab := &ConfigurationTab{
		gMAIP:        widget.NewEntry(),
		gMAUser:      widget.NewEntry(),
		gMAPassword:  widget.NewEntry(),
		sACNUniverse: widget.NewEntry(),
		xtouchesBox:  container.NewVBox(),
		formErrors:   widget.NewLabel(""),
		status:       widget.NewLabel("Status: stopped"),
		logs:         widget.NewMultiLineEntry(),
		xTouches:     make([]*XTouchConfig, 0),
		gui:          g,
	}

	tab.start = widget.NewButton("Start", tab.onStart)
	tab.stop = widget.NewButton("Stop", tab.onStop)
	tab.addXTouchBtn = widget.NewButton("Add XTouch", tab.onAddXTouch)

	return tab
}

func (g *ConfigurationTab) getTabItem() *container.TabItem {
	return widget.NewTabItem(
		"Configuration",
		container.NewScroll(
			container.NewVBox(
				widget.NewCard("GrandMA configuration", "",
					widget.NewForm(
						widget.NewFormItem("GrandMA IP: ", g.gMAIP),
						widget.NewFormItem("GrandMA User: ", g.gMAUser),
						widget.NewFormItem("GrandMA Password: ", g.gMAPassword),
						widget.NewFormItem("sACN Universe: ", g.sACNUniverse),
					),
				),
				g.xtouchesBox,
				g.addXTouchBtn,
				g.formErrors,
				g.start,
				g.stop,
				g.status,
				g.logs,
			),
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

func (g *ConfigurationTab) onAddXTouch() {
	g.addXTouch("", "0", 0)
}

func (g *ConfigurationTab) addXTouch(xtouchType string, offset string, port int) {
	xtouch := NewXTouchConfig(g)
	xtouch.SetSelectedXTouchType(xtouchType)
	xtouch.executorOffset.Text = offset
	xtouch.setPort(port)
	g.xTouches = append(g.xTouches, xtouch)
	g.xtouchesBox.Add(xtouch.cardItem())
}

func (g *ConfigurationTab) findXTouch(uuid string) int {
	for i, xt := range g.xTouches {
		if xt.UUID() == uuid {
			return i
		}
	}
	return -1
}

func (g *ConfigurationTab) removeXTouch(id string) {
	idx := g.findXTouch(id)
	if idx == -1 {
		return
	}
	xt := g.xTouches[idx]
	g.xTouches = append(g.xTouches[:idx], g.xTouches[idx+1:]...)
	g.xtouchesBox.Remove(xt.cardItem())
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

func (g *ConfigurationTab) nextXTouchPort(source string) int {
	port := 0
	for _, xt := range g.xTouches {
		if xt.UUID() == source {
			continue
		}
		if xt.Type() == xtouch.ServerTypeXTouchExt {
			p := xt.Port()
			if p > port {
				port = p
			}
		}
	}

	if port == 0 {
		port = 5002
	}
	return port + 2
}

func (g *ConfigurationTab) validInputs() bool {
	for _, xt := range g.xTouches {
		valid, err := xt.validInputs()
		if !valid {
			g.formErrors.Text = err
			return false
		}
	}

	xtouchCount := 0
	for _, xt := range g.xTouches {
		if xt.Type() == xtouch.ServerTypeXTouch {
			xtouchCount++
		}
	}

	if xtouchCount > 1 {
		g.formErrors.Text = "Only one xtouch can be connected"
		return false
	}

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
