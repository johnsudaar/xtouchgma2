package transport

import (
	"context"
	"net"
)

type Transport interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	SendMidiPacket(ctx context.Context, packet MidiMessage) error
	SendSysExPacket(ctx context.Context, message []byte) error
}

type Reader interface {
	OnUDPPacket(ctx context.Context, from *net.UDPAddr, packet MidiMessage) error
}
