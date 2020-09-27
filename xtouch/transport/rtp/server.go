package rtp

import (
	"bytes"
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

type Server struct {
	Port       int
	Name       string
	SSRC       uint32
	conn       *net.UDPConn
	client     *net.UDPAddr
	sequNumber uint16

	socketLock *sync.Mutex
	stop       bool
	stopLock   *sync.Mutex
	reader     transport.Reader
}

func NewServer(port int, name string, reader transport.Reader) *Server {
	return &Server{
		Port:   port,
		Name:   name,
		SSRC:   0x32320202,
		reader: reader,

		socketLock: &sync.Mutex{},
		stopLock:   &sync.Mutex{},
		stop:       false,
	}
}

func (s *Server) Start(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.Port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		return errors.Wrapf(err, "fail to listen on 0.0.0.0:%v", s.Port)
	}

	s.conn = conn

	go s.readLoop(ctx)

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.stopLock.Lock()
	s.stop = true
	s.stopLock.Unlock()
	s.socketLock.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.socketLock.Unlock()
	return nil
}

func (s *Server) readLoop(ctx context.Context) {
	log := logger.Get(ctx)
	for {
		s.stopLock.Lock()
		stop := s.stop
		s.stopLock.Unlock()
		if stop {
			return
		}

		buffer := make([]byte, 1024)
		n, from, err := s.conn.ReadFromUDP(buffer)
		if err != nil {
			log.WithError(err).Error("fail to read UDP buffer")
			continue
		}

		s.socketLock.Lock()
		s.client = from
		s.socketLock.Unlock()

		packet := buffer[:n]

		if packet[0] == 0xff && packet[1] == 0xff {
			err := s.handleAppleMessage(ctx, packet, from)
			if err != nil {
				log.WithError(err).Error("fail to handle apple packet")
				continue
			}
		} else {
			err := s.handleDataMessage(ctx, packet, from)
			if err != nil {
				log.WithError(err).Error("fail to handle data packet")
				continue
			}
		}
	}
}

func (s *Server) handleAppleMessage(ctx context.Context, packet []byte, from *net.UDPAddr) error {
	log := logger.Get(ctx)
	var exchange AppleMidiExchangePacket
	err := exchange.UnmarshalBinary(packet)
	if err != nil {
		return errors.Wrap(err, "fail to parse apple midi message")
	}

	if exchange.Command == AppleMidiCommandInvitation {
		log.Infof("New client %s", exchange.Name)
		response := AppleMidiExchangePacket{
			Preamble:       0xffff,
			Command:        AppleMidiCommandInvitationAccepted,
			Version:        2,
			InitiatorToken: exchange.InitiatorToken,
			SSRC:           s.SSRC,
			Name:           s.Name,
		}

		responseBuffer, err := response.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "fail to marshal new message")
		}
		log.Debugf("Send:\n%s", hex.Dump(responseBuffer))

		_, err = s.conn.WriteToUDP(responseBuffer, from)
		if err != nil {
			return errors.Wrap(err, "fail to send response")
		}
	}

	if exchange.Command == AppleMidiCommandInvitationTimestampSync {
		log.Debug("Timestamp sync")
		var timestamp AppleMidiTimestampPacket
		err := timestamp.UnmarshalBinary(packet)
		if err != nil {
			return errors.Wrap(err, "fail to decode apple timestamp packet")
		}

		now := time.Now().Unix() * 100000

		response := AppleMidiTimestampPacket{
			Preamble:   0xffff,
			Command:    AppleMidiCommandInvitationTimestampSync,
			Count:      1,
			SSRC:       s.SSRC,
			Timestamp1: timestamp.Timestamp1,
			Timestamp2: uint64(now),
			Timestamp3: 0,
		}
		responseBuffer, err := response.MarshalBinary()
		if err != nil {
			return errors.Wrap(err, "fail to marshal new message")
		}
		log.Debugf("Send:\n%s\n", hex.Dump(responseBuffer))

		_, err = s.conn.WriteToUDP(responseBuffer, from)
		if err != nil {
			return errors.Wrap(err, "fail to send response")
		}

	}

	return nil
}

func (s *Server) handleDataMessage(ctx context.Context, packet []byte, from *net.UDPAddr) error {
	log := logger.Get(ctx)
	var dataPacket AppleMidiDataPacket
	err := dataPacket.UnmarshalBinary(packet)
	if err != nil {
		return errors.Wrap(err, "fail to decode data packet")
	}

	if len(dataPacket.MidiMessage) == 0 {
		return nil
	}
	log.Debugf("Received: %+v", dataPacket)

	var midiMessage transport.MidiMessage
	err = midiMessage.UnmarshalBinary(dataPacket.MidiMessage)
	if err != nil {
		return errors.Wrap(err, "fail to decode midi message")
	}

	err = s.reader.OnUDPPacket(ctx, from, midiMessage)
	if err != nil {
		return errors.Wrap(err, "reader failed parse packet")
	}
	return nil
}

func (s *Server) SendMidiPacket(ctx context.Context, buff []byte) error {
	log := logger.Get(ctx)
	s.sequNumber++
	response := AppleMidiDataPacket{
		V:              2,
		P:              false,
		X:              false,
		CC:             0,
		M:              false,
		PayloadType:    97,
		SequenceNumber: s.sequNumber,
		Timestamp:      uint32(time.Now().Unix() * 100000),
		SSRC:           s.SSRC,
		Big:            true,
		Journal:        false,
		Z:              false,
		P2:             false,
		Len:            uint16(len(buff)),
		MidiMessage:    buff,
	}
	responseBuffer, err := response.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "fail to marshal new message")
	}

	s.socketLock.Lock()
	to := s.client
	s.socketLock.Unlock()

	log.Debugf("Send:\n%s", hex.Dump(responseBuffer))

	_, err = s.conn.WriteToUDP(responseBuffer, to)
	if err != nil {
		return errors.Wrap(err, "fail to send message")
	}
	return nil
}

func (s *Server) SendSysExPacket(ctx context.Context, message []byte) error {
	buff := new(bytes.Buffer)
	buff.Write([]byte{
		0xf0, 0x00, 0x20, 0x32, 0x15, 0x4c,
	})

	buff.Write(message)
	buff.WriteByte(0xf7)
	err := s.SendMidiPacket(ctx, buff.Bytes())
	if err != nil {
		return errors.Wrap(err, "fail to send sysex packet")
	}
	return nil
}
