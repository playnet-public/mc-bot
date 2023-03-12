package minecraft

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/seibert-media/golibs/log"
	rcon "github.com/willroberts/minecraft-client"
	"go.uber.org/zap"
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
func (c *ReconnectingRCON) SendCommand(ctx context.Context, command string) (rcon.Message, error) {
	ctx = log.WithFields(ctx, zap.String("command", command))

	c.l.Lock()
	defer c.l.Unlock()
	if c.client == nil {
		if err := c.Setup(); err != nil {
			return rcon.Message{}, fmt.Errorf("failed to setup client: %w", err)
		}
	}
	msg, err := c.client.SendCommand(command)
	if err == nil {
		return msg, nil
	}

	log.From(ctx).Error("sending rcon command", zap.Error(err))

	operr, isOpErr := err.(*net.OpError)
	if (isOpErr && (!operr.Temporary() || operr.Timeout())) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		if err := c.Reconnect(ctx); err != nil {
			log.From(ctx).Error("reconnecting rcon", zap.Error(err))
			return rcon.Message{}, err
		}
	}

	return rcon.Message{}, nil
}

// Reconnect the session
func (c *ReconnectingRCON) Reconnect(ctx context.Context) error {
	log.From(ctx).Error("reconnecting rcon")
	time.Sleep(1 * time.Millisecond)

	c.client.Close()

	return c.Setup()
}
