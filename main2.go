package main

import (
	"time"

	"github.com/johnsudaar/xtouchgma2/xtouch"
)

//func main() {
//	p := make([]byte, 2048)
//	conn, err := net.Dial("udp", "192.168.50.55:5004")
//	if err != nil {
//		fmt.Printf("Some error %v", err)
//		return
//	}
//	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
//	_, err = bufio.NewReader(conn).Read(p)
//	if err == nil {
//		fmt.Printf("%s\n", p)
//	} else {
//		fmt.Printf("Some error %v\n", err)
//	}
//	conn.Close()
//}

func main() {
	server1 := xtouch.NewServer(5004, "XTOUCH2GMA2")
	go func() {
		panic(server1.Start())
	}()

	server2 := xtouch.NewServer(5005, "XTOUCH2GMA2")
	go func() {
		panic(server2.Start())
	}()

	time.Sleep(2 * time.Second)
	server2.SendMidiPacket([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xf7,
	})

	/*
		server2.SendMidiPacket([]byte{
			0xf0, 0x00, 0x00, 0x66, 0x14, 0x12, 0x01, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0xf7,
			//                              |     |    \--------- AAAAAAAAAAAAAAAAA -------------------/
		})//                              |     \- Index
		  //                              \- Scribble write
	*/

	server2.SendMidiPacket([]byte{
		0xf0, 0x00, 0x00, 0x66, 0x58, 0x20, 0x41, 0x43, 0x68, 0x20, 0x31, 0x00, 0x00, 0x00, 0x20, 0x20, 0x20, 0x10, 0x61, 0x42, 0x33, 0xf7,
	})
	for {
		//server2.SendMidiPacket([]byte{
		//	0xf0, 0x00, 0x00, 0x66, 0x14, 0x00, 0xf7,
		//})

		time.Sleep(1 * time.Second)
		time.Sleep(4 * time.Second)
	}
}
