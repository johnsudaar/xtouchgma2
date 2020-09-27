package xtouch

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
)

const (
	MainEncoder = 8
)

type ButtonType string

const (
	ButtonTypeCommand ButtonType = "command"
	ButtonTypeSelect  ButtonType = "select"
	ButtonTypeMute    ButtonType = "mute"
	ButtonTypeSolo    ButtonType = "solo"
	ButtonTypeRec     ButtonType = "rec"
	ButtonTypeRotary  ButtonType = "rotary"
)

type FaderChangedEvent struct {
	Fader       int
	PositionRaw uint16
	MaxPosition uint16
}

type ButtonChangedEvent struct {
	Type     ButtonType
	Button   Button
	Executor int
	Status   ButtonStatus
}

type EncoderChangedEvent struct {
	Encoder byte
	Delta   int
}

func (e FaderChangedEvent) Position() float64 {
	return float64(e.PositionRaw) / float64(e.MaxPosition)
}

func (s *Server) dispatchMidiMessage(ctx context.Context, message transport.MidiMessage) {
	if s.serverType == ServerTypeXTouch && message.Type == transport.MidiMessageTypePitchWheel {
		s.sendFaderChange(ctx, FaderChangedEvent{
			Fader:       int(message.Channel),
			PositionRaw: message.PitchBend,
			MaxPosition: 16380,
		})
	}

	if message.Type == transport.MidiMessageTypeNoteOn {
		s.handleButtonChange(ctx, message)
	}

	if message.Type == transport.MidiMessageTypeControlChange {
		if s.serverType == ServerTypeXTouch && message.ControllerNumber == 60 {
			s.handleRotaryControlChange(ctx, message)
		}

		if s.serverType == ServerTypeXTouch && message.ControllerNumber >= 16 && message.ControllerNumber <= 23 {
			s.handleRotaryControlChange(ctx, message)
		}

		if s.serverType == ServerTypeXTouchExt && message.ControllerNumber >= 80 && message.ControllerNumber <= 87 {
			s.handleRotaryControlChange(ctx, message)
		}

		if s.serverType == ServerTypeXTouchExt && message.ControllerNumber >= 70 && message.ControllerNumber <= 87 {
			s.sendFaderChange(ctx, FaderChangedEvent{
				Fader:       int(message.ControllerNumber - 70),
				PositionRaw: uint16(message.ControlData),
				MaxPosition: 127,
			})
		}
	}
}

func (s *Server) handleRotaryControlChange(ctx context.Context, message transport.MidiMessage) {
	event := EncoderChangedEvent{}
	if message.ControllerNumber == 60 {
		event.Encoder = MainEncoder
	} else {
		if s.serverType == ServerTypeXTouch {
			event.Encoder = message.ControllerNumber - 16
		}
		if s.serverType == ServerTypeXTouchExt {
			event.Encoder = message.ControllerNumber - 80
		}
	}

	if message.ControlData > 64 {
		event.Delta = 64 - int(message.ControlData)
	} else {
		event.Delta = int(message.ControlData)
	}

	if s.serverType == ServerTypeXTouchExt {
		event.Delta *= -1
	}

	s.sendEncoderChangedEvent(ctx, event)
}

func (s *Server) handleButtonChange(ctx context.Context, message transport.MidiMessage) {
	event := ButtonChangedEvent{
		Status: ButtonStatus(message.Velocity),
	}
	button, ok := s.noteToButton[message.NoteNumber]
	if ok {
		event.Button = button
		event.Type = ButtonTypeCommand
		if message.NoteNumber >= s.buttonToNote[ButtonSelect1] && message.NoteNumber <= s.buttonToNote[ButtonSelect8] {
			event.Type = ButtonTypeSelect
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonSelect1])
		}
		if message.NoteNumber >= s.buttonToNote[ButtonMute1] && message.NoteNumber <= s.buttonToNote[ButtonMute8] {
			event.Type = ButtonTypeMute
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonMute1])
		}
		if message.NoteNumber >= s.buttonToNote[ButtonSolo1] && message.NoteNumber <= s.buttonToNote[ButtonSolo8] {
			event.Type = ButtonTypeSolo
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonSolo1])
		}
		if message.NoteNumber >= s.buttonToNote[ButtonSolo1] && message.NoteNumber <= s.buttonToNote[ButtonSolo8] {
			event.Type = ButtonTypeSolo
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonSolo1])
		}
		if message.NoteNumber >= s.buttonToNote[ButtonRec1] && message.NoteNumber <= s.buttonToNote[ButtonRec8] {
			event.Type = ButtonTypeRec
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonRec1])
		}
		if message.NoteNumber >= s.buttonToNote[ButtonRotary1] && message.NoteNumber <= s.buttonToNote[ButtonRotary8] {
			event.Type = ButtonTypeRotary
			event.Executor = int(message.NoteNumber) - int(s.buttonToNote[ButtonRotary1])
		}
		s.sendButtonChange(ctx, event)
	}
}
