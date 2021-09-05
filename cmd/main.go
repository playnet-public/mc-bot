package main

import (
	"fmt"
	"os"

	"github.com/playnet-public/mc-bot/pkg/bot"
	"github.com/playnet-public/mc-bot/pkg/commands/whitelist"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
)

func main() {
	token := os.Getenv("TOKEN")
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	minecraftApproverRole := os.Getenv("MC_APPROVERS")
	minecraftRconAddress := os.Getenv("MC_RCON_ADDRESS")
	minecraftRconPassword := os.Getenv("MC_RCON_PASSWORD")

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
