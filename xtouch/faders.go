package xtouch

import (
	"context"
	"fmt"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

const (
	MAIN_FADER = 8
)

func (s *Server) SetFaderPos(ctx context.Context, fader int, pos float64) error {
	if s.serverType == ServerTypeXTouch {
		return s.setFaderPosXtouch(ctx, fader, pos)
	}
	if s.serverType == ServerTypeXTouchExt {
		return s.setFaderPosXTouchExt(ctx, fader, pos)
	}
	return fmt.Errorf("invalid server type: %s", s.serverType)
}

func (s *Server) setFaderPosXtouch(ctx context.Context, fader int, pos float64) error {
	posInt := uint16(pos * 16383) // 2^7
	err := s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:      transport.MidiMessageTypePitchWheel,
		Channel:   byte(fader),
		PitchBend: posInt,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send xtouch packet for fader position")
	}
	return nil
}

func (s *Server) setFaderPosXTouchExt(ctx context.Context, fader int, pos float64) error {
	posByte := byte(pos * 127)
	err := s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: 70 + byte(fader),
		ControlData:      posByte,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send xtouch extender packet for fader position")
	}
	return nil
}

func (s *Server) SetRingPosition(ctx context.Context, fader int, pos float64) error {
	if s.serverType == ServerTypeXTouch {
		return s.setRingPositionXTouch(ctx, fader, pos)
	}
	if s.serverType == ServerTypeXTouchExt {
		return s.setRingPositionXTouchExt(ctx, fader, pos)
	}
	return fmt.Errorf("invalid server type: %s", s.serverType)
}

func (s *Server) setRingPositionXTouchExt(ctx context.Context, fader int, pos float64) error {
	posByte := byte(pos * 127)
	err := s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: 80 + byte(fader),
		ControlData:      posByte,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send xtouch extender packet for ring position")
	}
	return nil

}

func (s *Server) setRingPositionXTouch(ctx context.Context, fader int, pos float64) error {
	pos *= 2 ^ 13 - 1
	res := uint16(0)
	for i := 0; float64(i) < pos; i++ {
		res += 1 << i
	}
	byte1 := byte(res & 0b1111111)
	byte2 := byte((res & 0b1111110000000) >> 7)
	err := s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: byte(fader) + 48,
		ControlData:      byte1,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send first rotation encoder")
	}
	err = s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: byte(fader) + 56,
		ControlData:      byte2,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send first rotation encoder")
	}
	return nil
}
