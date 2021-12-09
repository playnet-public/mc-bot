package wakeup

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/responses"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

const (
	name = "wakeup"
)

// Command for waking up a scaled down server
type Command struct {
	Scaler interface {
		ScaleUp(ctx context.Context) error
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
		Description: "Wakeup the server",
		Options:     []*discordgo.ApplicationCommandOption{},
	}
}

// MatchInteraction returns if the Command can handle the interaction
func (c Command) MatchInteraction(id string) bool {
	return false
}

// HandleCommand handles the initial event
func (c Command) HandleCommand(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.tryWakeup(ctx, session, i, discordgo.InteractionResponseChannelMessageWithSource)
}

// HandleInteractions handles follow-up interactions with the original message
func (c Command) HandleInteractions(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return nil
}

func (c Command) tryWakeup(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	return c.wakeupNow(ctx, session, i, responseType)
}

func (c Command) wakeupNow(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	if err := c.Scaler.ScaleUp(ctx); err != nil {
		log.From(ctx).Error("scaling up server", zap.Error(err))
		return responses.NewInteractionError(session, i, fmt.Errorf("failed to scale up the server: %w", err))
	}
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Waking up Server",
					Description: "Use /winddown to bring it down.",
					Fields:      []*discordgo.MessageEmbedField{},
				},
			},
			Components: []discordgo.MessageComponent{},
		},
	})
}
