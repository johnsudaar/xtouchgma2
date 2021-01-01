package link

import (
	"context"
	"sync"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
	"github.com/pkg/errors"
)

var SACN_CID = [16]byte{0x13, 0x37, 0xde, 0xad, 0xbe, 0xef}

type Link struct {
	XTouches               XTouches
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
	encoderGMAValue        map[int]float64
	encoderLock            *sync.RWMutex
}

type XTouchParams struct {
	Type           xtouch.ServerType
	Port           int
	ExecutorOffset int
}

type NewLinkParams struct {
	GMAHost      string
	GMAUser      string
	GMAPassword  string
	SACNUniverse uint16
}

func New(params NewLinkParams) (*Link, error) {
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
		XTouches:               make([]XTouch, 0),
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
		encoderGMAValue:        make(map[int]float64),
		encoderLock:            &sync.RWMutex{},
	}

	return link, nil
}

func (l *Link) AddXTouch(ctx context.Context, params XTouchParams) error {
	if params.Type == xtouch.ServerTypeXTouch {
		for _, xt := range l.XTouches {
			if xt.xtouchType == xtouch.ServerTypeXTouch {
				return errors.New("only one xtouch can be added")
			}
		}
	}
	server := xtouch.NewServer(params.Port, params.Type)
	touch := XTouch{
		server:         server,
		xtouchType:     params.Type,
		executorOffset: params.ExecutorOffset,
		link:           l,
	}

	l.XTouches = append(l.XTouches, touch)
	touch.subscribeToEventChanges()
	return nil
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

	for _, xt := range l.XTouches {
		err = xt.server.Start(ctx)
		if err != nil {
			close(dmx)
			stop()
			l.stopAllXTouches(ctx)
			l.sacnDMX = nil
			l.stop = true
			return errors.Wrap(err, "fail to start xtouch")
		}
	}

	go l.startDMXSync(ctx)

	l.startEventLoop(ctx)
	return nil
}

func (l *Link) stopAllXTouches(ctx context.Context) {
	log := logger.Get(ctx)
	for _, xt := range l.XTouches {
		err := xt.server.Stop(ctx)
		if err != nil {
			log.WithError(err).Error("fail to stop xtouch")
		}
	}
}

func (l *Link) Stop(ctx context.Context) {
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
	l.stopAllXTouches(ctx)

	if l.sacnDMX != nil {
		close(l.sacnDMX)
		l.sacnDMX = nil
	}
}
