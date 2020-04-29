package gma2ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

type PlaybacksItemType int

const (
	PlaybacksItemTypeFader  PlaybacksItemType = 2
	PlaybacksItemTypeButton PlaybacksItemType = 3
)

type PlaybacksRange struct {
	Index    int
	Count    int
	ItemType PlaybacksItemType
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
		itemType[i] = int(r.ItemType)
	}
	request := ClientRequestPlaybacks{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypePlaybacks,
			Session:     c.session,
			MaxRequests: 1,
		},
		PageIndex:          page,
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

func (c *Client) FaderChanged(ctx context.Context, executor, page int, value float64) error {
	err := c.WriteJSON(ClientRequestPlaybacksUserInput{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypePlaybacksUserInput,
			Session:     c.session,
			MaxRequests: 0,
		},
		Executor: executor,
		Page:     page,
		Value:    value,
		Type:     1,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send fader values")
	}
	return nil
}

func (c *Client) ButtonChanged(ctx context.Context, executor, page, button int, value bool) error {
	err := c.WriteJSON(ClientRequestPlaybacksUserInput{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypePlaybacksUserInput,
			Session:     c.session,
			MaxRequests: 0,
		},
		Executor: executor,
		Page:     page,
		Type:     0,
		ButtonID: button,
		Pressed:  value,
		Released: !value,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send button")
	}
	return nil

}
