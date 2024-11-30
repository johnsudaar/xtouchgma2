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

type ScribbleValue struct {
	Color              xtouch.ScribbleColor
	SecondLineInverted bool
	Line1              string
	Line2              string
}

type XTouch struct {
	Server         *xtouch.Server
	xtouchType     xtouch.ServerType
	executorOffset int
	link           *Link
	Faders         []XTouchFader
	Buttons        []*Refresher[xtouch.ButtonStatus]
	RotaryEncoder  []*Refresher[float64]
	Assignment     *Refresher[int]
}

type XTouches []*XTouch

type XTouchFader struct {
	Fader    *Refresher[float64]
	Scribble *Refresher[ScribbleValue]
	ButtonA  *Refresher[xtouch.ButtonStatus]
	ButtonB  *Refresher[xtouch.ButtonStatus]
	ButtonC  *Refresher[xtouch.ButtonStatus]
}

func NewXTouch(server *xtouch.Server, xtouchType xtouch.ServerType, executorOffset int, link *Link, params RefreshParams) *XTouch {
	xt := XTouch{
		Server:         server,
		xtouchType:     xtouchType,
		executorOffset: executorOffset,
		link:           link,
	}

	xt.Faders = make([]XTouchFader, xt.size())
	for i := 0; i < xt.size(); i++ {
		fader := XTouchFader{
			Fader:   NewRefresher("fader", params, xt.faderRefreshFunc(i)),
			ButtonA: NewRefresher("button_select", params, xt.buttonRefreshFunc(i, xtouch.FaderButtonPositionSelect)),
		}

		if i != 8 { // The main fader doesn't have a button
			fader.Scribble = NewRefresher("scribble", params, xt.scribbleRefreshFunc(i))
			fader.ButtonB = NewRefresher("button_mute", params, xt.buttonRefreshFunc(i, xtouch.FaderButtonPositionMute))
			fader.ButtonC = NewRefresher("button_solo", params, xt.buttonRefreshFunc(i, xtouch.FaderButtonPositionSolo))
		}
		xt.Faders[i] = fader
	}

	xt.Buttons = make([]*Refresher[xtouch.ButtonStatus], 8)
	xt.RotaryEncoder = make([]*Refresher[float64], 8)
	for i := 0; i < 8; i++ {
		xt.Buttons[i] = NewRefresher("button", params, xt.buttonRefreshFunc(i, xtouch.FaderButtonPositionRec))
		xt.RotaryEncoder[i] = NewRefresher("rotary_encoder", params, xt.rotaryRingRefreshFunc(i))
	}

	if xtouchType == xtouch.ServerTypeXTouch {
		xt.Assignment = NewRefresher("assignment", params, func(ctx context.Context, value int) error {
			return xt.Server.SetAssignement(ctx, value)
		})
	}

	return &xt
}

func (x *XTouch) faderRefreshFunc(i int) func(context.Context, float64) error {
	return func(ctx context.Context, value float64) error {
		return x.Server.SetFaderPos(ctx, i, value)
	}
}

func (x *XTouch) buttonRefreshFunc(i int, button xtouch.FaderButtonPosition) func(context.Context, xtouch.ButtonStatus) error {
	return func(ctx context.Context, value xtouch.ButtonStatus) error {
		return x.Server.SetFaderButtonStatus(ctx, i, button, value)
	}
}

func (x *XTouch) scribbleRefreshFunc(i int) func(context.Context, ScribbleValue) error {
	return func(ctx context.Context, value ScribbleValue) error {
		return x.Server.SetScribble(ctx, i, value.Color, value.SecondLineInverted, value.Line1, value.Line2)
	}
}

func (x *XTouch) rotaryRingRefreshFunc(i int) func(context.Context, float64) error {
	return func(ctx context.Context, value float64) error {
		return x.Server.SetRingPosition(ctx, i, value)
	}
}

