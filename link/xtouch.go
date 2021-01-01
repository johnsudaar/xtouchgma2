package link

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/xtouch"
)

const (
	FadersStartOffset        = 0
	RotaryEncoderStartOffset = 50
	ButtonsStartOffset       = 100
)

type XTouch struct {
	server         *xtouch.Server
	xtouchType     xtouch.ServerType
	executorOffset int
	link           *Link
}

// How many executor is there on this device ?
func (x XTouch) size() int {
	if x.xtouchType == xtouch.ServerTypeXTouch {
		return 9
	}
	return 8
}

func (x XTouch) subscribeToEventChanges() {
	x.server.SubscribeToFaderChanges(x.onFaderChange)
	x.server.SubscribeButtonChanges(x.onButtonChange)
	x.server.SubscribeEncoderChanges(x.onEncoderChange)
}

func (x XTouch) onFaderChange(ctx context.Context, e xtouch.FaderChangedEvent) {
	executor := FadersStartOffset + e.Fader + x.executorOffset
	x.link.onFaderChangeEvent(ctx, executor, e.Position())
}

func (x XTouch) onButtonChange(ctx context.Context, e xtouch.ButtonChangedEvent) {
	// Translate the executor offset to the global GMA offset
	executor := ButtonsStartOffset + e.Executor + x.executorOffset
	// If we pressed a retorary encoder modify the global GMA offset
	if e.Type == xtouch.ButtonTypeRotary {
		executor = RotaryEncoderStartOffset + e.Executor + x.executorOffset
	}

	// If the button pressed was the flip button, consider it like the first button of the 9th fader.
	if e.Button == xtouch.ButtonFlip {
		executor = FadersStartOffset + x.executorOffset + 8
	}

	x.link.onButtonChange(ctx, e, executor, x.xtouchType)
}

func (x XTouch) onEncoderChange(ctx context.Context, e xtouch.EncoderChangedEvent) {
	encoder := RotaryEncoderStartOffset + int(e.Encoder) + x.executorOffset
	x.link.onEncoderChangedEvent(ctx, e, x.xtouchType, encoder)
}

type XTouches []XTouch

// Return the main XTouch if there's one, nil istead
func (x XTouches) XTouch() *xtouch.Server {
	for _, xt := range x {
		if xt.xtouchType == xtouch.ServerTypeXTouch {
			return xt.server
		}
	}
	return nil
}

// Minimum offset to fetch from GrandMA
func (x XTouches) executorStartOffset() int {
	if len(x) == 0 {
		return -1
	}
	min := x[0].executorOffset
	// Find the smallest executorOffset
	for _, xt := range x {
		if xt.executorOffset < min {
			min = xt.executorOffset
		}
	}
	return min
}

// Maximum offset to fetch from GrandMA
func (x XTouches) executorEndOffset() int {
	if len(x) == 0 {
		return -1
	}

	max := x[0].executorOffset + x[0].size()
	for _, xt := range x {
		val := xt.executorOffset + xt.size()
		if val > max {
			max = val
		}
	}
	return max
}

func (x XTouches) findExecutor(offset int) (bool, XTouch, int) {
	for _, xt := range x {
		if offset >= xt.executorOffset && offset < xt.executorOffset+xt.size() {
			return true, xt, offset - xt.executorOffset
		}
	}
	return false, XTouch{}, 0
}
