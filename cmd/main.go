package main

import (
	"fmt"
	"os"
	"time"

	"github.com/playnet-public/mc-bot/pkg/bot"
	"github.com/playnet-public/mc-bot/pkg/commands/players"
	"github.com/playnet-public/mc-bot/pkg/commands/restart"
	"github.com/playnet-public/mc-bot/pkg/commands/whitelist"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
	"github.com/playnet-public/mc-bot/pkg/operands/rcon"
)

func main() {
	token := os.Getenv("TOKEN")
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	minecraftApproverRole := os.Getenv("MC_APPROVERS")
	minecraftRconAddress := os.Getenv("MC_RCON_ADDRESS")
	minecraftRconPassword := os.Getenv("MC_RCON_PASSWORD")
	minecraftRCONChannelID := os.Getenv("MC_RCON_CHANNEL_ID")

	app, err := bot.New().Setup(token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mc, err := minecraft.NewClient().Setup(minecraftRconAddress, minecraftRconPassword)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bot := bot.NewGuild(appID, guildID)

	bot = bot.WithCommand(whitelist.Command{
		ApproverRole: minecraftApproverRole,
		Whitelister:  mc,
	})
	bot = bot.WithCommand(restart.Command{
		OverriderRole: minecraftApproverRole,
		PlayerCounter: mc,
		Restarter:     mc,
		MessageSender: mc,
	})
	bot = bot.WithCommand(players.Command{
		PlayerLister: mc,
		PollInterval: 10 * time.Second,
		Session:      app.Session(),
	})

	bot = bot.WithOperand(rcon.Operand{
		ChannelID:     minecraftRCONChannelID,
		RCONRole:      minecraftApproverRole,
		CommandSender: mc,
	})

	if err := bot.Finalize(app.Session()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer app.Stop()
	if err := app.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
