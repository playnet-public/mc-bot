package minecraft

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	rcon "github.com/willroberts/minecraft-client"
)

// ReconnectingRCON wraps a RCON client reconnecting it on connection errors
type ReconnectingRCON struct {
	address  string
	password string

	timeout time.Duration

	l      sync.Mutex
	client *rcon.Client
}

// NewReconnectingRCON for the provided address and password
func NewReconnectingRCON(address, password string) *ReconnectingRCON {
	return &ReconnectingRCON{
		address:  address,
		password: password,
		timeout:  1 * time.Second,
	}
}

// Setup establishes the initial connection and authenticates the session
func (c *ReconnectingRCON) Setup() error {
	client, err := rcon.NewClientTimeout(c.address, c.timeout)
	if err != nil {
		return err
	}
	if err := client.Authenticate(c.password); err != nil {
		return err
	}
	c.client = client
	return nil
}

// SendCommand reconnecting the underlying session on any permanent connection errors
func (c *ReconnectingRCON) SendCommand(command string) (rcon.Message, error) {
	c.l.Lock()
	defer c.l.Unlock()
	msg, err := c.client.SendCommand(command)
	if err == nil {
		return msg, nil
	}

	fmt.Println("failed sending rcon command:", err)

	operr, isOpErr := err.(*net.OpError)
	if (isOpErr && (!operr.Temporary() || operr.Timeout())) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		if err := c.Reconnect(); err != nil {
			fmt.Println("failed to reconnect rcon:", err)
			return rcon.Message{}, err
		}
	}

	return rcon.Message{}, nil
}

// Reconnect the session
func (c *ReconnectingRCON) Reconnect() error {
	fmt.Println("reconnecting rcon")
	time.Sleep(1 * time.Millisecond)

	c.client.Close()

	return c.Setup()
}
