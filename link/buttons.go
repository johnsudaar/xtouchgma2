package link

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
)

// Map xtouch command buttons to their DMX addresses
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

// handle button changes on xtouches
func (l *Link) onButtonChange(ctx context.Context, event xtouch.ButtonChangedEvent, executor int, xtouchType xtouch.ServerType) {
	// If the button comes from the command section of the xtouch
	if event.Type == xtouch.ButtonTypeCommand {
		// Map the button to the DMX address
		address, ok := buttonMap[event.Button]
		if ok {
			var value byte = 0
			if event.Status == xtouch.ButtonStatusOn {
				value = 255
			}
			l.SetDMXValue(address, value)
			return
		}

		// if the button is the flip button consider it as the first button from the 9th executor.
		if event.Button == xtouch.ButtonFlip {
			l.faderLock.Lock()
			page := l.faderPage
			l.faderLock.Unlock()
			l.GMA.ButtonChanged(ctx, executor, page, 0, event.Status == xtouch.ButtonStatusOn)
			return
		}

		// Handle the page up and page down buttons.
		if event.Button == xtouch.ButtonFaderNext && event.Status == xtouch.ButtonStatusOn {
			l.FaderPageUp()
			return
		}
		if event.Button == xtouch.ButtonFaderPrev && event.Status == xtouch.ButtonStatusOn {
			l.FaderPageDown()
			return
		}

		// Channel next and previous button are used to change the rotary encoder assignations
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

	// If it's not a custom button, it's an executor button
	if event.Type == xtouch.ButtonTypeSelect ||
		event.Type == xtouch.ButtonTypeMute ||
		event.Type == xtouch.ButtonTypeSolo ||
		event.Type == xtouch.ButtonTypeRec {
		l.faderLock.Lock()
		page := l.faderPage
		l.faderLock.Unlock()

		buttonID := 0
		if event.Type == xtouch.ButtonTypeMute {
			buttonID = 1
		} else if event.Type == xtouch.ButtonTypeSolo {
			buttonID = 2
		}

		l.GMA.ButtonChanged(ctx, executor, page, buttonID, event.Status == xtouch.ButtonStatusOn)
		return
	}

	// If it's not a custom button nor an executor it's a rotary encoder button
	if event.Type == xtouch.ButtonTypeRotary {
		l.encoderLock.RLock()
		encoderAsAttributes := l.encoderAsAttributes
		l.encoderLock.RUnlock()

		// If the it's an event from the XTouch and the XTouch is in attribute mode
		if encoderAsAttributes && xtouchType == xtouch.ServerTypeXTouch {
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
			l.GMA.ButtonChanged(ctx, executor, page, 0, event.Status == xtouch.ButtonStatusOn)
		}
	}
}

// This method will fetch lighting informations for non executor buttons
func (l *Link) updateButtons(ctx context.Context) error {
	// Do not run this if there is no xtouch connected
	mainXtouch := l.XTouches.XTouch()
	if mainXtouch == nil {
		return nil
	}

	// If there is an xtouch connected, start fetching and setting those lighting attributes
	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	l.encoderLock.RUnlock()

	// Update our internal buttons lighting (prev page and encoder status, if there is a xtouch)
	if encoderAsAttributes {
		err := mainXtouch.SetButtonStatus(ctx, xtouch.ButtonChannelPrev, xtouch.ButtonStatusOff)
		if err != nil {
			return errors.Wrap(err, "fail to set channel prev status")
		}
		err = mainXtouch.SetButtonStatus(ctx, xtouch.ButtonChannelNext, xtouch.ButtonStatusOn)
		if err != nil {
			return errors.Wrap(err, "fail to set channel next status")
		}
	} else {
		err := mainXtouch.SetButtonStatus(ctx, xtouch.ButtonChannelPrev, xtouch.ButtonStatusOn)
		if err != nil {
			return errors.Wrap(err, "fail to set channel prev status")
		}
		err = mainXtouch.SetButtonStatus(ctx, xtouch.ButtonChannelNext, xtouch.ButtonStatusOff)
		if err != nil {
			return errors.Wrap(err, "fail to set channel next status")
		}
	}

	// Fetch lighting informations about some buttons
	res, err := l.GMA.KeyStatuses("set", "edit", "clear", "solo", "high", "align")
	if err != nil {
		return errors.Wrap(err, "fail to get gma key statuses")
	}

	// Set those lighting informations on the xtouch button
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
	// Do not run this if there is no xtouch connected
	mainXtouch := l.XTouches.XTouch()
	if mainXtouch == nil {
		return nil
	}

	var bStatus xtouch.ButtonStatus = xtouch.ButtonStatusOff
	if status == gma2ws.KeyStatusOn {
		bStatus = xtouch.ButtonStatusOn
	}
	if status == gma2ws.KeyStatusBlink {
		bStatus = xtouch.ButtonStatusBlink
	}

	// Try to find which xtouch owns this encoder
	err := mainXtouch.SetButtonStatus(ctx, key, bStatus)
	if err != nil {
		return errors.Wrap(err, "fail to send button status")
	}
	return nil
}
