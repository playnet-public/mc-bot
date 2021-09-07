package minecraft

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

// PlayerLister for fetching the currently online players
type PlayerLister interface {
	Players() (int, []string, error)
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

// Restart the server via RCON
func (c Client) Restart() error {
	msg, err := c.rcon.SendCommand("restart")
	if err != nil {
		return err
	}
	fmt.Println("restart response:", msg)
	return nil
}

// SendCommand to the server via RCON
func (c Client) SendCommand(command string) (rcon.Message, error) {
	msg, err := c.rcon.SendCommand(command)
	if err != nil {
		return rcon.Message{}, err
	}
	fmt.Println("sendCommand response:", msg)
	return msg, nil
}

var (
	playerCountRegex = regexp.MustCompile(`[A-Za-z\s]+([0-9]+)[A-Za-z\s]+([0-9]+)[A-Za-z\s]+:`)
	playersRegex     = regexp.MustCompile(`[A-Za-z\s]+([0-9]+)[A-Za-z\s]+([0-9]+)[A-Za-z\s]+:\s?([A-Za-z_,\s]+)*`)
)

// CountPlayers returns the number of players returned by the RCON list command
func (c Client) CountPlayers() (int, error) {
	playerCount, _, err := c.listAndCountPlayers(playerCountRegex)
	if err != nil {
		return -1, err
	}

	return playerCount, nil
}

// Players returns the number of players and their names as returned by the RCON list command
func (c Client) Players() (int, []string, error) {
	playerCount, res, err := c.listAndCountPlayers(playersRegex)
	if err != nil {
		return -1, nil, err
	}

	if playerCount > 0 && len(res) < 3 {
		return playerCount, nil, fmt.Errorf("invalid player list response: %s", res)
	}
	if playerCount < 1 {
		return playerCount, nil, nil
	}

	playerList := make([]string, 0, playerCount)
	currentPlayers := res[3]

	for _, player := range strings.Split(currentPlayers, ",") {
		playerList = append(playerList, strings.TrimSpace(player))
	}

	return playerCount, playerList, nil
}

// listPlayers returns the number of players and the other matches of regex
func (c Client) listAndCountPlayers(regex *regexp.Regexp) (int, []string, error) {
	res, err := c.matchListCommand(regex)
	if err != nil {
		return -1, nil, err
	}

	currentPlayerCount := res[1]
	playerCount, err := strconv.Atoi(currentPlayerCount)
	if err != nil {
		return -1, nil, fmt.Errorf("invalid player count %s: %w", currentPlayerCount, err)
	}

	return playerCount, res, nil
}

// matchListCommand returns the matches of regex on the player list command
func (c Client) matchListCommand(regex *regexp.Regexp) ([]string, error) {
	msg, err := c.rcon.SendCommand("list")
	if err != nil {
		return nil, err
	}
	fmt.Println("list response:", msg)

	res := regex.FindAllStringSubmatch(msg.Body, -1)
	if len(res) < 1 || len(res[0]) < 2 {
		return nil, fmt.Errorf("invalid player list response: %s", msg.Body)
	}
	return res[0], nil
}
