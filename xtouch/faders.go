package xtouch

import (
	"context"

	"github.com/pkg/errors"
)

const (
	MAIN_FADER = 8
)

func (s *Server) SetFaderPos(ctx context.Context, fader int, pos float64) error {
	posInt := uint16(pos * 16383) // 2^7
	err := s.SendMidiPacket(ctx, MidiMessage{
		Type:      MidiMessageTypePitchWheel,
		Channel:   byte(fader),
		PitchBend: posInt,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send packet for packet position")
	}
	return nil
}

func (s *Server) SetRingPosition(ctx context.Context, fader int, pos float64) {
	pos *= 2 ^ 13 - 1
	res := uint16(0)
	for i := 0; float64(i) < pos; i++ {
		res += 1 << i
	}
	byte1 := byte(res & 0b1111111)
	byte2 := byte((res & 0b1111110000000) >> 7)
	s.SendMidiPacket(ctx, MidiMessage{
		Type:             MidiMessageTypeControlChange,
		ControllerNumber: byte(fader) + 48,
		ControlData:      byte1,
	})
	s.SendMidiPacket(ctx, MidiMessage{
		Type:             MidiMessageTypeControlChange,
		ControllerNumber: byte(fader) + 56,
		ControlData:      byte2,
	})
}
