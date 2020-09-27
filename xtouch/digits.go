package xtouch

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

/*
*0*
5 1
*6*
4 2
*3*
*/
var digitTo7Seg = []byte{
	0b0111111,
	0b0000110,
	0b1011011,
	0b1001111,
	0b1100110,
	0b1101101,
	0b1111101,
	0b0000111,
	0b1111111,
	0b1101111,
}

func (s *Server) SetAssignement(ctx context.Context, value int) error {
	left := (value / 10) % 10
	right := value % 10

	err := s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: 96,
		ControlData:      digitTo7Seg[left],
	})
	if err != nil {
		return errors.Wrap(err, "fail to send first assign message")
	}
	err = s.transport.SendMidiPacket(ctx, transport.MidiMessage{
		Type:             transport.MidiMessageTypeControlChange,
		ControllerNumber: 97,
		ControlData:      digitTo7Seg[right],
	})
	if err != nil {
		return errors.Wrap(err, "fail to send second assign message")
	}
	return nil
}
