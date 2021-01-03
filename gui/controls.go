package gui

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/sirupsen/logrus"
)

func (g *GUI) Start() {
	go g.startLogs()
	g.window.ShowAndRun()
}

func (g *GUI) startLink() {
	if g.link != nil {
		log := logger.Default(logger.WithHooks([]logrus.Hook{g}))
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
