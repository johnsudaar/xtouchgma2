package link

import (
	"context"
	"time"

	"github.com/Scalingo/go-utils/logger"
)

func (l *Link) SetDMXValue(address int, value byte) error {
	l.dmxLock.Lock()
	defer l.dmxLock.Unlock()

	l.dmxUniverse[address] = value
	return nil
}

func (l *Link) startDMXSync(ctx context.Context) {
	log := logger.Get(ctx)
	log.Info("Start DMX Sync")
	for {
		time.Sleep(50 * time.Millisecond)
		var universe [512]byte
		l.dmxLock.Lock()
		for i, v := range l.dmxUniverse {
			universe[i] = v
		}
		l.dmxLock.Unlock()
		l.stopLock.RLock()
		stop := l.stop
		if !stop {
			l.sacnDMX <- universe
		}
		l.stopLock.RUnlock()
		if stop {
			log.Info("Stop DMX Sync")
			return
		}

	}
}
