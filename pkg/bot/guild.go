package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Guild struct {
	session *discordgo.Session
	appID   string
	guildID string

	operands []Operand
	commands []Command
}

func NewGuild(appID, guildID string) Guild {
	return Guild{
		appID:   appID,
		guildID: guildID,
	}
}

func (b Guild) WithCommand(command Command) Guild {
	b.commands = append(b.commands, command)
	return b
}

func (b Guild) WithOperand(operands Operand) Guild {
	b.operands = append(b.operands, operands)
	return b
}

func (b Guild) Finalize(session *discordgo.Session) error {
	b.session = session

	b.installOperands()
	b.installCommands()

	return nil
}

func (b Guild) installOperands() {
	for _, operand := range b.operands {
		fmt.Println("installing operand", operand.Name())
		operand.AddHandlers(b.session)
		b.session.Identify.Intents |= operand.Intents()
	}
}

func (b Guild) installCommands() {
	for _, command := range b.commands {
		fmt.Println("installing command", command.Name())
		if _, err := b.session.ApplicationCommandCreate(b.appID, b.guildID, command.Build()); err != nil {
			fmt.Println(err)
		}
		b.session.AddHandler(loggingHandler(command))
	}
}

func loggingHandler(command Command) interface{} {
	return func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		if err := handleMatching(command, session, i); err != nil {
			fmt.Println("failed handling command:", err)
		}
	}
}

func handleMatching(command Command, session *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name != command.Name() {
			return nil
		}
		fmt.Println("handling command", i.Interaction.ID)
		return command.HandleCommand(session, i)
	case discordgo.InteractionMessageComponent:
		if !command.MatchInteraction(i.MessageComponentData().CustomID) {
			return nil
		}
		fmt.Println("handling interaction", i.Interaction.ID)
		return command.HandleInteractions(session, i)
	}
	return nil
}
