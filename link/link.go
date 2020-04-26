package link

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
)

type Link struct {
	XTouch  *xtouch.Server
	GMA     *gma2ws.Client
	gmaStop gma2ws.Stopper
}

type NewLinkParams struct {
	GMAHost     string
	GMAUser     string
	GMAPassword string
}

func New(params NewLinkParams) (*Link, error) {
	xtouch := xtouch.NewServer(10111)
	gma2, err := gma2ws.NewClient(params.GMAHost, params.GMAUser, params.GMAPassword)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create gma client")
	}
	return &Link{
		GMA:    gma2,
		XTouch: xtouch,
	}, nil
}

func (l *Link) Start(ctx context.Context) error {
	stop, err := l.GMA.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to start gma")
	}
	l.gmaStop = stop

	err = l.XTouch.Start(ctx)
	if err != nil {
		stop()
		return errors.Wrap(err, "fail to start xtouch")
	}

	l.startEventLoop(ctx)
	return nil
}

func (l *Link) Stop() {
	if l.gmaStop != nil {
		l.gmaStop()
	}
	if l.XTouch != nil {
		l.XTouch.Stop()
	}
}
