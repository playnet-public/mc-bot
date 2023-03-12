package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/playnet-public/mc-bot/pkg/bot"
	"github.com/playnet-public/mc-bot/pkg/commands/players"
	"github.com/playnet-public/mc-bot/pkg/commands/restart"
	"github.com/playnet-public/mc-bot/pkg/commands/wakeup"
	"github.com/playnet-public/mc-bot/pkg/commands/whitelist"
	"github.com/playnet-public/mc-bot/pkg/commands/winddown"
	"github.com/playnet-public/mc-bot/pkg/kubernetes"
	"github.com/playnet-public/mc-bot/pkg/minecraft"
	"github.com/playnet-public/mc-bot/pkg/noop"
	"github.com/playnet-public/mc-bot/pkg/operands/rcon"
	"github.com/playnet-public/mc-bot/pkg/valheim"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"

	kubernetesClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	token := os.Getenv("TOKEN")
	appID := os.Getenv("APP_ID")

	minecraftEnabled := os.Getenv("ENABLE_MINECRAFT")
	valheimEnabled := os.Getenv("ENABLE_VALHEIM")

	logger, err := log.New("", true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx := log.WithLogger(context.Background(), logger)

	app, err := bot.New().Setup(token)
	if err != nil {
		log.From(ctx).Fatal("setting up bot", zap.Error(err))
	}

	bot := bot.NewMulti(appID)

	if len(minecraftEnabled) > 0 {
		bot = enableMinecraft(ctx, bot)
	}

	if len(valheimEnabled) > 0 {
		bot = enableValheim(ctx, bot)
	}

	if err := bot.Finalize(ctx, app.Session()); err != nil {
		log.From(ctx).Fatal("finalizing bot", zap.Error(err))
	}

	defer app.Stop(ctx)
	if err := app.Start(ctx); err != nil {
		log.From(ctx).Fatal("running bot", zap.Error(err))
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

func enableMinecraft(ctx context.Context, bot bot.Service) bot.Service {
	minecraftApproverRole := os.Getenv("MC_APPROVERS")
	minecraftRconAddress := os.Getenv("MC_RCON_ADDRESS")
	minecraftRconPassword := os.Getenv("MC_RCON_PASSWORD")
	minecraftRCONChannelID := os.Getenv("MC_RCON_CHANNEL_ID")
	minecraftStatefulSetName := os.Getenv("MC_STS_NAME")
	minecraftStatefulSetNamespace := os.Getenv("MC_STS_NAMESPACE")

	mc, err := minecraft.NewClient().Setup(minecraftRconAddress, minecraftRconPassword)
	if err != nil {
		log.From(ctx).Error("setting up minecraft client", zap.Error(err))
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

	if len(minecraftStatefulSetName) > 0 && len(minecraftStatefulSetNamespace) > 0 {
		clientset, err := setupKubernetesClient()
		if err != nil {
			log.From(ctx).Fatal("setting up kubernetes client", zap.Error(err))
		}

		scaler := kubernetes.StatefulSetScaler{
			Namespace:    minecraftStatefulSetNamespace,
			Name:         minecraftStatefulSetName,
			ClientSet:    clientset,
			FieldManager: "minecraft-bot",
		}

		bot = bot.WithCommand(winddown.Command{
			OverriderRole: minecraftApproverRole,
			PlayerCounter: mc,
			Scaler:        scaler,
			MessageSender: mc,
		})

		bot = bot.WithCommand(wakeup.Command{
			Scaler: scaler,
		})
	}

	return bot
}

func enableValheim(ctx context.Context, bot bot.Service) bot.Service {
	valheimQueryAddress := os.Getenv("VALHEIM_QUERY_ADDRESS")
	valheimApproverRole := os.Getenv("VALHEIM_APPROVERS")
	valheimServerNamespace := os.Getenv("VALHEIM_SERVER_NAMESPACE")
	valheimServerPodLabel := os.Getenv("VALHEIM_POD_LABEL")
	valheimServerPodLabelKey := os.Getenv("VALHEIM_POD_LABEL_KEY")

	valheimClient, err := valheim.NewClient(valheimQueryAddress).Setup()
	if err != nil {
		log.From(ctx).Fatal("setting up valheim client", zap.Error(err))
	}

	clientset, err := setupKubernetesClient()
	if err != nil {
		log.From(ctx).Fatal("setting up kubernetes client", zap.Error(err))
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
