package xctl

import (
	"context"
	"encoding/hex"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
)

func (s *XCtl) readLoop(ctx context.Context) {
	log := logger.Get(ctx)
	buffer := make([]byte, 1024)
	for {
		s.stopLock.Lock()
		stop := s.stop
		s.stopLock.Unlock()
		if stop {
			return
		}
		n, from, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			log.WithError(err).Error("fail to read udp buffer")
			continue
		}
		s.socketLock.Lock()
		s.client = from
		s.socketLock.Unlock()

		packet := buffer[:n]

		log.WithField("from", from).Debug(hex.Dump(packet))
		if buffer[0] < 0xf0 {
			var midiMessage transport.MidiMessage
			midiMessage.UnmarshalBinary(buffer)
			err := s.reader.OnUDPPacket(ctx, from, midiMessage)
			if err != nil {
				log.WithError(err).Error("reader failed to process midi message")
			}
		}
	}
}
