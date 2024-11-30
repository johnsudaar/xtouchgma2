package link

import (
	"context"
	"time"

	"github.com/Scalingo/go-utils/logger"
)

func (l *Link) startEventLoop(ctx context.Context) {
	log := logger.Get(ctx)
	log.Info("Start main event loop")
	for {
		l.stopLock.RLock()
		stop := l.stop
		l.stopLock.RUnlock()
		if stop {
			log.Info("Stop main event loop")
			return
		}

		log.Debug("Running event loop")
		time.Sleep(l.eventLoopRefreshRate)

		err := l.faderGmaToXtouch(ctx)
		if err != nil {
			log.WithError(err).Error("fail to sync faders")
		}

		err = l.updateEncoderRings(ctx)
		if err != nil {
			log.WithError(err).Error("fail to update encoder rings")
		}
		err = l.updateButtons(ctx)
		if err != nil {
			log.WithError(err).Error("fail to update encoder rings")
		}

		for _, xtouch := range l.XTouches {
			xtouch.RunRefreshers(ctx)
		}
	}
}
