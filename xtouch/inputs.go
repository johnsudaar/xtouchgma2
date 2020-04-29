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

	ButtonSelect1 Button = "select1"
	ButtonSelect2 Button = "select2"
	ButtonSelect3 Button = "select3"
	ButtonSelect4 Button = "select4"
	ButtonSelect5 Button = "select5"
	ButtonSelect6 Button = "select6"
	ButtonSelect7 Button = "select7"
	ButtonSelect8 Button = "select8"

	ButtonMute1 Button = "mute1"
	ButtonMute2 Button = "mute2"
	ButtonMute3 Button = "mute3"
	ButtonMute4 Button = "mute4"
	ButtonMute5 Button = "mute5"
	ButtonMute6 Button = "mute6"
	ButtonMute7 Button = "mute7"
	ButtonMute8 Button = "mute8"

	ButtonSolo1 Button = "solo1"
	ButtonSolo2 Button = "solo2"
	ButtonSolo3 Button = "solo3"
	ButtonSolo4 Button = "solo4"
	ButtonSolo5 Button = "solo5"
	ButtonSolo6 Button = "solo6"
	ButtonSolo7 Button = "solo7"
	ButtonSolo8 Button = "solo8"

	ButtonRec1 Button = "rec1"
	ButtonRec2 Button = "rec2"
	ButtonRec3 Button = "rec3"
	ButtonRec4 Button = "rec4"
	ButtonRec5 Button = "rec5"
	ButtonRec6 Button = "rec6"
	ButtonRec7 Button = "rec7"
	ButtonRec8 Button = "rec8"

	ButtonRotary1 Button = "rotary1"
	ButtonRotary2 Button = "rotary2"
	ButtonRotary3 Button = "rotary3"
	ButtonRotary4 Button = "rotary4"
	ButtonRotary5 Button = "rotary5"
	ButtonRotary6 Button = "rotary6"
	ButtonRotary7 Button = "rotary7"
	ButtonRotary8 Button = "rotary8"

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

	ButtonSelect1: 24,
	ButtonSelect2: 25,
	ButtonSelect3: 26,
	ButtonSelect4: 27,
	ButtonSelect5: 28,
	ButtonSelect6: 29,
	ButtonSelect7: 30,
	ButtonSelect8: 31,

	ButtonMute1: 16,
	ButtonMute2: 17,
	ButtonMute3: 18,
	ButtonMute4: 19,
	ButtonMute5: 20,
	ButtonMute6: 21,
	ButtonMute7: 22,
	ButtonMute8: 23,

	ButtonSolo1: 8,
	ButtonSolo2: 9,
	ButtonSolo3: 10,
	ButtonSolo4: 11,
	ButtonSolo5: 12,
	ButtonSolo6: 13,
	ButtonSolo7: 14,
	ButtonSolo8: 15,

	ButtonRec1: 0,
	ButtonRec2: 1,
	ButtonRec3: 2,
	ButtonRec4: 3,
	ButtonRec5: 4,
	ButtonRec6: 5,
	ButtonRec7: 6,
	ButtonRec8: 7,

	ButtonRotary1: 32,
	ButtonRotary2: 33,
	ButtonRotary3: 34,
	ButtonRotary4: 35,
	ButtonRotary5: 36,
	ButtonRotary6: 37,
	ButtonRotary7: 38,
	ButtonRotary8: 39,
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
