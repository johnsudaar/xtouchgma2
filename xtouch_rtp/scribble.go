package xtouch

import (
	"bytes"

	"github.com/pkg/errors"
)

type ScribbleColor byte

const (
	ScribbleColorRed    ScribbleColor = 0x01
	ScribbleColorGreen  ScribbleColor = 0x02
	ScribbleColorYellow ScribbleColor = 0x03
	ScribbleColorBlue   ScribbleColor = 0x04
	ScribbleColorPink   ScribbleColor = 0x05
	ScribbleColorCyan   ScribbleColor = 0x06
	ScribbleColorWhite  ScribbleColor = 0x07
)

func (s *Server) SetScribble(channel int, color ScribbleColor, secondLineInverted bool, line1 string, line2 string) error {
	if len(line1) > 7 {
		line1 = line1[:7]
	}
	if len(line2) > 7 {
		line2 = line2[:7]
	}

	buff := new(bytes.Buffer)
	buff.WriteByte(0x20 + byte(channel))
	if secondLineInverted {
		buff.WriteByte(byte(color) + 0x40)
	} else {
		buff.WriteByte(byte(color))
	}

	buff.Write(toExactly7Char(line1))
	buff.Write(toExactly7Char(line2))

	err := s.SendSysExPacket(buff.Bytes())
	if err != nil {
		return errors.Wrap(err, "fail to send scribble packet")
	}

	return nil
}

func toExactly7Char(a string) []byte {
	buffer := new(bytes.Buffer)
	for i := 0; i < 7; i++ {
		if i < len(a) {
			buffer.WriteByte(a[i])
		} else {
			buffer.WriteByte(0)
		}
	}
	return buffer.Bytes()
}
