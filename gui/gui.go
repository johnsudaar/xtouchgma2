package gui

import (
	"context"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/johnsudaar/xtouchgma2/link"
)

type GUI struct {
	configurationTab *ConfigurationTab
	encoderTab       *EncoderTab
	app              fyne.App
	window           fyne.Window
	link             *link.Link
	logChan          chan []string
}

func New() *GUI {
	gui := &GUI{
		logChan: make(chan []string, 10),
	}
	gui.encoderTab = NewEncoderTab(gui)
	gui.configurationTab = NewConfigurationTab(gui)

	gui.app = app.New()
	gui.window = gui.app.NewWindow("XTouch2GMA")

	gui.buildApp(context.Background())

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

func (g *GUI) buildApp(ctx context.Context) {
	g.window.SetContent(
		widget.NewTabContainer(
			g.configurationTab.getTabItem(),
			g.encoderTab.getTabItem(),
		),
	)

	g.window.SetOnClosed(func() {
		go func() {
			time.Sleep(1 * time.Second)
			panic("QUIT")
		}()
		if g.link != nil {
			g.link.Stop(ctx)
		}
	})

	g.configurationTab.onWindowsInit()
}

func (g *GUI) startRequested() {
	g.configurationTab.disableInputs()
	g.configurationTab.disableStart()
	g.SetStatus("Connecting")

	link, err := link.New(g.configurationTab.linkParams())

	if err != nil {
		g.configurationTab.enableInputs()
		g.configurationTab.enableStart()
		g.SetStatus("Fail to init: " + err.Error())
		return
	}
	g.link = link

	g.configurationTab.enableStop()
	g.saveSettings()
	g.encoderTab.updateEncoders()
	go g.startLink()
	g.SetStatus("")
}

func (g *GUI) stopRequested() {
	if g.link != nil {
		g.link.Stop(context.Background())
	}
	g.configurationTab.enableInputs()
	g.configurationTab.disableStop()
	g.configurationTab.enableStart()
	g.SetStatus("stopped")
}
