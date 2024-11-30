package gui

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func (g *GUI) Start() {
	go g.startLogs()
	g.window.ShowAndRun()
}

func (g *GUI) startLink() {
	if g.link != nil {
		level := logrus.InfoLevel

		if g.settingsTab.debugEnabled.Checked {
			level = logrus.DebugLevel
		}
		hooks := []logrus.Hook{g}

		if g.settingsTab.logToFile.Checked {
			hooks = append(hooks, lfshook.NewHook("xtouchgma2.log", &logrus.TextFormatter{}))
		}

		log := logger.Default(logger.WithHooks(hooks), logger.WithLogLevel(level))
		ctx := logger.ToCtx(context.Background(), log)
		err := g.link.Start(ctx)
		if err != nil {
			g.SetStatus("Fail to start: " + err.Error())
			g.link.Stop(ctx)
			g.configurationTab.enableInputs()
			g.configurationTab.enableStart()
			g.configurationTab.disableStop()
			g.link = nil
		}
	}
}
