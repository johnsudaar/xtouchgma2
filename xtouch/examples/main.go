package main

import (
	"context"
	"math"
	"time"

	"github.com/johnsudaar/xtouchgma2/xtouch"
)

func main() {
	server1 := xtouch.NewServer(10111)
	defer server1.Stop()
	err := server1.Start(context.Background())
	if err != nil {
		panic(err)
	}
	Vegas(context.Background(), server1)
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
		for button, _ := range xtouch.ButtonToNote {
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
		for button, _ := range xtouch.ButtonToNote {
			server1.SetButtonStatus(ctx, button, xtouch.ButtonStatusOff)
			time.Sleep(25 * time.Millisecond)
		}
	}
}
