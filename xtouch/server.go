package xtouch

import (
	"context"
	"net"
	"sync"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/johnsudaar/xtouchgma2/xtouch/transport/rtp"
	"github.com/johnsudaar/xtouchgma2/xtouch/transport/xctl"
	"github.com/pkg/errors"
)

type ServerType string

const (
	ServerTypeXTouch    ServerType = "xtouch"
	ServerTypeXTouchExt ServerType = "xtouch-ext"
)

type Server struct {
	listenerLock            *sync.RWMutex
	faderChangedListeners   []FaderChangedListener
	buttonChangedListeners  []ButtonChangedListener
	encoderChangedListeners []EncoderChangedListener
	transport               transport.Transport
	serverType              ServerType
	noteToButton            map[byte]Button
	buttonToNote            map[Button]byte
}

func NewServer(port int, serverType ServerType) *Server {
	server := &Server{
		listenerLock:            &sync.RWMutex{},
		faderChangedListeners:   make([]FaderChangedListener, 0),
		buttonChangedListeners:  make([]ButtonChangedListener, 0),
		encoderChangedListeners: make([]EncoderChangedListener, 0),
		serverType:              serverType,
	}
	// TODO: Send an error if the type is not found

	server.initButtons()

	if serverType == ServerTypeXTouch {
		server.transport = xctl.New(port, server)
	} else {
		server.transport = rtp.New(port, server)
	}
	return server
}

func (s *Server) Start(ctx context.Context) error {
	err := s.transport.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to start transport")
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.transport.Stop(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to stop transport")
	}
	return nil
}

func (s *Server) OnUDPPacket(ctx context.Context, from *net.UDPAddr, packet transport.MidiMessage) error {
	s.dispatchMidiMessage(ctx, packet)
	return nil
}
