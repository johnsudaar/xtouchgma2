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

	executorBankSize := l.XTouches.executorEndOffset() - l.XTouches.executorStartOffset()
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

	posInBlock := 0
	// First let's assign the faders
	for i := 0; i < executorBankSize; i++ {
		if i%5 == 0 {
			posInBlock = 0
		}
		log := log.WithFields(logrus.Fields{
			"fader": i,
		})
		// Find the executor in our xtouch configuration

		// The GrandMA groups the executors in group of 5.
		// Fetch the correct executor
		executor := playbacks[0].Items[i/5][posInBlock]

		// Set the fader position
		for pos, block := range executor.ExecutorBlocks {
			found, xt, offset := l.XTouches.findExecutor(i + pos)
			// If the executor has not been found in our xtouch configuration
			if !found {
				continue
			}

			f := block.Fader
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
			// Set the fader text
			line1 := executor.TextTop.Text
			line2 := ""
			if pos == 0 {
				line2 = executor.TextTop.Text
				if len(executor.Cues.Items) == 3 {
					line2 = executor.Cues.Items[1].Text
				} else if len(executor.Cues.Items) >= 1 {
					line2 = executor.Cues.Items[0].Text
				}
			} else {
				line2 = block.Fader.TypeText
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
		// if there is more than one executor block on this executor, increment i
		if len(executor.ExecutorBlocks) > 0 {
			i += len(executor.ExecutorBlocks) - 1
		}
		posInBlock++

	}

	// Next work on the rotary encoder offsets
	l.encoderLock.Lock()
	defer l.encoderLock.Unlock()
	posInBlock = 0
	for i := 0; i < executorBankSize; i++ {
		if i%5 == 0 {
			posInBlock = 0
		}

		executor := playbacks[1].Items[i/5][posInBlock]
		for pos, block := range executor.ExecutorBlocks {
			// Find the executor in our xtouch configuration
			found, xt, offset := l.XTouches.findExecutor(i + pos)
			// If the executor has not been found in our xtouch configuration
			if !found {
				continue
			}

			// If we're on the XTouch we return 9 executor but there's only 8 rotary encoder so skip this one
			if offset == 8 {
				continue
			}

			f := block.Fader
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
		if len(executor.ExecutorBlocks) > 0 {
			i += len(executor.ExecutorBlocks) - 1
		}
		posInBlock++
	}

	// Finally set the button statuses
	posInBlock = 0
	for i := 0; i < executorBankSize; i++ {
		if i%5 == 0 {
			posInBlock = 0
		}

		executor := playbacks[2].Items[i/5][posInBlock]
		for pos := 0; pos < executor.CombinedItems; pos++ {
			// Find the executor in our xtouch configuration
			found, xt, offset := l.XTouches.findExecutor(i + pos)
			// If the executor has not been found in our xtouch configuration
			if !found {
				continue
			}

			// If we're on the XTouch we return 9 executor but there's only 8 buttons so skip this one
			if offset == 8 {
				continue
			}

			var value xtouch.ButtonStatus = xtouch.ButtonStatusOff
			if executor.IsRun != 0 {
				value = xtouch.ButtonStatusOn
			}
			err := xt.server.SetFaderButtonStatus(ctx, offset, xtouch.FaderButtonPositionRec, value)
			if err != nil {
				return errors.Wrap(err, "fail to change button status")
			}
		}
		if executor.CombinedItems > 0 {
			i += executor.CombinedItems - 1
		}
		posInBlock++
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
