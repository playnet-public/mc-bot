package whitelist

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/responses"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
)

const (
	name      = "whitelist"
	approveID = "approve_whitelist"
)

// Command for whitelisting users on a Minecraft server
type Command struct {
	ApproverRole string

	Whitelister minecraft.Whitelister
}

// Name of the Command
func (c Command) Name() string {
	return name
}

// Build the Command for installing
func (c Command) Build() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        name,
		Description: "Whitelist a player on the Minecraft server",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "minecraft-name",
				Description: "The name of your Minecraft Account",
				Required:    true,
			},
		},
	}
}

// MatchInteraction returns if the Command can handle the interaction
func (c Command) MatchInteraction(id string) bool {
	return id == approveID
}

// HandleCommand handles the initial event
func (c Command) HandleCommand(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(i.ApplicationCommandData().Options) < 1 {
		return errors.New("invalid amount of options")
	}
	option := i.ApplicationCommandData().Options[0]
	if option.Type != discordgo.ApplicationCommandOptionString {
		return errors.New("invalid option type: " + option.Type.String())
	}
	minecraftName := option.StringValue()

	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Requesting Whitelist",
					Description: "Please wait for approval :-)",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Username",
							Value: minecraftName,
						},
					},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "✅",
							},
							Label:    "Approve",
							Style:    discordgo.SuccessButton,
							CustomID: approveID,
						},
					},
				},
			},
		},
	})
}

// HandleInteractions handles follow-up interactions with the original message
func (c Command) HandleInteractions(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if !c.isApprover(i.Member) {
		return c.respondNotApprover(session, i)
	}

	if len(i.Message.Embeds) < 1 || len(i.Message.Embeds[0].Fields) < 1 {
		return fmt.Errorf("invalid interaction: %v", i.Message)
	}

	minecraftName := i.Message.Embeds[0].Fields[0].Value

	if err := c.Whitelister.Whitelist(minecraftName); err != nil {
		responses.NewInteractionError(session, i, err)
		return err
	}

	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Approved",
					Description: fmt.Sprintf("Welcome on the Server **%s**!", minecraftName),
				},
			},
		},
	})
}

func (c Command) isApprover(member *discordgo.Member) bool {
	for _, role := range member.Roles {
		if role == c.ApproverRole {
			return true
		}
	}
	return false
}

func (c Command) respondNotApprover(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Only members with the <@&%s> role can confirm your Whitelist. Please wait :-)", c.ApproverRole))
}
