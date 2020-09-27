package rtp

import (
	"context"

	"github.com/johnsudaar/xtouchgma2/xtouch/transport"
	"github.com/pkg/errors"
)

type RTP struct {
	port    int
	reader  transport.Reader
	server1 *Server
	server2 *Server
}

func New(port int, reader transport.Reader) *RTP {
	return &RTP{
		port:    port,
		reader:  reader,
		server1: NewServer(port, "XTOUCH2GMA2", reader),
		server2: NewServer(port+1, "XTOUCH2GMA2", reader),
	}
}

func (r *RTP) Start(ctx context.Context) error {
	err := r.server1.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to start server1")
	}

	err = r.server2.Start(ctx)
	if err != nil {
		r.server1.Stop(ctx)
		return errors.Wrap(err, "fail to start server2")
	}
	// TODO: Check connected
	return nil
}

func (r *RTP) Stop(ctx context.Context) error {
	err := r.server1.Stop(ctx)
	err2 := r.server2.Stop(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to stop server1")
	}
	if err2 != nil {
		return errors.Wrap(err, "fail to stop server2")
	}
	return nil
}

func (r *RTP) SendMidiPacket(ctx context.Context, packet transport.MidiMessage) error {
	buff, err := packet.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "fail to marshal midi packet")
	}
	err = r.server2.SendMidiPacket(ctx, buff)
	if err != nil {
		return errors.Wrap(err, "fail to send midi packet")
	}
	return nil
}

func (r *RTP) SendSysExPacket(ctx context.Context, message []byte) error {
	err := r.server2.SendSysExPacket(ctx, message)
	if err != nil {
		return errors.Wrap(err, "fail to send SysEX packet")
	}
	return nil
}
