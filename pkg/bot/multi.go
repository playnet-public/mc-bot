package bot

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

// Multi is a kind of bot that supports multiple Discord Guilds
type Multi struct {
	session *discordgo.Session
	appID   string

	operands []Operand
	commands []Command
}

// NewMulti returns a new Multi bot for the specified appID
func NewMulti(appID string) Service {
	return Multi{
		appID: appID,
	}
}

// WithCommand returns a Multi with the Command registered
func (b Multi) WithCommand(command ...Command) Service {
	b.commands = append(b.commands, command...)
	return b
}

// WithOperand returns a Multi with the Operand registered
func (b Multi) WithOperand(operands ...Operand) Service {
	b.operands = append(b.operands, operands...)
	return b
}

// Finalize installs all registered commands and operands into the provided session
func (b Multi) Finalize(ctx context.Context, session *discordgo.Session) error {
	b.session = session
	l := sync.Mutex{}
	guilds := make(map[string]struct{})
	session.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildCreate) {
		guildID := e.Guild.ID
		ctx := log.WithFields(ctx, zap.String("guildID", guildID), zap.String("guild", e.Guild.Name))

		l.Lock()
		if _, exists := guilds[guildID]; exists {
			log.From(ctx).Warn("skipping guild", zap.String("reason", "already exists"))
			return
		}
		guilds[guildID] = struct{}{}
		l.Unlock()

		log.From(ctx).Info("initializing guild")
		guild := NewGuild(b.appID, guildID)
		guild = guild.WithCommand(b.commands...)
		guild = guild.WithOperand(b.operands...)
		if err := guild.Finalize(ctx, session); err != nil {
			log.From(ctx).Info("initializing guild", zap.Error(err))
		}
	})

	return nil
}
