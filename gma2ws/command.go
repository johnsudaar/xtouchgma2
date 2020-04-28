package gma2ws

import (
	"context"

	"github.com/pkg/errors"
)

func (c *Client) SendCommand(ctx context.Context, command string) error {
	err := c.WriteJSON(ClientRequestCommand{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypeCommand,
			Session:     c.session,
			MaxRequests: 0,
		},
		Command: command,
	})
	if err != nil {
		return errors.Wrap(err, "fail to send command")
	}
	return nil
}
