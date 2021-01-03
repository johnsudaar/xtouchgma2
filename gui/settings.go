package gui

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

func (g *GUI) saveSettings() {
	cfg := ini.Empty()
	cfg.NewSection("GMA")
	cfg.Section("GMA").NewKey("IP", g.configurationTab.gMAIP.Text)
	cfg.Section("GMA").NewKey("User", g.configurationTab.gMAUser.Text)
	cfg.Section("GMA").NewKey("Password", g.configurationTab.gMAPassword.Text)
	cfg.NewSection("SACN")
	cfg.Section("SACN").NewKey("Universe", g.configurationTab.sACNUniverse.Text)
	cfg.NewSection("ENCODERS")
	for i := 0; i < 8; i++ {
		cfg.Section("ENCODERS").NewKey(strconv.Itoa(i+1), g.encoderTab.attributes[i].Text)
	}

	for i, xt := range g.configurationTab.xTouches {
		sec := cfg.Section(fmt.Sprintf("XTOUCH.%d", i))
		sec.NewKey("Type", xt.selectedType)
		sec.NewKey("ExecutorOffset", xt.executorOffset.Text)
		sec.NewKey("Port", strconv.Itoa(xt.Port()))
	}

	err := cfg.SaveTo("xtouch2gma.ini")
	if err != nil {
		g.logChan <- []string{"Fail to save config file: " + err.Error()}
	}
}

func (g *GUI) loadSettings() error {
	cfg, err := ini.LooseLoad("xtouch2gma.ini")
	if err != nil {
		return errors.Wrap(err, "fail to load settings")
	}

	for _, sec := range cfg.Section("XTOUCH").ChildSections() {
		port, _ := strconv.Atoi(sec.Key("Port").String())
		g.configurationTab.addXTouch(
			sec.Key("Type").String(),
			sec.Key("ExecutorOffset").String(),
			port,
		)
	}

	g.configurationTab.gMAIP.Text = cfg.Section("GMA").Key("IP").String()
	g.configurationTab.gMAUser.Text = cfg.Section("GMA").Key("User").String()
	g.configurationTab.gMAPassword.Text = cfg.Section("GMA").Key("Password").String()
	g.configurationTab.sACNUniverse.Text = cfg.Section("SACN").Key("Universe").String()
	for i := 0; i < 8; i++ {
		g.encoderTab.attributes[i].Text = cfg.Section("ENCODERS").Key(strconv.Itoa(i + 1)).String()
	}

	return nil
}
