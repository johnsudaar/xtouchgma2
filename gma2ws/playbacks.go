package gma2ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

type PlaybacksRange struct {
	Index int
	Count int
}

type PlaybackResonse struct {
	Response []ServerPlaybacks
	Error    error
}

func (c *Client) Playbacks(page int, ranges []PlaybacksRange) ([]ServerPlaybacks, error) {
	if len(ranges) == 0 {
		return nil, fmt.Errorf("You should provide at least one range")
	}

	startIndex := make([]int, len(ranges))
	itemCount := make([]int, len(ranges))
	itemType := make([]int, len(ranges))
	for i, r := range ranges {
		if r.Count%5 != 0 || r.Count == 0 {
			return nil, errors.New("fader count should be a multiple of 5")
		}
		startIndex[i] = r.Index
		itemCount[i] = r.Count
		itemType[i] = 2
	}
	request := ClientRequestPlaybacks{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypePlaybacks,
			Session:     c.session,
			MaxRequests: 1,
		},
		PageIndex:          0,
		StartIndex:         startIndex,
		ItemsCount:         itemCount,
		ItemsType:          itemType,
		View:               2,
		ExecButtonViewMode: ExecButtonViewModeFader,
		ButtonsViewMode:    0,
	}

	err := c.WriteJSON(request)
	if err != nil {
		return nil, errors.Wrap(err, "fail to send playback request")
	}

	timer := time.NewTimer(2 * time.Second)
	select {
	case playbacks := <-c.playbackChan:
		return playbacks, nil
	case <-timer.C:
		return nil, errors.New("Timeout")
	}
}

func (c *Client) serverResponseHandlePlaybacks(ctx context.Context, buffer []byte) {
	log := logger.Get(ctx)
	var playbacks ServerResponsePlayback
	err := json.Unmarshal(buffer, &playbacks)
	if err != nil {
		log.WithError(err).Error("fail to unmarshal playbacks")
		return
	}
	select {
	case c.playbackChan <- playbacks.ItemGroups:
	default:
	}
}
