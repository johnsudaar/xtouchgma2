package link

import (
	"context"
	"strings"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (l *Link) faderGmaToXtouch(ctx context.Context) error {
	log := logger.Get(ctx)
	l.faderLock.Lock()
	page := l.faderPage
	l.faderLock.Unlock()

	executorBankSize := l.XTouches.executorEndOffset() - l.XTouches.executorEndOffset()
	executorStartOffset := l.XTouches.executorStartOffset()

	// The GrandMA groups executor by groups of 5 so we'll keep those boundaries
	executorStartOffset -= executorStartOffset % 5
	if executorBankSize%5 != 0 {
		executorBankSize += 5 - executorBankSize%5
	}

	playbacks, err := l.GMA.Playbacks(page, []gma2ws.PlaybacksRange{
		gma2ws.PlaybacksRange{
			Index:    FadersStartOffset + executorStartOffset,
			Count:    executorBankSize,
			ItemType: gma2ws.PlaybacksItemTypeFader,
		},
		gma2ws.PlaybacksRange{
			Index:    RotaryEncoderStartOffset + executorStartOffset,
			Count:    executorBankSize,
			ItemType: gma2ws.PlaybacksItemTypeFader,
		},
		gma2ws.PlaybacksRange{
			Index:    ButtonsStartOffset + executorStartOffset,
			Count:    executorBankSize,
			ItemType: gma2ws.PlaybacksItemTypeButton,
		},
	})
	if err != nil {
		return errors.Wrap(err, "fail to sync GMA faders")
	}

	// First let's assign the faders
	for i := 0; i < executorBankSize; i++ {
		log := log.WithFields(logrus.Fields{
			"fader": i,
		})
		// Find the executor in our xtouch configuration
		found, xt, offset := l.XTouches.findExecutor(i)
		// If the executor has not been found in our xtouch configuration
		if !found {
			continue
		}

		// The GrandMA groups the executors in group of 5.
		// Fetch the correct executor
		executor := playbacks[0].Items[i/5][i%5]

		// Set the fader position
		f := executor.ExecutorBlocks[0].Fader
		var value float64
		if f.Max != 0 {
			value = float64(f.Value) / float64(f.Max-f.Min)
		}
		err := xt.server.SetFaderPos(ctx, offset, value)
		if err != nil {
			log.WithError(err).Error("fail to send fader to its position")
		}

		// If we're on the last fader of the XTouch
		if offset == 8 {
			// Do not try to update the text
			continue
		}

		// Set the fader textr
		line1 := executor.TextTop.Text
		line2 := ""
		if len(executor.Cues.Items) == 3 {
			line2 = executor.Cues.Items[1].Text
		} else if len(executor.Cues.Items) >= 1 {
			line2 = executor.Cues.Items[0].Text
		}
		color, err := xtouch.ClosestScribbleColor(f.BorderColor)
		if err != nil {
			log.WithError(err).Error("fail to get scribble color")
			color = xtouch.ScribbleColorWhite
		}

		err = xt.server.SetScribble(ctx, offset, color, true, strings.TrimSpace(line1), strings.TrimSpace(line2))
		if err != nil {
			log.WithError(err).Error("fail to send scribble data")
		}

		// Set the button light
		var buttonStatus xtouch.ButtonStatus = xtouch.ButtonStatusOff
		if executor.IsRun != 0 {
			buttonStatus = xtouch.ButtonStatusOn
		}
		err = xt.server.SetFaderButtonStatus(ctx, offset, xtouch.FaderButtonPositionSelect, buttonStatus)
		if err != nil {
			return errors.Wrap(err, "fail to change button status")
		}
	}

	// Next work on the rotary encoder offsets
	l.encoderLock.Lock()
	defer l.encoderLock.Unlock()
	for i := 0; i < executorBankSize; i++ {
		// Find the executor in our xtouch configuration
		found, xt, offset := l.XTouches.findExecutor(i)
		// If the executor has not been found in our xtouch configuration
		if !found {
			continue
		}

		// If we're on the XTouch we return 9 executor but there's only 8 rotary encoder so skip this one
		if offset == 8 {
			continue
		}

		executor := playbacks[1].Items[i/5][i%5]
		f := executor.ExecutorBlocks[0].Fader
		var value float64
		if f.Max != 0 {
			value = float64(f.Value) / float64(f.Max-f.Min)
		}
		l.encoderGMAValue[i] = value
		// If we are using the encoders in attribute mode and we're on the xtouch do not copy the encoder status to the rotary encoders
		if l.encoderAsAttributes && xt.xtouchType == xtouch.ServerTypeXTouch {
			continue
		}
		// Set the xtouch value
		xt.server.SetRingPosition(ctx, offset, value)
	}

	// Finally set the button statuses
	for i := 0; i < 8; i++ {
		// Find the executor in our xtouch configuration
		found, xt, offset := l.XTouches.findExecutor(i)
		// If the executor has not been found in our xtouch configuration
		if !found {
			continue
		}

		// If we're on the XTouch we return 9 executor but there's only 8 buttons so skip this one
		if offset == 8 {
			continue
		}

		executor := playbacks[2].Items[i/5][i%5]
		var value xtouch.ButtonStatus = xtouch.ButtonStatusOff
		if executor.IsRun != 0 {
			value = xtouch.ButtonStatusOn
		}
		err := xt.server.SetFaderButtonStatus(ctx, offset, xtouch.FaderButtonPositionRec, value)
		if err != nil {
			return errors.Wrap(err, "fail to change button status")
		}
	}

	// Set the page display
	// Find the main XTouch
	mainXtouch := l.XTouches.XTouch()
	if mainXtouch == nil {
		// if there isn't any => Exit
		return nil
	}

	mainXtouch.SetAssignement(ctx, page+1)
	return nil
}

func (l *Link) onFaderChangeEvent(ctx context.Context, executor int, position float64) {
	l.faderLock.Lock()
	page := l.faderPage
	l.faderLock.Unlock()

	log := logger.Get(ctx)
	err := l.GMA.FaderChanged(ctx, executor, page, position)
	if err != nil {
		log.WithError(err).Error("fail to send fader position")
	}
}

func (l *Link) FaderPageUp() {
	l.faderLock.Lock()
	defer l.faderLock.Unlock()
	l.faderPage++
}

func (l *Link) FaderPageDown() {
	l.faderLock.Lock()
	defer l.faderLock.Unlock()
	l.faderPage--
	if l.faderPage < 0 {
		l.faderPage = 0
	}
}
