package link

import (
	"context"
	"sync"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
)

var SACN_CID = [16]byte{0x13, 0x37, 0xde, 0xad, 0xbe, 0xef}

type Link struct {
	XTouch                 *xtouch.Server
	GMA                    *gma2ws.Client
	SACN                   sacn.Transmitter
	sacnDMX                chan<- [512]byte
	sacnUniverse           uint16
	dmxUniverse            [512]byte
	dmxLock                *sync.Mutex
	gmaHost                string
	gmaStop                gma2ws.Stopper
	stop                   bool
	stopLock               *sync.RWMutex
	faderLock              *sync.Mutex
	faderPage              int
	encoderAttributes      [8]string
	encoderAttributesCoeff [8]int
	encoderAsAttributes    bool
	encoderGMAValue        [8]float64
	encoderLock            *sync.RWMutex
}

type NewLinkParams struct {
	GMAHost      string
	GMAUser      string
	GMAPassword  string
	SACNUniverse uint16
}

func New(params NewLinkParams) (*Link, error) {
	xtouch := xtouch.NewServer(10111)

	sacn, err := sacn.NewTransmitter("", SACN_CID, "XTOUCH_TO_GMA2")
	if err != nil {
		return nil, errors.Wrap(err, "fail to send sacn informations")
	}

	gma2, err := gma2ws.NewClient(params.GMAHost, params.GMAUser, params.GMAPassword)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create gma client")
	}
	link := &Link{
		GMA:                    gma2,
		XTouch:                 xtouch,
		SACN:                   sacn,
		sacnUniverse:           params.SACNUniverse,
		dmxUniverse:            [512]byte{},
		dmxLock:                &sync.Mutex{},
		gmaHost:                params.GMAHost,
		stop:                   false,
		stopLock:               &sync.RWMutex{},
		faderLock:              &sync.Mutex{},
		faderPage:              0,
		encoderAttributes:      [8]string{},
		encoderAttributesCoeff: [8]int{},
		encoderAsAttributes:    false,
		encoderLock:            &sync.RWMutex{},
	}

	xtouch.SubscribeToFaderChanges(link.onFaderChangeEvent)
	xtouch.SubscribeButtonChanges(link.onButtonChange)
	xtouch.SubscribeEncoderChanges(link.onEncoderChangedEvent)

	return link, nil
}

func (l *Link) Start(ctx context.Context) error {
	errs := l.SACN.SetDestinations(l.sacnUniverse, []string{l.gmaHost})
	if errs != nil {
		return errors.Wrap(errs[0], "fail to set sacn destination")
	}
	dmx, err := l.SACN.Activate(l.sacnUniverse)
	if err != nil {
		return errors.Wrap(err, "fail to start sacn")
	}

	l.sacnDMX = dmx

	stop, err := l.GMA.Start(ctx)
	if err != nil {
		close(dmx)
		l.sacnDMX = nil
		l.stop = true
		return errors.Wrap(err, "fail to start gma")
	}
	l.gmaStop = stop

	err = l.XTouch.Start(ctx)
	if err != nil {
		close(dmx)
		stop()
		l.sacnDMX = nil
		l.stop = true
		return errors.Wrap(err, "fail to start xtouch")
	}

	go l.startDMXSync(ctx)

	l.startEventLoop(ctx)
	return nil
}

func (l *Link) Stop() {
	l.stopLock.Lock()
	if l.stop {
		l.stopLock.Unlock()
		return
	}

	l.stop = true
	l.stopLock.Unlock()
	if l.gmaStop != nil {
		l.gmaStop()
		l.gmaStop = nil
	}
	if l.XTouch != nil {
		l.XTouch.Stop()
		l.XTouch = nil
	}

	if l.sacnDMX != nil {
		close(l.sacnDMX)
		l.sacnDMX = nil
	}
}
