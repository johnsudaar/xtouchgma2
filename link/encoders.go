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

func (l *Link) onEncoderChangedEvent(ctx context.Context, e xtouch.EncoderChangedEvent) {
	if e.Delta == 0 {
		return
	}
	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	l.encoderLock.RUnlock()

	if e.Encoder == xtouch.MainEncoder || encoderAsAttributes {
		l.updateEncoderAttribute(ctx, e)
	} else {
		l.updateEncoderFader(ctx, e)
	}

}

func (l *Link) updateEncoderRings(ctx context.Context) error {
	l.encoderLock.RLock()
	encoderAsAttributes := l.encoderAsAttributes
	encoderAttributes := l.encoderAttributes
	encoderAttributesCoeff := l.encoderAttributesCoeff
	l.encoderLock.RUnlock()

	if !encoderAsAttributes {
		return nil
	}

	for i := 0; i < 8; i++ {
		if encoderAttributes[i] == "" {
			l.XTouch.SetRingPosition(ctx, i, 0)
		} else {
			l.XTouch.SetRingPosition(ctx, i, encodersCoeff[encoderAttributesCoeff[i]]/encodersCoeffMax)
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

func (l *Link) updateEncoderFader(ctx context.Context, e xtouch.EncoderChangedEvent) {
	l.faderLock.Lock()
	page := l.faderPage
	l.faderLock.Unlock()

	l.encoderLock.Lock()
	value := l.encoderGMAValue[e.Encoder]
	value += float64(e.Delta) * 0.01
	l.encoderGMAValue[e.Encoder] = value
	defer l.encoderLock.Unlock()

	l.GMA.FaderChanged(ctx, int(e.Encoder)+15, page, value)
}
