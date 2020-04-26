package link

import (
	"context"
	"time"

	"github.com/Scalingo/go-utils/logger"
)

func (l *Link) startEventLoop(ctx context.Context) {
	log := logger.Get(ctx)
	for {
		time.Sleep(50 * time.Millisecond)
		err := l.faderGmaToXtouch(ctx)
		if err != nil {
			log.WithError(err).Error("fail to sync faders")
		}
	}
}
