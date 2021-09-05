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
		b.session.AddHandler(func(session *discordgo.Session, i *discordgo.InteractionCreate) {
			if err := command.Handle(session, i); err != nil {
				fmt.Println("failed handling command:", err)
			}
		})
	}
}
