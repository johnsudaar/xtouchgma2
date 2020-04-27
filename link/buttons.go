package link

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

var buttonMap map[xtouch.Button]int = map[xtouch.Button]int{
	xtouch.ButtonTrack:       1,
	xtouch.ButtonPan:         2,
	xtouch.ButtonEQ:          3,
	xtouch.ButtonSend:        4,
	xtouch.ButtonPlugin:      5,
	xtouch.ButtonInst:        6,
	xtouch.ButtonName:        7,
	xtouch.ButtonTimecode:    8,
	xtouch.ButtonGlobalView:  9,
	xtouch.ButtonMidiTracks:  10,
	xtouch.ButtonInputs:      11,
	xtouch.ButtonAudioTracks: 12,
	xtouch.ButtonAudioInst:   13,
	xtouch.ButtonAux:         14,
	xtouch.ButtonBuses:       15,
	xtouch.ButtonOutputs:     16,
	xtouch.ButtonUser:        17,
	xtouch.ButtonF1:          18,
	xtouch.ButtonF2:          19,
	xtouch.ButtonF3:          20,
	xtouch.ButtonF4:          21,
	xtouch.ButtonF5:          22,
	xtouch.ButtonF6:          23,
	xtouch.ButtonF7:          24,
	xtouch.ButtonF8:          25,
	xtouch.ButtonShift:       26,
	xtouch.ButtonOption:      27,
	xtouch.ButtonReadOff:     28,
	xtouch.ButtonWrite:       29,
	xtouch.ButtonTrim:        30,
	xtouch.ButtonSave:        31,
	xtouch.ButtonUndo:        32,
	xtouch.ButtonGroup:       33,
	xtouch.ButtonCancel:      34,
	xtouch.ButtonEnter:       35,
	xtouch.ButtonReplace:     36,
	xtouch.ButtonClick:       37,
	xtouch.ButtonSolo:        38,
	xtouch.ButtonControl:     39,
	xtouch.ButtonAlt:         40,
	xtouch.ButtonTouch:       41,
	xtouch.ButtonLatch:       42,
	xtouch.ButtonMarker:      43,
	xtouch.ButtonNudge:       44,
	xtouch.ButtonCycle:       45,
	xtouch.ButtonDrop:        46,
	xtouch.ButtonScrub:       47,
	xtouch.ButtonUp:          48,
	xtouch.ButtonRight:       49,
	xtouch.ButtonDown:        50,
	xtouch.ButtonLeft:        51,
	xtouch.ButtonZoom:        52,
}

func (l *Link) onButtonChange(ctx context.Context, event xtouch.ButtonChangedEvent) {
	log := logger.Get(ctx)
	address, ok := buttonMap[event.Button]
	if ok {
		var value byte = 0
		if event.Status == xtouch.ButtonStatusOn {
			value = 255
		}
		log.Info("sending", address, value)

		l.SetDMXValue(address, value)
	}
}
