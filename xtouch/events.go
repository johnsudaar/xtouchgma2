package xtouch

import "context"

const (
	MainEncoder = 8
)

type FaderChangedEvent struct {
	Fader       int
	PositionRaw uint16
}

type ButtonChangedEvent struct {
	Button Button
	Status ButtonStatus
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
		button, ok := noteToButton[message.NoteNumber]
		if ok {
			s.sendButtonChange(ctx, ButtonChangedEvent{
				Button: button,
				Status: ButtonStatus(message.Velocity),
			})
		}
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
