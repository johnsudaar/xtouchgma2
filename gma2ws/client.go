package gma2ws

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Stopper func()

var InvalidPasswordError error = fmt.Errorf("Invalid password")

type Client struct {
	host                string
	user                string
	hashedPassword      string
	session             int
	writeLock           *sync.Mutex
	ws                  *websocket.Conn
	playbackChan        chan []ServerPlaybacks
	stopResponseHandler chan bool
}

func NewClient(host, user, password string) (*Client, error) {
	hash := md5.Sum([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])
	return &Client{
		host:                host,
		user:                user,
		hashedPassword:      hashedPassword,
		playbackChan:        make(chan []ServerPlaybacks),
		stopResponseHandler: make(chan bool),
		writeLock:           &sync.Mutex{},
	}, nil
}

func (c *Client) Start(ctx context.Context) (Stopper, error) {
	log := logger.Get(ctx)

	url := fmt.Sprintf("ws://%s:80/?ma=1", c.host)
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create Websocket")
	}
	c.ws = ws

	err = c.login(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to login")
	}
	go c.serverResponseHandler(ctx)
	go c.startKeepAlive()

	return func() {
		log.Info("Close")
		c.ws.WriteJSON(
			ClientRequestGeneric{
				MaxRequests: 0,
				Session:     c.session,
				RequestType: RequestTypeClose,
			},
		)
		c.stopResponseHandler <- true
	}, nil

}

func (c *Client) login(ctx context.Context) error {
	log := logger.Get(ctx)
	err := c.WriteJSON(ClientHanshake{})
	if err != nil {
		return errors.Wrap(err, "fail to send client handshake")
	}

	_, buffer, err := c.ws.ReadMessage()
	if err != nil {
		return errors.Wrap(err, "fail to read first server response")
	}
	log.WithField("source", "gma2").Debug(string(buffer))
	// Scrap that, we all know that you're a grandMA2 stop bragging

	_, buffer, err = c.ws.ReadMessage()
	if err != nil {
		return errors.Wrap(err, "fail to read second server response")
	}
	log.WithField("source", "gma2").Debug(string(buffer))
	// This one is a bit more interesting

	serverLoginParams := ServerLoginParams{}
	err = json.Unmarshal(buffer, &serverLoginParams)
	if err != nil {
		return errors.Wrap(err, "fail to parse login params")
	}
	c.session = serverLoginParams.Session

	login := ClientRequestLogin{
		ClientRequestGeneric: ClientRequestGeneric{
			RequestType: RequestTypeLogin,
			Session:     c.session,
			MaxRequests: 10, // ???
		},
		Username: c.user,
		Password: c.hashedPassword,
	}

	err = c.WriteJSON(login)
	if err != nil {
		return errors.Wrap(err, "fail to send login details")
	}

	_, buffer, err = c.ws.ReadMessage()
	if err != nil {
		return errors.Wrap(err, "fail to read server response to login")
	}
	log.WithField("source", "gma2").Debug(string(buffer))

	resp := ServerLoginResponse{}
	err = json.Unmarshal(buffer, &resp)
	if err != nil {
		return errors.Wrap(err, "fail to parse login params")
	}
	if !resp.Result {
		return fmt.Errorf("Fail to login: invalid password")
	}
	return nil
}

func (c *Client) startKeepAlive() {
	t := time.NewTicker(20 * time.Second)
	for {
		<-t.C
		c.WriteJSON(ClientHanshake{
			Session: c.session,
		})
	}
}

func (c *Client) serverResponseHandler(ctx context.Context) {
	log := logger.Get(ctx)
	// Let's hope 10MB buffer is big enough
	for {
		select {
		case <-c.stopResponseHandler:
			log.Info("Stopping")
			c.ws.Close()
			return
		default:
		}
		var generic ServerResponseGeneric
		_, buffer, err := c.ws.ReadMessage()
		if err != nil {
			log.WithError(err).Error("fail to read message")
		}
		log.WithField("source", "gma2").Debug(string(buffer))

		err = json.Unmarshal(buffer, &generic)
		if err != nil {
			log.WithError(err).Error("Fail to decode server data")
		}

		switch generic.ResponseType {
		case RequestTypeLogin:
			log.Info("Receive request type login. Ignoring...")
		case RequestTypeGetData:
			log.Info("Receive request type get data. Ignoring...")
		case RequestTypePlaybacks:
			go c.serverResponseHandlePlaybacks(ctx, bufferCopy(buffer))
		}
	}
}

func (c *Client) WriteJSON(data interface{}) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.ws.WriteJSON(data)
}

func bufferCopy(buffer []byte) []byte {
	if buffer == nil {
		return nil
	}
	result := make([]byte, len(buffer))
	copy(result, buffer)
	return result
}
