package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/gdamore/tcell"
	"github.com/johnsudaar/xtouchgma2/gma2ws"
	"github.com/rivo/tview"
)

func main() {
	log := logger.Default()
	ctx := logger.ToCtx(context.Background(), log)
	c, err := gma2ws.NewClient("192.168.1.21", "john", "john")
	if err != nil {
		panic(err)
	}

	stop, err := c.Start(ctx)
	if err != nil {
		panic(err)
	}
	defer stop()

	app := tview.NewApplication()
	grid := tview.NewGrid()
	app.SetRoot(grid, true)
	go app.Run()

	for {
		playbacks, err := c.Playbacks(0, []gma2ws.PlaybacksRange{
			gma2ws.PlaybacksRange{
				Index: 0,
				Count: 5,
			},
		})
		if err != nil {
			log.WithError(err).Error("Error while refreshing playbacks")
		}
		grid.Clear()

		for i, executor := range playbacks[0].Items[0] {
			box := CreateBoxForExecutor(executor)
			grid.AddItem(box, 0, i, 1, 1, 0, 0, true)
		}
		app.Draw()
		time.Sleep(100 * time.Millisecond)
	}
}

func CreateBoxForExecutor(executor gma2ws.ServerPlayback) *tview.Grid {
	grid := tview.NewGrid().SetRows(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	// HEADER
	headerGrid := tview.NewGrid().SetColumns(0, 0, 0).SetRows(0, 0, 0)
	headerGrid.AddItem(TextViewFromTextItem(
		executor.Index, executor.HeaderBackgroundColor, tview.AlignLeft),
		0, 0, 1, 1, 0, 0, true)
	headerGrid.AddItem(
		TextViewFromTextItem(executor.ObjectType, executor.HeaderBackgroundColor, tview.AlignCenter),
		0, 1, 1, 1, 0, 0, true)
	headerGrid.AddItem(
		TextViewFromTextItem(executor.ObjectIndex, executor.HeaderBackgroundColor, tview.AlignRight),
		0, 2, 1, 1, 0, 0, true)
	headerGrid.AddItem(
		TextViewFromTextItem(executor.TextTop, executor.HeaderBackgroundColor, tview.AlignLeft),
		1, 0, 2, 3, 0, 0, true)
	headerGrid.SetBorder(true)
	headerGrid.SetBorderColor(tcell.GetColor(executor.HeaderBorderColor))
	grid.AddItem(headerGrid, 0, 0, 2, 1, 0, 0, true)

	// Cues
	cuesGrid := tview.NewGrid()
	for i, cue := range executor.Cues.Items {
		cuesGrid.AddItem(
			TextViewFromTextItem(cue, executor.Cues.BackgroundColor, tview.AlignCenter),
			i, 0, 1, 1, 0, 0, true,
		)
	}
	grid.AddItem(cuesGrid, 2, 0, 3, 1, 0, 0, true)

	// Buttons
	b3 := executor.ExecutorBlocks[0].Button3
	button3 := tview.NewButton(b3.Text)
	button3.SetBackgroundColor(tcell.GetColor(b3.Color))
	button3.SetBorder(true).SetBorderColor(tcell.GetColor(b3.BorderColor))
	grid.AddItem(button3, 5, 0, 2, 1, 0, 0, true)
	b2 := executor.ExecutorBlocks[0].Button2
	button2 := tview.NewButton(b2.Text)
	button2.SetBackgroundColor(tcell.GetColor(b2.Color))
	button2.SetBorder(true).SetBorderColor(tcell.GetColor(b2.BorderColor))
	grid.AddItem(button2, 7, 0, 2, 1, 0, 0, true)
	b1 := executor.ExecutorBlocks[0].Button1
	button1 := tview.NewButton(b1.Text)
	button1.SetBackgroundColor(tcell.GetColor(b1.Color))
	button1.SetBorder(true).SetBorderColor(tcell.GetColor(b1.BorderColor))
	grid.AddItem(button1, 18, 0, 2, 1, 0, 0, true)

	f := executor.ExecutorBlocks[0].Fader
	for i := 0; i < 9; i++ {
		fader := tview.NewTextView()
		lowThreshold := (f.Max - f.Min) / float64(9) * float64(i)
		highThreshold := (f.Max - f.Min) / float64(9) * float64(i+1)
		if (f.Value >= lowThreshold && f.Value <= highThreshold && f.Max != 0) || (f.Max == 0 && i == 0) {
			fader.SetText(fmt.Sprintf("%s - %s", f.TypeText, f.ValueText))
			fader.SetTextAlign(tview.AlignCenter)
			fader.SetBackgroundColor(tcell.GetColor(f.Color))
			fader.SetBorder(true).SetBorderColor(tcell.GetColor(f.BorderColor))
		} else {
			fader.SetText("|")
			fader.SetTextColor(tcell.GetColor(f.BorderColor))
			fader.SetTextAlign(tview.AlignCenter)
		}
		grid.AddItem(fader, 17-i, 0, 1, 1, 0, 0, true)
	}

	return grid
}

func TextViewFromTextItem(s gma2ws.ServerPlaybackTextItem, bg string, align int) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetText(s.Text)
	tv.SetTextColor(tcell.GetColor(s.Color))
	if s.Progress != nil && s.Progress.Value > 0 {
		tv.SetBackgroundColor(tcell.GetColor(s.Progress.BackgroundColor))
	} else {
		tv.SetBackgroundColor(tcell.GetColor(bg))
	}
	tv.SetTextAlign(align)
	return tv
}
