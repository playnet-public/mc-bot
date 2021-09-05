package whitelist

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
)

const name = "whitelist"

type Command struct {
	ApproverRole string

	Whitelister minecraft.Whitelister
}

func (c Command) Name() string {
	return name
}

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

func (c Command) Handle(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		return c.HandleCommand(session, i)
	case discordgo.InteractionMessageComponent:
		return c.HandleInteractions(session, i)
	}
	return nil
}

const approveID = "approve_whitelist"

func (c Command) HandleCommand(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(i.ApplicationCommandData().Options) < 1 {
		return errors.New("invalid amount of options")
	}
	option := i.ApplicationCommandData().Options[0]
	if option.Type != discordgo.ApplicationCommandOptionString {
		return errors.New("invalid option type: " + option.Type.String())
	}
	minecraftName := option.StringValue()

	if err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
	}); err != nil {
		return err
	}
	return nil
}

const flagEphemeral = 1 << 6

func (c Command) HandleInteractions(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if !c.isApprover(i.Member) {
		return c.respondNotApprover(session, i)
	}

	if id := i.Interaction.MessageComponentData().CustomID; id != approveID {
		fmt.Println("skipping unknown interaction: " + id)
		return nil
	}

	if len(i.Message.Embeds) < 1 || len(i.Message.Embeds[0].Fields) < 1 {
		return fmt.Errorf("invalid interaction: %v", i.Message)
	}

	minecraftName := i.Message.Embeds[0].Fields[0].Value

	if err := c.Whitelister.Whitelist(minecraftName); err != nil {
		respondError(session, i, err)
		return err
	}

	if err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
	}); err != nil {
		return err
	}
	return nil
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
	return respondEphemeral(session, i, fmt.Sprintf("Only members with the <@&%s> role can confirm your Whitelist. Please wait :-)", c.ApproverRole))
}

func respondError(session *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return respondEphemeral(session, i, fmt.Sprintf("The bot encountered an error:\n%v", err))
}

func respondEphemeral(session *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
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
