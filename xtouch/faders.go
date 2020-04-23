package xtouch

import "github.com/pkg/errors"

const (
	MAIN_FADER = 8
)

func (s *Server) SetFaderPos(fader int, pos float64) error {
	posInt := uint16(pos * 16384) // 2^7
	message := []byte{
		0xE0 + byte(fader),
		byte(posInt % 128),
		byte(posInt / 128),
	}
	err := s.SendMidiPacket(message)
	if err != nil {
		return errors.Wrap(err, "fail to send packet for packet position")
	}
	return nil
}
