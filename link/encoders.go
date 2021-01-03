package link

import (
	"context"
	"fmt"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

var encodersCoeff []float64 = []float64{1, .5, .1}
var encodersCoeffMax float64 = 1

func (l *Link) SetEncoderAttributes(attributes [8]string) {
	l.encoderLock.Lock()
	defer l.encoderLock.Unlock()
	l.encoderAttributes = attributes
}

func (l *Link) UseEncoderAsAttributes(v bool) {
	l.encoderLock.Lock()
	defer l.encoderLock.Unlock()

	l.encoderAsAttributes = v
}

func (l *Link) onEncoderChangedEvent(ctx context.Context, e xtouch.EncoderChangedEvent, xtouchType xtouch.ServerType, offset int) {
	if e.Delta == 0 {
		return
	}
	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	l.encoderLock.RUnlock()

	if e.Encoder == xtouch.MainEncoder || (encoderAsAttributes && xtouchType == xtouch.ServerTypeXTouch) {
		l.updateEncoderAttribute(ctx, e)
	} else {
		l.updateEncoderFader(ctx, e, offset)
	}

}

func (l *Link) updateEncoderRings(ctx context.Context) error {
	// Try to find the main XTouch
	mainXtouch := l.XTouches.XTouch()
	if mainXtouch == nil {
		// If there is no main XTouch => No need to update the encoder ring
		return nil
	}

	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	encoderAttributes := l.encoderAttributes
	encoderAttributesCoeff := l.encoderAttributesCoeff
	l.encoderLock.RUnlock()

	// If we're not using the encoder as attributes
	if !encoderAsAttributes {
		// exit, no need to do anything there
		return nil
	}

	// Set the 8 encoder values
	for i := 0; i < 8; i++ {
		if encoderAttributes[i] == "" {
			mainXtouch.SetRingPosition(ctx, i, 0)
		} else {
			mainXtouch.SetRingPosition(ctx, i, encodersCoeff[encoderAttributesCoeff[i]]/encodersCoeffMax)
		}
	}
	return nil
}

func (l *Link) updateEncoderAttribute(ctx context.Context, e xtouch.EncoderChangedEvent) {
	l.encoderLock.RLock()
	defer l.encoderLock.RUnlock()
	log := logger.Get(ctx)
	var attribute string
	var coeff float64 = 1
	if e.Encoder == xtouch.MainEncoder {
		attribute = "dim"
	} else {

		attribute = l.encoderAttributes[e.Encoder]
		if attribute == "" {
			return
		}
		coeff = encodersCoeff[l.encoderAttributesCoeff[e.Encoder]%len(encodersCoeff)]
	}
	sign := "+"
	delta := float64(e.Delta)
	if e.Delta < 0 {
		sign = "-"
		delta *= -1
	}

	command := fmt.Sprintf("Attribute \"%s\" At %s %v", attribute, sign, delta*coeff)
	log.Debugf("Sending command: %s", command)
	err := l.GMA.SendCommand(ctx, command)
	if err != nil {
		log.WithField("cmd", command).WithError(err).Error("fail to send encoder data")
	}
}

func (l *Link) updateEncoderFader(ctx context.Context, e xtouch.EncoderChangedEvent, offset int) {
	l.faderLock.Lock()
	page := l.faderPage
	l.faderLock.Unlock()

	l.encoderLock.Lock()
	value := l.encoderGMAValue[offset]
	value += float64(e.Delta) * 0.01
	l.encoderGMAValue[offset] = value
	defer l.encoderLock.Unlock()

	l.GMA.FaderChanged(ctx, offset, page, value)
}
