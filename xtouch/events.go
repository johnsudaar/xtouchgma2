package xtouch

import "context"

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
	return float64(e.PositionRaw) / 16380
}

func (s *Server) dispatchMidiMessage(ctx context.Context, message MidiMessage) {
	if message.Type == MidiMessageTypePitchWheel {
		s.sendFaderChange(ctx, FaderChangedEvent{
			Fader:       int(message.Channel),
			PositionRaw: message.PitchBend,
		})
	}

	if message.Type == MidiMessageTypeNoteOn {
		s.handleButtonChange(ctx, message)
	}

	if message.Type == MidiMessageTypeControlChange {
		if message.ControllerNumber >= 16 && message.ControllerNumber <= 23 || message.ControllerNumber == 60 {
			s.handleRotaryControlChange(ctx, message)
		}
	}
}

func (s *Server) handleRotaryControlChange(ctx context.Context, message MidiMessage) {
	event := EncoderChangedEvent{}
	if message.ControllerNumber == 60 {
		event.Encoder = MainEncoder
	} else {
		event.Encoder = message.ControllerNumber - 16
	}

	if message.ControlData > 64 {
		event.Delta = 64 - int(message.ControlData)
	} else {
		event.Delta = int(message.ControlData)
	}
	s.sendEncoderChangedEvent(ctx, event)
}

func (s *Server) handleButtonChange(ctx context.Context, message MidiMessage) {
	event := ButtonChangedEvent{
		Status: ButtonStatus(message.Velocity),
	}
	button, ok := noteToButton[message.NoteNumber]
	if ok {
		event.Button = button
		event.Type = ButtonTypeCommand
		if message.NoteNumber >= buttonToNote[ButtonSelect1] && message.NoteNumber <= buttonToNote[ButtonSelect8] {
			event.Type = ButtonTypeSelect
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonSelect1])
		}
		if message.NoteNumber >= buttonToNote[ButtonMute1] && message.NoteNumber <= buttonToNote[ButtonMute8] {
			event.Type = ButtonTypeMute
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonMute1])
		}
		if message.NoteNumber >= buttonToNote[ButtonSolo1] && message.NoteNumber <= buttonToNote[ButtonSolo8] {
			event.Type = ButtonTypeSolo
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonSolo1])
		}
		if message.NoteNumber >= buttonToNote[ButtonSolo1] && message.NoteNumber <= buttonToNote[ButtonSolo8] {
			event.Type = ButtonTypeSolo
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonSolo1])
		}
		if message.NoteNumber >= buttonToNote[ButtonRec1] && message.NoteNumber <= buttonToNote[ButtonRec8] {
			event.Type = ButtonTypeRec
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonRec1])
		}
		if message.NoteNumber >= buttonToNote[ButtonRotary1] && message.NoteNumber <= buttonToNote[ButtonRotary8] {
			event.Type = ButtonTypeRotary
			event.Executor = int(message.NoteNumber) - int(buttonToNote[ButtonRotary1])
		}
		s.sendButtonChange(ctx, event)
	}
}
