package main

import (
	"context"
	"math"
	"time"

	"github.com/johnsudaar/xtouchgma2/xtouch"
)

func main() {
	ctx := context.Background()
	server1 := xtouch.NewServer(10111, xtouch.ServerTypeXTouch)
	server2 := xtouch.NewServer(5004, xtouch.ServerTypeXTouchExt)
	defer server1.Stop(ctx)
	defer server2.Stop(ctx)
	err := server1.Start(ctx)
	if err != nil {
		panic(err)
	}
	err = server2.Start(ctx)
	if err != nil {
		panic(err)
	}

	go Vegas(context.Background(), server1)
	Vegas(context.Background(), server2)
}
func Vegas(ctx context.Context, server1 *xtouch.Server) {
	go func() {
		var t float64 = 0.01
		x := 0
		for {
			for i := 0; i < 9; i++ {
				res := math.Sin(t+float64(i)/2)/2 + 0.5
				server1.SetFaderPos(ctx, i, res)
				if i != 8 {
					server1.SetRingPosition(ctx, i, res)
				}
				if i == x {
					server1.SetScribble(ctx, i, xtouch.ScribbleColorRed, false, "A", "B")
				} else {
					server1.SetScribble(ctx, i, xtouch.ScribbleColorBlack, false, "A", "B")
				}
			}
			x++
			if x == 8 {
				x = 0
			}
			t += 0.03
			time.Sleep(50 * time.Millisecond)
		}
	}()

	for {
		for _, button := range server1.ButtonsSupported() {
			server1.SetButtonStatus(ctx, button, xtouch.ButtonStatusOn)
			time.Sleep(25 * time.Millisecond)
		}

		for i := 0; i < 8; i++ {
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionSelect, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionMute, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionSolo, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionRec, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
		}
		for i := 0; i < 8; i++ {
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionSelect, xtouch.ButtonStatusOn)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionMute, xtouch.ButtonStatusOn)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionSolo, xtouch.ButtonStatusOn)
			time.Sleep(25 * time.Millisecond)
			server1.SetFaderButtonStatus(ctx, i, xtouch.FaderButtonPositionRec, xtouch.ButtonStatusOn)
			time.Sleep(25 * time.Millisecond)
		}

		for _, button := range server1.ButtonsSupported() {
			server1.SetButtonStatus(ctx, button, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
		}
	}
}
