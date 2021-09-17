package bot

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Service interface {
	WithCommand(command ...Command) Service
	WithOperand(operands ...Operand) Service
	Finalize(ctx context.Context, session *discordgo.Session) error
}

// Namer interface for identifying things
type Namer interface {
	Name() string
}

// Operand defines the interface for a default bot component listening for events
type Operand interface {
	Namer
	AddHandlers(ctx context.Context, session *discordgo.Session)
	Intents() discordgo.Intent
}

// Command defines the interface for a Discord application
type Command interface {
	Namer
	Build() *discordgo.ApplicationCommand
	MatchInteraction(id string) bool
	HandleCommand(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error
	HandleInteractions(ctx context.Context, session *discordgo.Session, i *discordgo.InteractionCreate) error
}
