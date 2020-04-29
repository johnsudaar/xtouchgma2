package gma2ws

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/pkg/errors"
)

type KeyName string
type KeyStatus string

const (
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

	KeyStatusOff   KeyStatus = "off"
	KeyStatusOn    KeyStatus = "on"
	KeyStatusBlink KeyStatus = "blink"
)

var autoSubmitKeys map[KeyName]bool = map[KeyName]bool{
	KeyNameHighlight: true,
	KeyNameSolo:      true,
}

func (c *Client) SendKey(ctx context.Context, key KeyName, pressed bool) error {
	status := 0
	if pressed {
		status = 1
	}
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

func (c *Client) serverResponseHandleGetData(ctx context.Context, buffer []byte) {
	log := logger.Get(ctx)
	var getData ServerResponseGetData
	err := json.Unmarshal(buffer, &getData)
	if err != nil {
		log.WithError(err).Error("fail to unmarshal getdata")
		return
	}

	select {
	case c.getDataChan <- getData.Data:
	default:
	}
}

func (c *Client) KeyStatuses(keys ...string) (map[string]KeyStatus, error) {
	request := ClientRequestGetData{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypeGetData,
			Session:     c.session,
			MaxRequests: 1,
		},
		Data: strings.Join(keys, ","),
	}

	err := c.WriteJSON(request)
	if err != nil {
		return nil, errors.Wrap(err, "fail to send getdata request")
	}

	timer := time.NewTimer(2 * time.Second)
	var data []map[string]string
	select {
	case data = <-c.getDataChan:
	case <-timer.C:
		return nil, errors.New("timeout")
	}

	res := make(map[string]KeyStatus)
	for _, d := range data {
		for k, v := range d {
			switch v {
			case "0":
				res[k] = KeyStatusOff
			case "1":
				res[k] = KeyStatusOn
			case "b":
				res[k] = KeyStatusBlink
			}
		}
	}
	return res, nil
}
