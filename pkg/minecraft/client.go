package minecraft

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/seibert-media/golibs/log"
	rcon "github.com/willroberts/minecraft-client"
	"go.uber.org/zap"
)

// CommandSender defines the minimal interface for sending RCON Commands
type CommandSender interface {
	SendCommand(ctx context.Context, command string) (rcon.Message, error)
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

	c.rcon = rcon
	if err := rcon.Setup(); err != nil {
		return c, err
	}

	return c, nil
}

// Whitelist the provided username
func (c Client) Whitelist(ctx context.Context, username string) error {
	msg, err := c.rcon.SendCommand(ctx, "whitelist add "+username)
	if err != nil {
		return err
	}
	log.From(ctx).Info("receiving whitelist response", zap.String("payload", msg.Body))
	return nil
}

// Restart the server via RCON
func (c Client) Restart(ctx context.Context) error {
	msg, err := c.rcon.SendCommand(ctx, "restart")
	if err != nil {
		return err
	}

	log.From(ctx).Info("receiving restart response", zap.String("payload", msg.Body))

	return nil
}

// SendCommand to the server via RCON
func (c Client) SendCommand(ctx context.Context, command string) (rcon.Message, error) {
	msg, err := c.rcon.SendCommand(ctx, command)
	if err != nil {
		return rcon.Message{}, err
	}

	log.From(ctx).Info("receiving sendCommand response", zap.String("payload", msg.Body))

	return msg, nil
}

// SendMessage to the server via RCON
func (c Client) SendMessage(ctx context.Context, msg string) error {
	resp, err := c.rcon.SendCommand(ctx, "say "+msg)
	if err != nil {
		return err
	}

	log.From(ctx).Info("receiving message response", zap.String("payload", resp.Body))

	return nil
}

var (
	playerCountRegex = regexp.MustCompile(`[A-Za-z\s]+([0-9]+)[A-Za-z\s]+([0-9]+)[A-Za-z\s]+:`)
	playersRegex     = regexp.MustCompile(`[A-Za-z\s]+([0-9]+)[A-Za-z\s]+([0-9]+)[A-Za-z\s]+:\s?([A-Za-z_,\s]+)*`)
)

// CountPlayers returns the number of players returned by the RCON list command
func (c Client) CountPlayers(ctx context.Context) (int, error) {
	playerCount, _, err := c.listAndCountPlayers(ctx, playerCountRegex)
	if err != nil {
		return -1, err
	}

	return playerCount, nil
}

// Players returns the number of players and their names as returned by the RCON list command
func (c Client) Players(ctx context.Context) (int, []string, error) {
	playerCount, res, err := c.listAndCountPlayers(ctx, playersRegex)
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
func (c Client) listAndCountPlayers(ctx context.Context, regex *regexp.Regexp) (int, []string, error) {
	res, err := c.matchListCommand(ctx, regex)
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
func (c Client) matchListCommand(ctx context.Context, regex *regexp.Regexp) ([]string, error) {
	msg, err := c.rcon.SendCommand(ctx, "list")
	if err != nil {
		return nil, err
	}

	log.From(ctx).Info("receiving list response", zap.String("payload", msg.Body))

	res := regex.FindAllStringSubmatch(msg.Body, -1)
	if len(res) < 1 || len(res[0]) < 2 {
		return nil, fmt.Errorf("invalid player list response: %s", msg.Body)
	}
	return res[0], nil
}
