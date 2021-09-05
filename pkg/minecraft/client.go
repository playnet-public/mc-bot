package minecraft

import (
	"fmt"

	rcon "github.com/willroberts/minecraft-client"
)

type CommandSender interface {
	SendCommand(command string) (rcon.Message, error)
}

type Whitelister interface {
	Whitelist(username string) error
}

type Client struct {
	rcon CommandSender
}

func NewClient() Client {
	return Client{}
}

func (c Client) Setup(address string, password string) (Client, error) {
	rcon := NewSafeRCON(address, password)

	if err := rcon.Setup(); err != nil {
		return c, err
	}
	c.rcon = rcon

	return c, nil
}

func (c Client) Whitelist(username string) error {
	msg, err := c.rcon.SendCommand("whitelist add " + username)
	if err != nil {
		return err
	}
	fmt.Println("whitelist response:", msg)
	return nil
}
