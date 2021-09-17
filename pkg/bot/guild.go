package bot

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

// Guild is a kind of bot that's specific to a Discord guild
type Guild struct {
	session *discordgo.Session
	appID   string
	guildID string

	operands []Operand
	commands []Command
}

// NewGuild returns a new Guild bot for the specified appID and guildID
func NewGuild(appID, guildID string) Service {
	return Guild{
		appID:   appID,
		guildID: guildID,
	}
}

// WithCommand returns a Guild with the Command registered
func (b Guild) WithCommand(command ...Command) Service {
	b.commands = append(b.commands, command...)
	return b
}

// WithOperand returns a Guild with the Operand registered
func (b Guild) WithOperand(operands ...Operand) Service {
	b.operands = append(b.operands, operands...)
	return b
}

// Finalize installs all registered commands and operands into the provided session
func (b Guild) Finalize(ctx context.Context, session *discordgo.Session) error {
	b.session = session

	b.installOperands(ctx)
	b.installCommands(ctx)

	return nil
}

func (b Guild) installOperands(ctx context.Context) {
	for _, operand := range b.operands {
		log.From(ctx).Info("installing operand", zap.String("name", operand.Name()))
		operand.AddHandlers(ctx, b.session)
		b.session.Identify.Intents |= operand.Intents()
	}
}

func (b Guild) installCommands(ctx context.Context) {
	for _, command := range b.commands {
		ctx := log.WithFields(ctx, zap.String("name", command.Name()))
		log.From(ctx).Info("installing command")
		if _, err := b.session.ApplicationCommandCreate(b.appID, b.guildID, command.Build()); err != nil {
			log.From(ctx).Error("installing command", zap.Error(err))
		}
		b.session.AddHandler(loggingHandler(ctx, command))
	}
}

func loggingHandler(ctx context.Context, command Command) interface{} {
	return func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		if err := handleMatching(ctx, command, session, i); err != nil {
			log.From(ctx).Error("handling command", zap.Error(err))
		}
	}
}

func handleMatching(ctx context.Context, command Command, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx = log.WithFields(ctx, zap.String("interaction", i.Interaction.ID))
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name != command.Name() {
			return nil
		}
		log.From(ctx).Info("handling command")
		return command.HandleCommand(ctx, session, i)
	case discordgo.InteractionMessageComponent:
		if !command.MatchInteraction(i.MessageComponentData().CustomID) {
			return nil
		}
		log.From(ctx).Info("handling interaction")
		return command.HandleInteractions(ctx, session, i)
	}
	return nil
}
