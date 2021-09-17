package restart

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/debounce"
	"github.com/playnet-public/mc-bot/pkg/bot/extract"
	"github.com/playnet-public/mc-bot/pkg/bot/responses"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
)

const (
	name       = "restart"
	overrideID = "override_restart"
	retryID    = "retry_restart"
	abortID    = "abort_restart"
)

// Command for restarting a Minecraft server on user requests
type Command struct {
	OverriderRole string

	PlayerCounter minecraft.PlayerCounter
	Restarter     minecraft.Restarter
	MessageSender minecraft.MessageSender
}

// Name of the Command
func (c Command) Name() string {
	return name
}

// Build the Command for installing
func (c Command) Build() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        name,
		Description: "Restart the Minecraft server",
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
func (c Command) HandleCommand(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	var mention string
	if i.Member != nil {
		mention = i.Member.Mention()
	} else if i.User != nil {
		mention = i.User.Mention()
	}
	if err := c.MessageSender.SendMessage(fmt.Sprintf("%s is requesting a server restart. You can leave the server to comply with their request.", mention)); err != nil {
		fmt.Println("failed sending restart message:", err)
	}
	return c.tryRestart(session, i, discordgo.InteractionResponseChannelMessageWithSource)
}

const debounceSeconds = 10

// HandleInteractions handles follow-up interactions with the original message
func (c Command) HandleInteractions(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch id := i.Interaction.MessageComponentData().CustomID; id {
	case overrideID:
		return c.handleOverride(session, i)
	case abortID:
		return c.handleAbort(session, i)
	case retryID:
		debouncer := debounce.InteractionTimestamp(extract.EmbedFieldValue(0, 1), debounceSeconds*time.Second)
		if shouldDebounce, duration := debouncer(i); shouldDebounce {
			return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Please wait at least %.f seconds before retrying.", duration.Seconds()))
		}
		return c.tryRestart(session, i, discordgo.InteractionResponseUpdateMessage)
	default:
		return nil
	}
}

func (c Command) tryRestart(session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	playerCount, err := c.PlayerCounter.CountPlayers()
	if err != nil {
		return responses.NewInteractionError(session, i, fmt.Errorf("failed getting player count: %w", err))
	}

	if playerCount < 1 {
		return c.restartNow(session, i, responseType)
	}

	if err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Requesting Restart",
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
	}); err != nil {
		fmt.Println("failed to update restart message type:", responseType, i.Interaction.ID)
		return err
	}
	return nil
}

func (c Command) handleOverride(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	if !c.isApprover(i.Member) {
		return c.respondNotOverrider(session, i)
	}

	return c.restartNow(session, i, discordgo.InteractionResponseUpdateMessage)
}

func (c Command) handleAbort(session *discordgo.Session, i *discordgo.InteractionCreate) error {
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Aborted",
					Description: fmt.Sprintf("The restart was aborted by %s.", i.Member.Mention()),
				},
			},
		},
	})
}

func (c Command) restartNow(session *discordgo.Session, i *discordgo.InteractionCreate, responseType discordgo.InteractionResponseType) error {
	if err := c.Restarter.Restart(); err != nil {
		fmt.Println(err)
		return responses.NewInteractionError(session, i, fmt.Errorf("failed to restart the server: %w", err))
	}
	return session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: responseType,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Restarting Server",
					Description: "The server will be back shortly. Please stand by.",
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
	return responses.NewInteractionEphemeral(session, i, fmt.Sprintf("Only members with the <@&%s> role can override your Restart. Please wait :-)", c.OverriderRole))
}
