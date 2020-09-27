package xctl

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

type XCtl struct {
	reader transport.Reader
	port   int

	conn       *net.UDPConn
	client     *net.UDPAddr
	socketLock *sync.Mutex
	stop       bool
	stopLock   *sync.Mutex
}

func New(port int, reader transport.Reader) *XCtl {
	return &XCtl{
		port:   port,
		reader: reader,
		stop:   false,

		socketLock: &sync.Mutex{},
		stopLock:   &sync.Mutex{},
	}
}

func (s *XCtl) Start(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		return errors.Wrapf(err, "fail to listen on 0.0.0.0:%v", s.port)
	}

	s.conn = conn

	go s.keepAliveLoop(ctx)

	go s.readLoop(ctx)

	now := time.Now()
	for {
		if time.Since(now) > 10*time.Second {
			return fmt.Errorf("fail to connect: timeout")
		}
		s.socketLock.Lock()
		client := s.client
		s.socketLock.Unlock()
		if client != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *XCtl) Stop(ctx context.Context) error {
	if s.conn != nil {
		s.conn.Close()
	}
	s.stopLock.Lock()
	s.stop = true
	s.stopLock.Unlock()
	return nil
}
