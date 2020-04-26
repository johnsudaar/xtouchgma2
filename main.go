package main

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/johnsudaar/xtouchgma2/xtouch"
)

func main() {

	log := logger.Default()
	ctx := logger.ToCtx(context.Background(), log)
	xtouch.ClosestScribbleColor("")
	server1 := xtouch.NewServer(10111)
	go func() {
		panic(server1.Start(ctx))
	}()

	time.Sleep(2 * time.Second)

	//Vegas(ctx, server1)

	c, err := gma2ws.NewClient("192.168.1.21", "john", "john")
	if err != nil {
		panic(err)
	}

	stop, err := c.Start(ctx)
	if err != nil {
		panic(err)
	}
	defer stop()

	time.Sleep(3 * time.Second)

	log.Info("Start")
	for {
		time.Sleep(100 * time.Millisecond)
		playbacks, err := c.Playbacks(0, []gma2ws.PlaybacksRange{
			gma2ws.PlaybacksRange{
				Index: 0,
				Count: 10,
			},
		})
		if err != nil {
			panic(err)
		}

		for i := 0; i < 8; i++ {
			executor := playbacks[0].Items[i/5][i%5]
			f := executor.ExecutorBlocks[0].Fader
			if f.Max == 0 {
				server1.SetFaderPos(ctx, i, 0)
			} else {
				value := float64(f.Value) / float64(f.Max-f.Min)
				if value > 1 {
					value = 1
				}
				server1.SetFaderPos(ctx, i, value)
			}
			line1 := executor.TextTop.Text
			line2 := ""
			if len(executor.Cues.Items) == 3 {
				line2 = strings.TrimSpace(executor.Cues.Items[1].Text)
			} else if len(executor.Cues.Items) >= 1 {
				line2 = strings.TrimSpace(executor.Cues.Items[0].Text)
			}
			color, err := xtouch.ClosestScribbleColor(f.BorderColor)
			if err != nil {
				panic(err)
			}
			server1.SetScribble(ctx, i, color, true, line1, line2)
		}
	}
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
