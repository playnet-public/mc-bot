package responses

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// NewInteractionError sends an ephemeral response containing the provided error
func NewInteractionError(session *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return NewInteractionEphemeral(session, i, fmt.Sprintf("The bot encountered an error:\n%v", err))
}

const flagEphemeral = 1 << 6

// NewInteractionEphemeral sends an ephemeral response only visible to the invoker
func NewInteractionEphemeral(session *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flagEphemeral,
		},
	})
}
