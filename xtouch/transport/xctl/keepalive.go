package xctl

import (
	"context"
	"time"

	"github.com/Scalingo/go-utils/logger"
)

func (s *XCtl) keepAliveLoop(ctx context.Context) {
	for {
		s.stopLock.Lock()
		stop := s.stop
		s.stopLock.Unlock()
		if stop {
			return
		}
		time.Sleep(2 * time.Second)
		s.keepAlive(ctx)
	}
}

func (s *XCtl) keepAlive(ctx context.Context) {
	s.socketLock.Lock()
	defer s.socketLock.Unlock()
	log := logger.Get(ctx)
	if s.client == nil {
		return
	}

	_, err := s.conn.WriteTo([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xf7,
	}, s.client)
	if err != nil {
		log.WithError(err).Error("fail to send xtouch heartbeat")
	}
}
