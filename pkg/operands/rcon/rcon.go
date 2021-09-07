package rcon

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
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
func (o Operand) AddHandlers(session *discordgo.Session) {
	session.AddHandler(o.buildHandler(o.messageCreate))
}

func (o Operand) buildHandler(h func(session *discordgo.Session, m *discordgo.MessageCreate) error) interface{} {
	return func(session *discordgo.Session, m *discordgo.MessageCreate) {
		if err := h(session, m); err != nil {
			fmt.Println("failed handling operand:", err)
		}
	}
}

func (o Operand) messageCreate(session *discordgo.Session, m *discordgo.MessageCreate) error {
	if m.Author.ID == session.State.User.ID {
		return nil
	}
	if m.ChannelID != o.ChannelID {
		return nil
	}
	if !o.hasRCONRole(m.Member) {
		err := fmt.Errorf("%s missing <@&%s> role", m.Author.Mention(), o.RCONRole)
		sendErrorMessage(session, m, err)
		return err
	}

	fmt.Println("sending rcon command:", m.Content)
	resp, err := o.CommandSender.SendCommand(m.Content)
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

func sendErrorMessage(session *discordgo.Session, m *discordgo.MessageCreate, err error) {
	_, sendErr := session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("failed to send RCON command: %s", err))
	if sendErr != nil {
		fmt.Println(err)
	}
}
