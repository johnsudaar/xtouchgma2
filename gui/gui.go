package gui

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/link"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GUI struct {
	gMAIP          *widget.Entry
	gMAUser        *widget.Entry
	gMAPassword    *widget.Entry
	sACNUniverse   *widget.Entry
	status         *widget.Label
	formErrors     *widget.Label
	logs           *widget.Entry
	start          *widget.Button
	stop           *widget.Button
	updateEncoders *widget.Button

	encoderAttributes []*widget.Entry
	app               fyne.App
	window            fyne.Window
	link              *link.Link
	logChan           chan []string
}

func New() *GUI {
	gui := &GUI{
		gMAIP:             widget.NewEntry(),
		gMAUser:           widget.NewEntry(),
		gMAPassword:       widget.NewPasswordEntry(),
		sACNUniverse:      widget.NewEntry(),
		status:            widget.NewLabel("Status: stopped"),
		formErrors:        widget.NewLabel(""),
		encoderAttributes: make([]*widget.Entry, 8),
		logs:              widget.NewMultiLineEntry(),
		logChan:           make(chan []string, 10),
	}

	for i := 0; i < 8; i++ {
		gui.encoderAttributes[i] = widget.NewEntry()
	}

	gui.start = widget.NewButton("Start", gui.onStart)
	gui.stop = widget.NewButton("Stop", gui.onStop)
	gui.updateEncoders = widget.NewButton("Update", gui.onUpdateEncoders)
	gui.app = app.New()
	gui.window = gui.app.NewWindow("XTouch2GMA")

	gui.gMAIP.SetPlaceHolder("192.168.1.21")
	gui.gMAUser.SetPlaceHolder("john")
	gui.gMAPassword.SetPlaceHolder("john")
	gui.sACNUniverse.SetPlaceHolder("10")

	gui.buildApp()

	err := gui.loadSettings()
	if err != nil {
		gui.window.SetContent(
			widget.NewVBox(
				widget.NewLabel("Error: "+err.Error()),
				widget.NewButton("Ok", func() {
					gui.app.Quit()
				}),
			),
		)
		gui.window.ShowAndRun()
		panic(err)
	}

	return gui
}

func (g *GUI) buildApp() {
	g.window.SetContent(
		widget.NewTabContainer(
			widget.NewTabItem(
				"Configuration",
				widget.NewVBox(
					widget.NewForm(
						&widget.FormItem{
							Text:   "GrandMA IP: ",
							Widget: g.gMAIP,
						},
						&widget.FormItem{
							Text:   "GrandMA User: ",
							Widget: g.gMAUser,
						},
						&widget.FormItem{
							Text:   "GrandMA Password: ",
							Widget: g.gMAPassword,
						},
						&widget.FormItem{
							Text:   "sACN Universe: ",
							Widget: g.sACNUniverse,
						},
					),
					g.formErrors,
					g.start,
					g.stop,
					g.status,
					g.logs,
				),
			),
			widget.NewTabItem(
				"Encoders",
				widget.NewVBox(
					widget.NewForm(
						widget.NewFormItem("Encoder 1 attribute: ", g.encoderAttributes[0]),
						widget.NewFormItem("Encoder 2 attribute: ", g.encoderAttributes[1]),
						widget.NewFormItem("Encoder 3 attribute: ", g.encoderAttributes[2]),
						widget.NewFormItem("Encoder 4 attribute: ", g.encoderAttributes[3]),
						widget.NewFormItem("Encoder 5 attribute: ", g.encoderAttributes[4]),
						widget.NewFormItem("Encoder 6 attribute: ", g.encoderAttributes[5]),
						widget.NewFormItem("Encoder 7 attribute: ", g.encoderAttributes[6]),
						widget.NewFormItem("Encoder 8 attribute: ", g.encoderAttributes[7]),
					),
					g.updateEncoders,
				),
			),
		),
	)
	g.logs.SetReadOnly(true)
	g.stop.Disable()
	g.window.SetOnClosed(func() {
		if g.link != nil {
			g.link.Stop()
		}
	})
}

