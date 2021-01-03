package xctl

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

func (s *XCtl) sendRawPacket(ctx context.Context, buff []byte) error {
	s.socketLock.Lock()
	defer s.socketLock.Unlock()
	if s.client == nil {
		return errors.New("not connected")
	}
	log := logger.Get(ctx)
	log.WithField("send_to", s.client).Debug(hex.Dump(buff))
	_, err := s.conn.WriteToUDP(buff, s.client)
	if err != nil {
		return errors.Wrap(err, "fail to send message")
	}
	return nil
}

func (s *XCtl) SendSysExPacket(ctx context.Context, message []byte) error {
	buff := new(bytes.Buffer)
	buff.Write([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x58,
	})

	buff.Write(message)
	buff.WriteByte(0xf7)
	err := s.sendRawPacket(ctx, buff.Bytes())
	if err != nil {
		return errors.Wrap(err, "fail to send sysex packet")
	}
	return nil
}

func (s *XCtl) SendMidiPacket(ctx context.Context, packet transport.MidiMessage) error {
	buff, err := packet.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "fail to marshal midi packet")
	}

	err = s.sendRawPacket(ctx, buff)
	if err != nil {
		return errors.Wrap(err, "fail to send midi packet")
	}
	return nil
}
