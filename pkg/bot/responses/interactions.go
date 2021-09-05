package responses

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func NewInteractionError(session *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return NewInteractionEphemeral(session, i, fmt.Sprintf("The bot encountered an error:\n%v", err))
}

const flagEphemeral = 1 << 6

func NewInteractionEphemeral(session *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flagEphemeral,
		},
	}); err != nil {
		return err
	}
	return nil
}
