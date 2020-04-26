package xtouch

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

type Server struct {
	Port       int
	conn       *net.UDPConn
	client     *net.UDPAddr
	socketLock *sync.Mutex
	stop       bool
	stopLock   *sync.Mutex
}

func NewServer(port int) *Server {
	return &Server{
		Port:       port,
		socketLock: &sync.Mutex{},
		stopLock:   &sync.Mutex{},
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.Get(ctx)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.Port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		return errors.Wrapf(err, "fail to listen on 0.0.0.0:%v", s.Port)
	}

	s.conn = conn

	go func() {
		for {
			s.stopLock.Lock()
			stop := s.stop
			s.stopLock.Unlock()
			if stop {
				return
			}
			time.Sleep(6 * time.Second)
			s.keepAlive(ctx)
		}
	}()

	go func() {
		buffer := make([]byte, 1024)
		for {
			s.stopLock.Lock()
			stop := s.stop
			s.stopLock.Unlock()
			if stop {
				return
			}
			sendKeepalive := false
			n, from, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.WithError(err).Error("fail to read udp buffer")
				continue
			}
			s.socketLock.Lock()
			if s.client == nil {
				sendKeepalive = true
			}
			s.client = from
			s.socketLock.Unlock()

			if sendKeepalive {
				s.keepAlive(ctx)
			}

			packet := buffer[:n]

			log.WithField("from", from).Debug(hex.Dump(packet))
			if buffer[0] < 0xf0 {
				var midiMessage MidiMessage
				midiMessage.UnmarshalBinary(buffer)
				fmt.Printf("%+v\n", midiMessage)
			}
		}
	}()

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

func (s *Server) Stop() {
	if s.conn != nil {
		s.conn.Close()
	}
	s.stopLock.Lock()
	s.stop = true
	s.socketLock.Unlock()
}

func (s *Server) keepAlive(ctx context.Context) {
	s.socketLock.Lock()
	defer s.socketLock.Unlock()
	log := logger.Get(ctx)
	if s.client == nil {
		return
	}

	_, err := s.conn.WriteTo([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xf7,
	}, s.client)
	if err != nil {
		log.WithError(err).Error("fail to send xtouch heartbeat")
	}
}

func (s *Server) SendRawPacket(ctx context.Context, buff []byte) error {
	s.socketLock.Lock()
	defer s.socketLock.Unlock()
	log := logger.Get(ctx)
	log.WithField("send_to", s.client).Debug(hex.Dump(buff))
	_, err := s.conn.WriteToUDP(buff, s.client)
	if err != nil {
		return errors.Wrap(err, "fail to send message")
	}
	return nil
}

func (s *Server) SendSysExPacket(ctx context.Context, message []byte) error {
	buff := new(bytes.Buffer)
	buff.Write([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x58,
	})

	buff.Write(message)
	buff.WriteByte(0xf7)
	err := s.SendRawPacket(ctx, buff.Bytes())
	if err != nil {
		return errors.Wrap(err, "fail to send sysex packet")
	}
	return nil
}

func (s *Server) SendMidiPacket(ctx context.Context, packet MidiMessage) error {
	buff, err := packet.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "fail to marshal midi packet")
	}

	err = s.SendRawPacket(ctx, buff)
	if err != nil {
		return errors.Wrap(err, "fail to send midi packet")
	}
	return nil
}
