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

type SafeRCON struct {
	address  string
	password string

	maxReconnects     int
	currentReconnects int

	l      sync.Mutex
	client *rcon.Client
}

func NewSafeRCON(address, password string) *SafeRCON {
	return &SafeRCON{
		address:       address,
		password:      password,
		maxReconnects: 3,
	}
}

func (c *SafeRCON) Setup() error {
	client, err := rcon.NewClient(c.address)
	if err != nil {
		return err
	}
	if err := client.Authenticate(c.password); err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *SafeRCON) SendCommand(command string) (rcon.Message, error) {
	c.l.Lock()
	defer c.l.Unlock()
	msg, err := c.client.SendCommand(command)
	if err == nil {
		return msg, nil
	}

	fmt.Println("failed sending rcon command:", err)

	operr, isOpErr := err.(*net.OpError)
	if (isOpErr && !operr.Temporary()) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) {
		if err := c.Reconnect(); err != nil {
			fmt.Println("failed to reconnect rcon:", err)
			return rcon.Message{}, err
		}
	}

	return rcon.Message{}, nil
}

func (c *SafeRCON) Reconnect() error {
	if c.currentReconnects >= c.maxReconnects {
		return errors.New("maximum amount of rcon reconnects reached")
	}
	c.currentReconnects++

	fmt.Println("reconnecting rcon")
	time.Sleep(1 * time.Millisecond)

	if err := c.Setup(); err != nil {
		return err
	}

	return nil
}
