package main

import (
	"context"

	"github.com/Scalingo/go-utils/logger"
	"github.com/johnsudaar/xtouchgma2/link"
)

func main() {
	log := logger.Default()
	ctx := logger.ToCtx(context.Background(), log)

	link, err := link.New(link.NewLinkParams{
		GMAHost:      "192.168.1.21",
		GMAUser:      "john",
		GMAPassword:  "john",
		SACNUniverse: 10,
	})
	if err != nil {
		panic(err)
	}
	defer link.Stop()

	err = link.Start(ctx)
	if err != nil {
		panic(err)
	}
}
