package players

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/debounce"
	"github.com/playnet-public/mc-bot/pkg/bot/extract"
	"github.com/playnet-public/mc-bot/pkg/bot/responses"
)

const (
	name      = "players"
	refreshID = "refresh_players"
)

// Command for listing users on a server
type Command struct {
	PlayerLister interface {
		Players(ctx context.Context) (int, []string, error)
	}
	PollInterval time.Duration
}

// Name of the Command
func (c Command) Name() string {
	return name
}

// Build the Command for installing
func (c Command) Build() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        name,
		Description: "List the players currently online on the server",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
}

// MatchInteraction returns if the Command can handle the interaction
func (c Command) MatchInteraction(id string) bool {
	return id == refreshID
}

// HandleCommand handles the initial event
func (c Command) HandleCommand(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.refreshPlayers(ctx, session, i, discordgo.InteractionResponseChannelMessageWithSource)
}

const debounceSeconds = 10

// HandleInteractions handles follow-up interactions with the original message
func (c Command) HandleInteractions(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	debouncer := debounce.InteractionTimestamp(extract.EmbedFieldValue(0, 2), debounceSeconds*time.Second)
	if shouldDebounce, duration := debouncer(i); shouldDebounce {
		return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Please wait at least %.f seconds before retrying.", duration.Seconds()))
	}
	return c.refreshPlayers(ctx, session, i, discordgo.InteractionResponseUpdateMessage)
}

func (c Command) refreshPlayers(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	playerCount, players, err := c.PlayerLister.Players(ctx)
	if err != nil {
		return responses.NewInteractionError(session, i, fmt.Errorf("failed getting player count: %w", err))
	}

	playersValue := "<none>"
	if len(players) > 0 {
		playersValue = strings.Join(players, ", ")
	}

	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Players on the Server",
					Description: "Click Refresh to get the current status.",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Player Count",
							Value: strconv.Itoa(playerCount),
						},
						{
							Name:  "Players",
							Value: playersValue,
						},
						{
							Name:  "Last Refresh",
							Value: debounce.NewTimestampFor(time.Now()),
						},
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "♻️",
							},
							Label:    "Refresh",
							Style:    discordgo.SecondaryButton,
							CustomID: refreshID,
						},
					},
				},
			},
		},
	})
}
