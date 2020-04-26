package xtouch

import (
	"context"

	"github.com/pkg/errors"
)

type Button string
type ButtonStatus byte
type FaderButtonPosition byte

const (
	ButtonTrack  Button = "track"
	ButtonPan    Button = "pan"
	ButtonEQ     Button = "eq"
	ButtonSend   Button = "send"
	ButtonPlugin Button = "plugin"
	ButtonInst   Button = "inst"

	ButtonName     Button = "name"
	ButtonTimecode Button = "timecode"

	ButtonGlobalView  Button = "global_view"
	ButtonMidiTracks  Button = "midi_tracks"
	ButtonInputs      Button = "inputs"
	ButtonAudioTracks Button = "audio_tracks"
	ButtonAudioInst   Button = "audio_inst"
	ButtonAux         Button = "aux"
	ButtonBuses       Button = "buses"
	ButtonOutputs     Button = "outputs"
	ButtonUser        Button = "user"

	ButtonF1 Button = "f1"
	ButtonF2 Button = "f2"
	ButtonF3 Button = "f3"
	ButtonF4 Button = "f4"
	ButtonF5 Button = "f5"
	ButtonF6 Button = "f6"
	ButtonF7 Button = "f7"
	ButtonF8 Button = "f8"

	ButtonShift   Button = "shift"
	ButtonOption  Button = "option"
	ButtonControl Button = "control"
	ButtonAlt     Button = "alt"

	ButtonReadOff Button = "read_off"
	ButtonWrite   Button = "write"
	ButtonTrim    Button = "trim"
	ButtonTouch   Button = "touch"
	ButtonLatch   Button = "latch"
	ButtonGroup   Button = "group"

	ButtonSave   Button = "save"
	ButtonUndo   Button = "undo"
	ButtonCancel Button = "cancel"
	ButtonEnter  Button = "enter"

	ButtonMarker  Button = "marker"
	ButtonNudge   Button = "nudge"
	ButtonCycle   Button = "cycle"
	ButtonDrop    Button = "drop"
	ButtonReplace Button = "replace"
	ButtonClick   Button = "click"
	ButtonSolo    Button = "solo"

	ButtonFaderPrev   Button = "fader_prev"
	ButtonFaderNext   Button = "fader_next"
	ButtonChannelPrev Button = "channel_prev"
	ButtonChannelNext Button = "channel_next"

	ButtonUp    Button = "up"
	ButtonDown  Button = "down"
	ButtonLeft  Button = "left"
	ButtonRight Button = "right"
	ButtonZoom  Button = "zoom"

	ButtonScrub Button = "scrub"

	ButtonReverse     Button = "reverse"
	ButtonFastForward Button = "fast_forward"
	ButtonStop        Button = "stop"
	ButtonPlay        Button = "play"
	ButtonRec         Button = "rec"

	ButtonFlip Button = "flip"

	ButtonStatusOff   = 0
	ButtonStatusOn    = 127
	ButtonStatusBlink = 1

	FaderButtonPositionSelect  = 24
	FaderButtonPositionMute    = 16
	FaderButtonPositionSolo    = 8
	FaderButtonPositionRec     = 0
	FaderButtonPositionEncoder = 32
)

var buttonToNote map[Button]byte = map[Button]byte{
	ButtonTrack:       40,
	ButtonPan:         42,
	ButtonEQ:          44,
	ButtonSend:        41,
	ButtonPlugin:      43,
	ButtonInst:        45,
	ButtonName:        52,
	ButtonTimecode:    53,
	ButtonGlobalView:  51,
	ButtonMidiTracks:  62,
	ButtonInputs:      63,
	ButtonAudioTracks: 64,
	ButtonAudioInst:   65,
	ButtonAux:         66,
	ButtonBuses:       67,
	ButtonOutputs:     68,
	ButtonUser:        69,
	ButtonF1:          54,
	ButtonF2:          55,
	ButtonF3:          56,
	ButtonF4:          57,
	ButtonF5:          58,
	ButtonF6:          59,
	ButtonF7:          60,
	ButtonF8:          61,
	ButtonShift:       70,
	ButtonOption:      71,
	ButtonControl:     72,
	ButtonAlt:         73,
	ButtonReadOff:     74,
	ButtonWrite:       75,
	ButtonTrim:        76,
	ButtonTouch:       77,
	ButtonLatch:       78,
	ButtonGroup:       79,
	ButtonSave:        80,
	ButtonUndo:        81,
	ButtonCancel:      82,
	ButtonEnter:       83,
	ButtonMarker:      84,
	ButtonNudge:       85,
	ButtonCycle:       86,
	ButtonDrop:        87,
	ButtonReplace:     88,
	ButtonClick:       89,
	ButtonSolo:        90,
	ButtonFaderPrev:   46,
	ButtonFaderNext:   47,
	ButtonChannelPrev: 48,
	ButtonChannelNext: 49,
	ButtonUp:          96,
	ButtonDown:        97,
	ButtonLeft:        98,
	ButtonRight:       99,
	ButtonZoom:        100,
	ButtonScrub:       101,
	ButtonReverse:     91,
	ButtonFastForward: 92,
	ButtonStop:        93,
	ButtonPlay:        94,
	ButtonRec:         95,
	ButtonFlip:        50,
}

var noteToButton map[byte]Button

func init() {
	noteToButton = make(map[byte]Button, len(buttonToNote))
	for button, note := range buttonToNote {
		noteToButton[note] = button
	}
}

func (s *Server) SetFaderButtonStatus(ctx context.Context, fader int, pos FaderButtonPosition, status ButtonStatus) error {
	err := s.setRawButtonStatus(ctx, byte(pos)+byte(fader), status)
	if err != nil {
		return errors.Wrap(err, "fail to send faderButtonStatus")
	}
	return nil
}

func (s *Server) SetButtonStatus(ctx context.Context, b Button, status ButtonStatus) error {
	err := s.setRawButtonStatus(ctx, buttonToNote[b], status)
	if err != nil {
		return errors.Wrap(err, "fail to send button status")
	}
	return nil
}

func (s *Server) setRawButtonStatus(ctx context.Context, button byte, status ButtonStatus) error {
	midiMessage := MidiMessage{
		Type:       MidiMessageTypeNoteOn,
		NoteNumber: button,
		Velocity:   byte(status),
	}

	err := s.SendMidiPacket(ctx, midiMessage)
	if err != nil {
		return errors.Wrap(err, "fail to send midi message")
	}
	return nil
}
