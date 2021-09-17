package rcon

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

// Operand forwarding chat messages in a channel to the RCON server and replying
// with the response
type Operand struct {
	ChannelID string
	RCONRole  string

	CommandSender minecraft.CommandSender
}

const (
	name = "rcon"
)

// Name of the operand
func (o Operand) Name() string {
	return name
}

// Intents used by this operand
func (o Operand) Intents() discordgo.Intent {
	return discordgo.IntentsGuildMessages
}

// AddHandlers to the provided session
func (o Operand) AddHandlers(ctx context.Context, session *discordgo.Session) {
	session.AddHandler(o.buildHandler(ctx, o.messageCreate))
}

func (o Operand) buildHandler(ctx context.Context, h func(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) error) interface{} {
	return func(session *discordgo.Session, m *discordgo.MessageCreate) {
		if err := h(ctx, session, m); err != nil {
			log.From(ctx).Error("handling operand", zap.Error(err))
		}
	}
}

func (o Operand) messageCreate(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) error {
	if m.Author.ID == session.State.User.ID {
		return nil
	}
	if m.ChannelID != o.ChannelID {
		return nil
	}
	if !o.hasRCONRole(m.Member) {
		err := fmt.Errorf("%s missing <@&%s> role", m.Author.Mention(), o.RCONRole)
		sendErrorMessage(ctx, session, m, err)
		return err
	}

	log.From(ctx).Info("sending rcon command", zap.String("command", m.Content))
	resp, err := o.CommandSender.SendCommand(ctx, m.Content)
	if err != nil {
		return err
	}

	if len(resp.Body) < 1 {
		return nil
	}

	if _, err := session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s`", resp.Body)); err != nil {
		return err
	}
	return nil
}

func (o Operand) hasRCONRole(member *discordgo.Member) bool {
	for _, role := range member.Roles {
		if role == o.RCONRole {
			return true
		}
	}
	return false
}

func sendErrorMessage(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate, err error) {
	_, sendErr := session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("failed to send RCON command: %s", err))
	if sendErr != nil {
		log.From(ctx).Error("sending error message", zap.Error(err))
	}
}
