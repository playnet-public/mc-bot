package minecraft

import (
	"fmt"
	"regexp"
	"strconv"

	rcon "github.com/willroberts/minecraft-client"
)

// CommandSender defines the minimal interface for sending RCON Commands
type CommandSender interface {
	SendCommand(command string) (rcon.Message, error)
}

// Whitelister for whitelisting users
type Whitelister interface {
	Whitelist(username string) error
}

// PlayerCounter for fetching the current player count
type PlayerCounter interface {
	CountPlayers() (int, error)
}

// Restarter for restarting a server
type Restarter interface {
	Restart() error
}

// Client wraps a RCON connection exposing required features
type Client struct {
	rcon CommandSender
}

// NewClient with default settings
func NewClient() Client {
	return Client{}
}

// Setup brings the Client into a functional state by starting a RCON session
// with the provided credentials
func (c Client) Setup(address string, password string) (Client, error) {
	rcon := NewReconnectingRCON(address, password)

	if err := rcon.Setup(); err != nil {
		return c, err
	}
	c.rcon = rcon

	return c, nil
}

// Whitelist the provided username
func (c Client) Whitelist(username string) error {
	msg, err := c.rcon.SendCommand("whitelist add " + username)
	if err != nil {
		return err
	}
	fmt.Println("whitelist response:", msg)
	return nil
}

var playerCountRegex = regexp.MustCompile(`[A-Za-z\s]+([0-9]+)[A-Za-z\s]+([0-9]+)[A-Za-z\s]+:`)

// CountPlayers returns the number of players returned by the RCON list command
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

// Restart the server via RCON
func (c Client) Restart() error {
	msg, err := c.rcon.SendCommand("restart")
	if err != nil {
		return err
	}
	fmt.Println("restart response:", msg)
	return nil
}
