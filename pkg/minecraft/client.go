package minecraft

import (
	"fmt"
	"regexp"
	"strconv"

	rcon "github.com/willroberts/minecraft-client"
)

type CommandSender interface {
	SendCommand(command string) (rcon.Message, error)
}

type Whitelister interface {
	Whitelist(username string) error
}

type PlayerCounter interface {
	CountPlayers() (int, error)
}

type Restarter interface {
	Restart() error
}

type Client struct {
	rcon CommandSender
}

func NewClient() Client {
	return Client{}
}

func (c Client) Setup(address string, password string) (Client, error) {
	rcon := NewReconnectingRCON(address, password)

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

var playerCountRegex = regexp.MustCompile("[A-Za-z\\s]+([0-9]+)[A-Za-z\\s]+([0-9]+)[A-Za-z\\s]+:")

func (c Client) CountPlayers() (int, error) {
	msg, err := c.rcon.SendCommand("list")
	if err != nil {
		return -1, err
	}
	fmt.Println("list response:", msg)

	res := playerCountRegex.FindAllStringSubmatch(msg.Body, -1)
	if len(res) < 1 || len(res[0]) < 2 {
		return -1, fmt.Errorf("invalid player list response: %s", msg.Body)
	}

	currentPlayers := res[0][1]
	playerCount, err := strconv.Atoi(currentPlayers)
	if err != nil {
		return -1, fmt.Errorf("invalid player count %s: %w", currentPlayers, err)
	}

	return playerCount, nil
}

func (c Client) Restart() error {
	msg, err := c.rcon.SendCommand("restart")
	if err != nil {
		return err
	}
	fmt.Println("restart response:", msg)
	return nil
}
