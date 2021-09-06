package bot

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
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
func (s App) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("running")
	if err := s.session.Open(); err != nil {
		return err
	}
	<-stop
	return nil
}

// Stop the underlying session
func (s App) Stop() error {
	fmt.Println("terminating")
	return s.session.Close()
}
