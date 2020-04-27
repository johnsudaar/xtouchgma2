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
	gMAIP        *widget.Entry
	gMAUser      *widget.Entry
	gMAPassword  *widget.Entry
	sACNUniverse *widget.Entry
	status       *widget.Label
	formErrors   *widget.Label
	logs         *widget.Entry
	start        *widget.Button
	stop         *widget.Button
	app          fyne.App
	window       fyne.Window
	link         *link.Link
	logChan      chan []string
}

func New() *GUI {
	gui := &GUI{
		gMAIP:        widget.NewEntry(),
		gMAUser:      widget.NewEntry(),
		gMAPassword:  widget.NewPasswordEntry(),
		sACNUniverse: widget.NewEntry(),
		status:       widget.NewLabel("Status: stopped"),
		formErrors:   widget.NewLabel(""),
		logs:         widget.NewMultiLineEntry(),
		logChan:      make(chan []string, 10),
	}

	gui.start = widget.NewButton("Start", gui.onStart)
	gui.stop = widget.NewButton("Stop", gui.onStop)
	gui.app = app.New()
	gui.window = gui.app.NewWindow("XTouch2GMA")

	gui.gMAIP.SetPlaceHolder("192.168.1.21")
	gui.gMAUser.SetPlaceHolder("john")
	gui.gMAPassword.SetPlaceHolder("john")
	gui.sACNUniverse.SetPlaceHolder("10")

	gui.buildApp()

	return gui
}

func (g *GUI) buildApp() {
	g.window.SetContent(
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
		buffer = append(logs, buffer...)
		if len(buffer) > 1000 {
			buffer = buffer[:1000]
		}
		g.logs.Text = strings.Join(buffer, "\n")
		g.logs.Refresh()
	}
}
