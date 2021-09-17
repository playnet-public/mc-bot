package valheim

import (
	"time"

	a2s "github.com/rumblefrog/go-a2s"
)

// Client for talking to Valheim servers
type Client struct {
	address   string
	timeout   time.Duration
	a2sClient *a2s.Client
}

// New client for the provided address
func NewClient(address string) Client {
	return Client{
		address: address,
		timeout: 1 * time.Second,
	}
}

// Setup establishes the initial connection and authenticates the session
func (c Client) Setup() (Client, error) {
	a2sClient, err := a2s.NewClient(c.address)
	if err != nil {
		return c, err
	}
	c.a2sClient = a2sClient
	return c, nil
}
