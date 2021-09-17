package main

import (
	"fmt"
	"os"
	"time"

	"github.com/playnet-public/mc-bot/pkg/bot"
	"github.com/playnet-public/mc-bot/pkg/commands/players"
	"github.com/playnet-public/mc-bot/pkg/commands/restart"
	"github.com/playnet-public/mc-bot/pkg/commands/whitelist"
	"github.com/playnet-public/mc-bot/pkg/kubernetes"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
	"github.com/playnet-public/mc-bot/pkg/noop"
	"github.com/playnet-public/mc-bot/pkg/operands/rcon"
	"github.com/playnet-public/mc-bot/pkg/valheim"

	kubernetesClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	token := os.Getenv("TOKEN")
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")

	minecraftEnabled := os.Getenv("ENABLE_MINECRAFT")
	valheimEnabled := os.Getenv("ENABLE_VALHEIM")

	app, err := bot.New().Setup(token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bot := bot.NewGuild(appID, guildID)

	if len(minecraftEnabled) > 0 {
		bot = enableMinecraft(bot)
	}

	if len(valheimEnabled) > 0 {
		bot = enableValheim(bot)
	}

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

func setupKubernetesClient() (*kubernetesClient.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetesClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func enableMinecraft(bot bot.Guild) bot.Guild {
	minecraftApproverRole := os.Getenv("MC_APPROVERS")
	minecraftRconAddress := os.Getenv("MC_RCON_ADDRESS")
	minecraftRconPassword := os.Getenv("MC_RCON_PASSWORD")
	minecraftRCONChannelID := os.Getenv("MC_RCON_CHANNEL_ID")

	mc, err := minecraft.NewClient().Setup(minecraftRconAddress, minecraftRconPassword)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
	})

	bot = bot.WithOperand(rcon.Operand{
		ChannelID:     minecraftRCONChannelID,
		RCONRole:      minecraftApproverRole,
		CommandSender: mc,
	})
	return bot
}

func enableValheim(bot bot.Guild) bot.Guild {
	valheimQueryAddress := os.Getenv("VALHEIM_QUERY_ADDRESS")
	valheimApproverRole := os.Getenv("VALHEIM_APPROVERS")
	valheimServerNamespace := os.Getenv("VALHEIM_SERVER_NAMESPACE")
	valheimServerPodLabel := os.Getenv("VALHEIM_POD_LABEL")
	valheimServerPodLabelKey := os.Getenv("VALHEIM_POD_LABEL_KEY")

	valheimClient, err := valheim.NewClient(valheimQueryAddress).Setup()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clientset, err := setupKubernetesClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bot = bot.WithCommand(restart.Command{
		OverriderRole: valheimApproverRole,
		PlayerCounter: valheimClient,
		Restarter: kubernetes.PodRestarter{
			Namespace:  valheimServerNamespace,
			LabelKey:   valheimServerPodLabelKey,
			LabelValue: valheimServerPodLabel,
			ClientSet:  clientset,
		},
		MessageSender: noop.MessageSender{},
	})
	bot = bot.WithCommand(players.Command{
		PlayerLister: valheimClient,
		PollInterval: 10 * time.Second,
	})

	return bot
}
