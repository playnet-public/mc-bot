package bot

import (
	"context"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
)

// App running a bot session
type App struct {
	session *discordgo.Session
}

// New app with default settings
func New() App {
	return App{}
}

// Setup the app by creating a session with the provided token
func (s App) Setup(token string) (App, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return s, err
	}
	s.session = session
	return s, nil
}

// Session returns the underlying session
func (s App) Session() *discordgo.Session {
	return s.session
}

// Start the underlying session and wait for an Interrupt event to end it
func (s App) Start(ctx context.Context) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.From(ctx).Info("running")
	if err := s.session.Open(); err != nil {
		return err
	}
	<-stop
	return nil
}

// Stop the underlying session
func (s App) Stop(ctx context.Context) error {
	log.From(ctx).Fatal("stopping")
	return s.session.Close()
}
