package link

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

var buttonMap map[xtouch.Button]gma2ws.KeyName = map[xtouch.Button]gma2ws.KeyName{
	xtouch.ButtonTrack: gma2ws.KeyNameHighlight,
	xtouch.ButtonPan:   gma2ws.KeyNameSolo,
	xtouch.ButtonSend:  gma2ws.KeyNameSelect,

	xtouch.ButtonTrim: gma2ws.KeyName1,
}

func (l *Link) onButtonChange(ctx context.Context, event xtouch.ButtonChangedEvent) {
	log := logger.Get(ctx)
	key, ok := buttonMap[event.Button]
	if ok {
		status := gma2ws.KeyStatusReleased
		if event.Status == xtouch.ButtonStatusOn {
			status = gma2ws.KeyStatusPressed
		}

		log.Info("sending", key, status)

		err := l.GMA.SendKey(ctx, key, status)
		if err != nil {
			log.WithError(err).Error("fail to send key")
		}
	}
}
