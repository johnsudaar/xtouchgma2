package gui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/gofrs/uuid"
	"github.com/johnsudaar/xtouchgma2/link"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

const (
	TypeXTouch         = "XTouch"
	TypeXTouchExtender = "XTouch Extender"
)

// TODO:
// - Check that the executor offset is valid => Done
// - Check that only one XTouch is added => Done
// - Generate the ports
// - Load and save to config
// - Add it to the link

type XTouchConfig struct {
	card             *widget.Card
	xtouchType       *widget.Select
	xtouchPort       *widget.Label
	executorOffset   *widget.Entry
	configurationTab *ConfigurationTab
	uuid             string
	selectedType     string
	port             int
}

func NewXTouchConfig(configurationTab *ConfigurationTab) *XTouchConfig {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	config := &XTouchConfig{
		xtouchPort:       widget.NewLabel(""),
		executorOffset:   widget.NewEntry(),
		configurationTab: configurationTab,
		uuid:             id.String(),
	}
	config.xtouchType = widget.NewSelect([]string{
		"XTouch",
		"XTouch Extender",
	}, config.setSelectedXTouchTypeRaw)
	return config
}

func (g *XTouchConfig) cardItem() *widget.Card {
	if g.card == nil {
		g.card = widget.NewCard("XTouch configuration", "",
			container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("XTouch Type: ", g.xtouchType),
					widget.NewFormItem("XTouch Port: ", g.xtouchPort),
					widget.NewFormItem("Executor Offset: ", g.executorOffset),
				),
				widget.NewButton("Remove", g.remove),
			),
		)
	}
	return g.card
}

func (g *XTouchConfig) UUID() string {
	return g.uuid
}

func (g *XTouchConfig) Port() int {
	return g.port
}

func (g *XTouchConfig) remove() {
	g.configurationTab.removeXTouch(g.uuid)
}

func (g *XTouchConfig) setSelectedXTouchTypeRaw(value string) {
	g.selectedType = value
	if g.Type() == xtouch.ServerTypeXTouch {
		g.setPort(10111)
	} else {
		g.setPort(g.configurationTab.nextXTouchPort(g.UUID()))
	}
}

func (g *XTouchConfig) SetSelectedXTouchType(value string) {
	g.xtouchType.SetSelected(value)
	g.setSelectedXTouchTypeRaw(value)
}

func (g *XTouchConfig) validInputs() (bool, string) {
	executorOffset, err := strconv.Atoi(g.executorOffset.Text)
	if err != nil {
		return false, fmt.Sprintf("Executor offset: %s is not a number", g.executorOffset.Text)
	}

	if executorOffset < 0 {
		return false, "Executor offset should be >= 0"
	}
	if executorOffset > 40 {
		return false, "Executor offset should be <= 40"
	}

	if g.selectedType == "" {
		return false, "You should select a XTouch type"
	}

	if g.port == 0 {
		return false, "No port generated for an xtouch"
	}

	return true, ""
}

func (g *XTouchConfig) Type() xtouch.ServerType {
	switch g.selectedType {
	case TypeXTouch:
		return xtouch.ServerTypeXTouch
	case TypeXTouchExtender:
		return xtouch.ServerTypeXTouchExt
	}
	return xtouch.ServerType("")
}

func (g *XTouchConfig) setPort(port int) {
	g.port = port
	g.xtouchPort.Text = strconv.Itoa(port)
}

func (g *XTouchConfig) toParams() link.XTouchParams {
	executorOffset, _ := strconv.Atoi(g.executorOffset.Text)
	return link.XTouchParams{
		Type:           g.Type(),
		Port:           g.Port(),
		ExecutorOffset: executorOffset,
	}
}
