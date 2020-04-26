package xtouch

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

type Server struct {
	Port   int
	conn   *net.UDPConn
	client *net.UDPAddr
}

func NewServer(port int) *Server {
	return &Server{
		Port: port,
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
	defer conn.Close()

	s.conn = conn

	go func() {
		for {
			time.Sleep(6 * time.Second)
			s.keepAlive(ctx)
		}
	}()

	buffer := make([]byte, 1024)
	for {
		n, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return errors.Wrap(err, "fail to read udp buffer")
		}
		if s.client == nil {
			s.client = from
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

	return nil
}

func (s *Server) keepAlive(ctx context.Context) {
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
