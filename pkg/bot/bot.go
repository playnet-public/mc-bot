package bot

import "github.com/bwmarrin/discordgo"

// Namer interface for identifying things
type Namer interface {
	Name() string
}

// Operand defines the interface for a default bot component listening for events
type Operand interface {
	Namer
	AddHandlers(session *discordgo.Session)
	Intents() discordgo.Intent
}

// Command defines the interface for a Discord application
type Command interface {
	Namer
	Build() *discordgo.ApplicationCommand
	MatchInteraction(id string) bool
	HandleCommand(session *discordgo.Session, i *discordgo.InteractionCreate) error
	HandleInteractions(session *discordgo.Session, i *discordgo.InteractionCreate) error
}
