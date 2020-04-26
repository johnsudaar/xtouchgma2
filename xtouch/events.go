package xtouch

import "context"

type FaderChangedEvent struct {
	Fader       int
	PositionRaw uint16
}

type ButtonChangedEvent struct {
	Button Button
	Status ButtonStatus
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
}
