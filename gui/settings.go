package gui

import (
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

func (g *GUI) saveSettings() {
	cfg := ini.Empty()
	cfg.NewSection("GMA")
	cfg.Section("GMA").NewKey("IP", g.gMAIP.Text)
	cfg.Section("GMA").NewKey("User", g.gMAUser.Text)
	cfg.Section("GMA").NewKey("Password", g.gMAPassword.Text)
	cfg.NewSection("SACN")
	cfg.Section("SACN").NewKey("Universe", g.sACNUniverse.Text)
	cfg.NewSection("ENCODERS")
	for i := 0; i < 8; i++ {
		cfg.Section("ENCODERS").NewKey(strconv.Itoa(i+1), g.encoderAttributes[i].Text)
	}
	if g.mapEncodersToAttributes.Checked {
		cfg.Section("ENCODERS").NewKey("MAP_ENCODER_TO_ATTRIBUTES", "true")
	} else {
		cfg.Section("ENCODERS").NewKey("MAP_ENCODER_TO_ATTRIBUTES", "false")
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

	g.gMAIP.Text = cfg.Section("GMA").Key("IP").String()
	g.gMAUser.Text = cfg.Section("GMA").Key("User").String()
	g.gMAPassword.Text = cfg.Section("GMA").Key("Password").String()
	g.sACNUniverse.Text = cfg.Section("SACN").Key("Universe").String()
	for i := 0; i < 8; i++ {
		g.encoderAttributes[i].Text = cfg.Section("ENCODERS").Key(strconv.Itoa(i + 1)).String()
	}

	g.mapEncodersToAttributes.Checked = cfg.Section("ENCODERS").Key("MAP_ENCODER_TO_ATTRIBUTES").String() == "true"

	return nil
}
