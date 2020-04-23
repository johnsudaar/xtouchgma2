package xtouch

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
)

type Server struct {
	Port       int
	Name       string
	SSRC       uint32
	conn       *net.UDPConn
	sequNumber uint16
}

func NewServer(port int, name string) *Server {
	return &Server{
		Port: port,
		Name: name,
		SSRC: 0x12345678, // TODO: Generate it randomly
	}
}

func (s *Server) Start() error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.Port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		return errors.Wrapf(err, "fail to listen on 0.0.0.0:%v", s.Port)
	}
	defer conn.Close()

	s.conn = conn

	buffer := make([]byte, 1024)
	for {
		n, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return errors.Wrap(err, "fail to read udp buffer")
		}

		packet := buffer[:n]
		fmt.Println(from)

		fmt.Printf("Received:\n%s\n", hex.Dump(packet))
		if packet[0] == 0xff && packet[1] == 0xff {
			err := s.handleAppleMessage(packet, from)
			if err != nil {
				return errors.Wrap(err, "fail to handle apple packet")
			}
		} else {
			err := s.handleDataMessage(packet, from)
			if err != nil {
				return errors.Wrap(err, "fail to decode data packet")
			}
		}
	}

	return nil
}

func (s *Server) handleAppleMessage(packet []byte, from *net.UDPAddr) error {
	var exchange AppleMidiExchangePacket
	err := exchange.UnmarshalBinary(packet)
	if err != nil {
		return errors.Wrap(err, "fail to parse apple midi message")
	}

	if exchange.Command == AppleMidiCommandInvitation {
		fmt.Printf("New client: %s\n", exchange.Name)
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
		fmt.Printf("Send:\n%s\n", hex.Dump(responseBuffer))

		_, err = s.conn.WriteToUDP(responseBuffer, from)
		if err != nil {
			return errors.Wrap(err, "fail to send response")
		}
	}

	if exchange.Command == AppleMidiCommandInvitationTimestampSync {
		fmt.Println("Timestamp sync")
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
		fmt.Printf("Send:\n%s\n", hex.Dump(responseBuffer))

		_, err = s.conn.WriteToUDP(responseBuffer, from)
		if err != nil {
			return errors.Wrap(err, "fail to send response")
		}

	}

	return nil
}

func (s *Server) handleDataMessage(packet []byte, from *net.UDPAddr) error {
	var dataPacket AppleMidiDataPacket
	err := dataPacket.UnmarshalBinary(packet)
	if err != nil {
		return errors.Wrap(err, "fail to decode data packet")
	}

	fmt.Printf("%+v\n", dataPacket)

	return nil
}

func (s *Server) SendMidiPacket(buff []byte) error {
	s.sequNumber++
	to, err := net.ResolveUDPAddr("udp", "192.168.55.50:5005")
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
	fmt.Printf("Send:\n%s\n", hex.Dump(responseBuffer))
	_, err = s.conn.WriteToUDP(responseBuffer, to)
	if err != nil {
		return errors.Wrap(err, "fail to send message")
	}
	return nil
}

func (s *Server) SendSysExPacket(message []byte) error {
	buff := new(bytes.Buffer)
	buff.Write([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x58,
	})

	buff.Write(message)
	buff.WriteByte(0xf7)
	err := s.SendMidiPacket(buff.Bytes())
	if err != nil {
		return errors.Wrap(err, "fail to send sysex packet")
	}
	return nil
}
