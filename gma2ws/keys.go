package gma2ws

import (
	"context"

	"github.com/pkg/errors"
)

type KeyName string
type KeyStatus int

const (
	KeyStatusPressed  KeyStatus = 1
	KeyStatusReleased KeyStatus = 0

	KeyName0 KeyName = "0"
	KeyName1 KeyName = "1"
	KeyName2 KeyName = "2"
	KeyName3 KeyName = "3"
	KeyName4 KeyName = "4"
	KeyName5 KeyName = "5"
	KeyName6 KeyName = "6"
	KeyName7 KeyName = "7"
	KeyName8 KeyName = "8"
	KeyName9 KeyName = "9"

	KeyNameOn        KeyName = "ON"
	KeyNameOff       KeyName = "OFF"
	KeyNameHighlight KeyName = "HIGH"
	KeyNameSolo      KeyName = "SOLO"
	KeyNameSelect    KeyName = "SELECT"
)

var autoSubmitKeys map[KeyName]bool = map[KeyName]bool{
	KeyNameHighlight: true,
	KeyNameSolo:      true,
}

func (c *Client) SendKey(ctx context.Context, key KeyName, status KeyStatus) error {
	autoSubmit := autoSubmitKeys[key]
	err := c.WriteJSON(ClientRequestKeyName{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypeKeyname,
			Session:     c.session,
			MaxRequests: 0,
		},
		KeyName:    key,
		Value:      status,
		AutoSubmit: autoSubmit,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send fader values")
	}
	return nil
}
