package link

import (
	"context"
	"strings"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (l *Link) faderGmaToXtouch(ctx context.Context) error {
	log := logger.Get(ctx)
	l.faderLock.Lock()
	page := l.faderPage
	l.faderLock.Unlock()

	playbacks, err := l.GMA.Playbacks(page, []gma2ws.PlaybacksRange{
		gma2ws.PlaybacksRange{
			Index: 0,
			Count: 10,
		},
	})
	if err != nil {
		return errors.Wrap(err, "fail to sync GMA faders")
	}

	for i := 0; i < 8; i++ {
		log := log.WithFields(logrus.Fields{
			"fader": i,
		})
		executor := playbacks[0].Items[i/5][i%5]
		f := executor.ExecutorBlocks[0].Fader
		var value float64
		if f.Max != 0 {
			value = float64(f.Value) / float64(f.Max-f.Min)
		}
		err := l.XTouch.SetFaderPos(ctx, i, value)
		if err != nil {
			log.WithError(err).Error("fail to send fader to its position")
		}
		line1 := executor.TextTop.Text
		line2 := ""
		if len(executor.Cues.Items) == 3 {
			line2 = executor.Cues.Items[1].Text
		} else if len(executor.Cues.Items) >= 1 {
			line2 = executor.Cues.Items[0].Text
		}
		color, err := xtouch.ClosestScribbleColor(f.BorderColor)
		if err != nil {
			log.WithError(err).Error("fail to get scribble color")
			color = xtouch.ScribbleColorWhite
		}

		err = l.XTouch.SetScribble(ctx, i, color, true, strings.TrimSpace(line1), strings.TrimSpace(line2))
		if err != nil {
			log.WithError(err).Error("fail to send scribble data")
		}
	}

	l.XTouch.SetAssignement(ctx, page+1)
	return nil
}

func (l *Link) onFaderChangeEvent(ctx context.Context, e xtouch.FaderChangedEvent) {
	log := logger.Get(ctx)
	err := l.GMA.FaderChanged(ctx, e.Fader, 0, e.Position())
	if err != nil {
		log.WithError(err).Error("fail to send fader position")
	}
}

func (l *Link) FaderPageUp() {
	l.faderLock.Lock()
	defer l.faderLock.Unlock()
	l.faderPage++
}

func (l *Link) FaderPageDown() {
	l.faderLock.Lock()
	defer l.faderLock.Unlock()
	l.faderPage--
	if l.faderPage < 0 {
		l.faderPage = 0
	}
}