func (x *XTouch) RunRefreshers(ctx context.Context) {
	for _, fader := range x.Faders {
		fader.Fader.Refresh(ctx)
		fader.ButtonA.Refresh(ctx)
		if fader.Scribble != nil {
			fader.Scribble.Refresh(ctx)
			fader.ButtonB.Refresh(ctx)
			fader.ButtonC.Refresh(ctx)
		}
	}
	for i := 0; i < 8; i++ {
		x.Buttons[i].Refresh(ctx)
		x.RotaryEncoder[i].Refresh(ctx)
	}

	if x.Assignment != nil {
		x.Assignment.Refresh(ctx)
	}
}

func (x *XTouch) ForceRefresh() {
	for _, fader := range x.Faders {
		fader.Fader.ForceRefresh()
		fader.ButtonA.ForceRefresh()
		if fader.Scribble != nil {
			fader.Scribble.ForceRefresh()
			fader.ButtonB.ForceRefresh()
			fader.ButtonC.ForceRefresh()
		}
	}
	for i := 0; i < 8; i++ {
		x.Buttons[i].ForceRefresh()
		x.RotaryEncoder[i].ForceRefresh()
	}

	if x.Assignment != nil {
		x.Assignment.ForceRefresh()
	}
}

// How many executor is there on this device ?
func (x *XTouch) size() int {
	if x.xtouchType == xtouch.ServerTypeXTouch {
		return 9
	}
	return 8
}

func (x XTouch) subscribeToEventChanges() {
	x.Server.SubscribeToFaderChanges(x.onFaderChange)
	x.Server.SubscribeButtonChanges(x.onButtonChange)
	x.Server.SubscribeEncoderChanges(x.onEncoderChange)
}

func (x *XTouch) onFaderChange(ctx context.Context, e xtouch.FaderChangedEvent) {
	// Inhibit the button
	x.Faders[e.Fader].Fader.Inhibit()

	executor := FadersStartOffset + e.Fader + x.executorOffset
	x.link.onFaderChangeEvent(ctx, executor, e.Position())
}

func (x *XTouch) onButtonChange(ctx context.Context, e xtouch.ButtonChangedEvent) {
	// Translate the executor offset to the global GMA offset
	executor := ButtonsStartOffset + e.Executor + x.executorOffset
	// If we pressed a button linked to a fader
	if e.Type == xtouch.ButtonTypeSelect ||
		e.Type == xtouch.ButtonTypeMute ||
		e.Type == xtouch.ButtonTypeSolo {
		executor = FadersStartOffset + e.Executor + x.executorOffset
	}
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

func (x *XTouch) onEncoderChange(ctx context.Context, e xtouch.EncoderChangedEvent) {
	encoder := RotaryEncoderStartOffset + int(e.Encoder) + x.executorOffset
	x.link.onEncoderChangedEvent(ctx, e, x.xtouchType, encoder)
}

// Return the main XTouch if there's one, nil istead
func (x XTouches) XTouch() *XTouch {
	for _, xt := range x {
		if xt.xtouchType == xtouch.ServerTypeXTouch {
			return xt
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

func (x XTouches) findFader(offset int) (XTouchFader, bool) {
	for _, xt := range x {
		if offset >= xt.executorOffset && offset < xt.executorOffset+xt.size() {
			return xt.Faders[offset-xt.executorOffset], true
		}
	}
	return XTouchFader{}, false
}

func (x XTouches) RotaryEncoder(offset int) (*Refresher[float64], *XTouch, bool) {
	for _, xt := range x {
		if offset >= xt.executorOffset && offset < xt.executorOffset+8 {
			return xt.RotaryEncoder[offset-xt.executorOffset], xt, true
		}
	}
	return nil, nil, false
}

func (x XTouches) Button(offset int) (*Refresher[xtouch.ButtonStatus], bool) {
	for _, xt := range x {
		if offset >= xt.executorOffset && offset < xt.executorOffset+8 {
			return xt.Buttons[offset-xt.executorOffset], true
		}
	}
	return nil, false
}

func (x XTouches) ForceRefresh() {
	for _, xt := range x {
		xt.ForceRefresh()
	}
}