func (g *GUI) Start() {
	go g.startLogs()
	g.window.ShowAndRun()
}

func (g *GUI) onStart() {
	g.formErrors.Text = ""
	valid := g.validInputs()
	g.formErrors.Refresh()
	if !valid {
		return
	}

	g.disableInputs()
	g.start.Disable()
	g.SetStatus("Connecting")

	u, _ := strconv.Atoi(g.sACNUniverse.Text)

	link, err := link.New(link.NewLinkParams{
		GMAHost:      g.gMAIP.Text,
		GMAUser:      g.gMAUser.Text,
		GMAPassword:  g.gMAPassword.Text,
		SACNUniverse: uint16(u),
	})
	if err != nil {
		g.enableInputs()
		g.start.Enable()
		g.SetStatus("Fail to init: " + err.Error())
		return
	}
	g.link = link

	g.stop.Enable()
	g.saveSettings()
	g.updateLinkEncoders()
	go g.startLink()
	g.SetStatus("Connected!")
}

func (g *GUI) startLink() {
	if g.link != nil {
		log := logger.Default(logger.WithHooks([]logrus.Hook{g}))
		ctx := logger.ToCtx(context.Background(), log)
		err := g.link.Start(ctx)
		if err != nil {
			g.SetStatus("Fail to start: " + err.Error())
			g.link.Stop()
			g.enableInputs()
			g.start.Enable()
			g.stop.Disable()
			g.link = nil
		}
	}
}

func (g *GUI) disableInputs() {
	g.gMAIP.Disable()
	g.gMAUser.Disable()
	g.gMAPassword.Disable()
	g.sACNUniverse.Disable()
}

func (g *GUI) enableInputs() {
	g.gMAIP.Enable()
	g.gMAUser.Enable()
	g.gMAPassword.Enable()
	g.sACNUniverse.Enable()
}

func (g *GUI) onStop() {
	if g.link != nil {
		g.link.Stop()
	}
	g.enableInputs()
	g.stop.Disable()
	g.start.Enable()
	g.SetStatus("stopped")
}

func (g *GUI) onUpdateEncoders() {
	g.saveSettings()
	g.updateLinkEncoders()
}

func (g *GUI) updateLinkEncoders() {
	if g.link == nil {
		return
	}

	g.link.SetEncoderAttributes([8]string{
		g.encoderAttributes[0].Text,
		g.encoderAttributes[1].Text,
		g.encoderAttributes[2].Text,
		g.encoderAttributes[3].Text,
		g.encoderAttributes[4].Text,
		g.encoderAttributes[5].Text,
		g.encoderAttributes[6].Text,
		g.encoderAttributes[7].Text,
	})
}

func (g *GUI) SetStatus(status string) {
	g.status.Text = "Status : " + status
	g.status.Refresh()
}

func (g *GUI) validInputs() bool {
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

// Logrus hook

func (g *GUI) Levels() []logrus.Level {
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
	// TODO: debug opts

	return levels
}

func (g *GUI) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return nil
	}
	formater := logrus.TextFormatter{DisableColors: true}
	formatted, err := formater.Format(entry)
	if err != nil {
		return errors.Wrap(err, "fail to format")
	}

	values := strings.Split(string(formatted), "\n")
	g.logChan <- values

	return nil
}

func (g *GUI) startLogs() {
	buffer := []string{}
	for logs := range g.logChan {
		logs := removeEmpty(logs)
		buffer = append(logs, buffer...)
		if len(buffer) > 15 {
			buffer = buffer[:15]
		}
		g.logs.Text = strings.Join(buffer, "\n")
		g.logs.Refresh()
	}
}

func removeEmpty(logs []string) []string {
	res := make([]string, 0)
	for _, s := range logs {
		if s == "" {
			continue
		}
		res = append(res, s)
	}
	return res
}
