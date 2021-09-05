package bot

import "github.com/bwmarrin/discordgo"

type Namer interface {
	Name() string
}
type Operand interface {
	Namer
	AddHandlers(session *discordgo.Session)
	Intents() discordgo.Intent
}

type Command interface {
	Namer
	Build() *discordgo.ApplicationCommand
	MatchInteraction(id string) bool
	HandleCommand(session *discordgo.Session, i *discordgo.InteractionCreate) error
	HandleInteractions(session *discordgo.Session, i *discordgo.InteractionCreate) error
}
