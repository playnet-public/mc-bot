package winddown

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/debounce"
	"github.com/playnet-public/mc-bot/pkg/bot/extract"
	"github.com/playnet-public/mc-bot/pkg/bot/responses"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

const (
	name       = "winddown"
	overrideID = "override_winddown"
	retryID    = "retry_winddown"
	abortID    = "abort_winddown"
)

// Command for scaling down and pausing a server when not needed
type Command struct {
	OverriderRole string

	PlayerCounter interface {
		CountPlayers(ctx context.Context) (int, error)
	}
	Scaler interface {
		ScaleDown(ctx context.Context) error
	}
	MessageSender interface {
		SendMessage(ctx context.Context, msg string) error
	}
}

// Name of the Command
func (c Command) Name() string {
	return name
}

// Build the Command for installing
func (c Command) Build() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        name,
		Description: "Wind down the server",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
}

// MatchInteraction returns if the Command can handle the interaction
func (c Command) MatchInteraction(id string) bool {
	return id == overrideID ||
		id == abortID ||
		id == retryID
}

// HandleCommand handles the initial event
func (c Command) HandleCommand(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	var mention string
	if i.Member != nil && i.Member.User != nil {
		mention = i.Member.User.String()
	} else if i.User != nil {
		mention = i.User.String()
	}
	if err := c.MessageSender.SendMessage(ctx, fmt.Sprintf("%s is requesting a server wind down. You can leave the server to comply with their request.", mention)); err != nil {
		log.From(ctx).Error("sending winddown message", zap.Error(err))
	}
	return c.tryWindDown(ctx, session, i, discordgo.InteractionResponseChannelMessageWithSource)
}

const debounceSeconds = 10

// HandleInteractions handles follow-up interactions with the original message
func (c Command) HandleInteractions(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch id := i.Interaction.MessageComponentData().CustomID; id {
	case overrideID:
		return c.handleOverride(ctx, session, i)
	case abortID:
		return c.handleAbort(session, i)
	case retryID:
		debouncer := debounce.InteractionTimestamp(extract.EmbedFieldValue(0, 1), debounceSeconds*time.Second)
		if shouldDebounce, duration := debouncer(i); shouldDebounce {
			return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Please wait at least %.f seconds before retrying.", duration.Seconds()))
		}
		return c.tryWindDown(ctx, session, i, discordgo.InteractionResponseUpdateMessage)
	default:
		return nil
	}
}

func (c Command) tryWindDown(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	playerCount, err := c.PlayerCounter.CountPlayers(ctx)
	if err != nil {
		return responses.NewInteractionError(session, i, fmt.Errorf("failed getting player count: %w", err))
	}

	if playerCount < 1 {
		return c.windDownNow(ctx, session, i, responseType)
	}

	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Requesting Winddown",
					Description: "The server is waiting for all players to leave. Retry when it's empty.",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Players",
							Value: strconv.Itoa(playerCount),
						},
						{
							Name:  "Last try",
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
								Name: "⚠️",
							},
							Label:    "Override",
							Style:    discordgo.DangerButton,
							CustomID: overrideID,
						},
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "🛑",
							},
							Label:    "Abort",
							Style:    discordgo.SecondaryButton,
							CustomID: abortID,
						},
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "🔃",
							},
							Label:    "Retry",
							Style:    discordgo.PrimaryButton,
							CustomID: retryID,
						},
					},
				},
			},
		},
	})
}

func (c Command) handleOverride(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if !c.isApprover(i.Member) {
		return c.respondNotOverrider(session, i)
	}

	return c.windDownNow(ctx, session, i, discordgo.InteractionResponseUpdateMessage)
}

func (c Command) handleAbort(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Aborted",
					Description: fmt.Sprintf("The winddown was aborted by %s.", i.Member.Mention()),
				},
			},
		},
	})
}

func (c Command) windDownNow(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	if err := c.Scaler.ScaleDown(ctx); err != nil {
		log.From(ctx).Error("scaling down server", zap.Error(err))
		return responses.NewInteractionError(session, i, fmt.Errorf("failed to wind down the server: %w", err))
	}
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Winding down Server",
					Description: "Use /wakeup to bring it back.",
					Fields:      []*discordgo.MessageEmbedField{},
				},
			},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func (c Command) isApprover(member *discordgo.Member) bool {
	for _, role := range member.Roles {
		if role == c.OverriderRole {
			return true
		}
	}
	return false
}

func (c Command) respondNotOverrider(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Only members with the <@&%s> role can override your winddown. Please wait :-)", c.OverriderRole))
}
