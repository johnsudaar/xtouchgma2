package link

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
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
	if event.Type == xtouch.ButtonTypeCommand {
		address, ok := buttonMap[event.Button]
		if ok {
			var value byte = 0
			if event.Status == xtouch.ButtonStatusOn {
				value = 255
			}
			l.SetDMXValue(address, value)
			return
		}

		if event.Button == xtouch.ButtonFlip {
			l.faderLock.Lock()
			page := l.faderPage
			l.faderLock.Unlock()
			l.GMA.ButtonChanged(ctx, 8, page, 0, event.Status == xtouch.ButtonStatusOn)
			return
		}

		if event.Button == xtouch.ButtonFaderNext && event.Status == xtouch.ButtonStatusOn {
			l.FaderPageUp()
			return
		}
		if event.Button == xtouch.ButtonFaderPrev && event.Status == xtouch.ButtonStatusOn {
			l.FaderPageDown()
			return
		}

		if event.Button == xtouch.ButtonChannelNext && event.Status == xtouch.ButtonStatusOn {
			l.encoderLock.Lock()
			l.encoderAsAttributes = true
			l.encoderLock.Unlock()
			return
		}

		if event.Button == xtouch.ButtonChannelPrev && event.Status == xtouch.ButtonStatusOn {
			l.encoderLock.Lock()
			l.encoderAsAttributes = false
			l.encoderLock.Unlock()
			return
		}
	}

	if event.Type == xtouch.ButtonTypeSelect ||
		event.Type == xtouch.ButtonTypeMute ||
		event.Type == xtouch.ButtonTypeSolo ||
		event.Type == xtouch.ButtonTypeRec {
		l.faderLock.Lock()
		page := l.faderPage
		l.faderLock.Unlock()

		executor := event.Executor
		buttonID := 0
		if event.Type == xtouch.ButtonTypeMute {
			buttonID = 1
		} else if event.Type == xtouch.ButtonTypeSolo {
			buttonID = 2
		}

		if event.Type == xtouch.ButtonTypeRec {
			executor += 100
		}

		l.GMA.ButtonChanged(ctx, executor, page, buttonID, event.Status == xtouch.ButtonStatusOn)
	}

	if event.Type == xtouch.ButtonTypeRotary {
		l.encoderLock.RLock()
		encoderAsAttributes := l.encoderAsAttributes
		l.encoderLock.RUnlock()

		if encoderAsAttributes {
			if event.Status == xtouch.ButtonStatusOff {
				return
			}
			l.encoderLock.Lock()
			l.encoderAttributesCoeff[event.Executor] = (l.encoderAttributesCoeff[event.Executor] + 1) % len(encodersCoeff)
			l.encoderLock.Unlock()
		} else {
			l.faderLock.Lock()
			page := l.faderPage
			l.faderLock.Unlock()
			l.GMA.ButtonChanged(ctx, 15+event.Executor, page, 0, event.Status == xtouch.ButtonStatusOn)
		}
	}
}

func (l *Link) updateButtons(ctx context.Context) error {
	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	l.encoderLock.RUnlock()

	if encoderAsAttributes {
		err := l.XTouch.SetButtonStatus(ctx, xtouch.ButtonChannelPrev, xtouch.ButtonStatusOff)
		if err != nil {
			return errors.Wrap(err, "fail to set channel prev status")
		}
		err = l.XTouch.SetButtonStatus(ctx, xtouch.ButtonChannelNext, xtouch.ButtonStatusOn)
		if err != nil {
			return errors.Wrap(err, "fail to set channel next status")
		}
	} else {
		err := l.XTouch.SetButtonStatus(ctx, xtouch.ButtonChannelPrev, xtouch.ButtonStatusOn)
		if err != nil {
			return errors.Wrap(err, "fail to set channel prev status")
		}
		err = l.XTouch.SetButtonStatus(ctx, xtouch.ButtonChannelNext, xtouch.ButtonStatusOff)
		if err != nil {
			return errors.Wrap(err, "fail to set channel next status")
		}
	}

	res, err := l.GMA.KeyStatuses("set", "edit", "clear", "solo", "high", "align")
	if err != nil {
		return errors.Wrap(err, "fail to get gma key statuses")
	}

	for key, status := range res {
		var button xtouch.Button
		switch key {
		case "set":
			button = xtouch.ButtonZoom
		case "edit":
			button = xtouch.ButtonF3
		case "clear":
			button = xtouch.ButtonF5
		case "solo":
			button = xtouch.ButtonPan
		case "high":
			button = xtouch.ButtonTrack
		case "align":
			button = xtouch.ButtonInst
		}
		err := l.setKeyStatus(ctx, button, status)
		if err != nil {
			return errors.Wrap(err, "fail to update button")
		}
	}

	return nil
}

func (l *Link) setKeyStatus(ctx context.Context, key xtouch.Button, status gma2ws.KeyStatus) error {
	var bStatus xtouch.ButtonStatus = xtouch.ButtonStatusOff
	if status == gma2ws.KeyStatusOn {
		bStatus = xtouch.ButtonStatusOn
	}
	if status == gma2ws.KeyStatusBlink {
		bStatus = xtouch.ButtonStatusBlink
	}

	err := l.XTouch.SetButtonStatus(ctx, key, bStatus)
	if err != nil {
		return errors.Wrap(err, "fail to send button status")
	}
	return nil
}
