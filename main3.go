package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 10111,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, 1024)
	first := true
	for {
		n, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		packet := buffer[:n]
		fmt.Println(from)

		fmt.Printf("Received:\n%s\n", hex.Dump(packet))
		if first {
			first = false
			go func() {
				for {

					conn.WriteTo([]byte{
						0xf0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xf7,
					}, from)
					conn.WriteTo([]byte{
						0xf0, 0x00, 0x00, 0x66, 0x58, 0x20, 0x41, 0x43, 0x68, 0x20, 0x31, 0x00, 0x00, 0x00, 0x20, 0x20, 0x20, 0x10, 0x61, 0x42, 0x33, 0xf7,
					}, from)

					time.Sleep(6 * time.Second)
				}
			}()
		}
	}
}
