package bot

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

type Namer interface {
	Name() string
}
type Operand interface {
	Namer
	AddHandlers(session *discordgo.Session)
	Intents() discordgo.Intent
}

type Command interface {
	Namer
	Build() *discordgo.ApplicationCommand
	Handle(session *discordgo.Session, i *discordgo.InteractionCreate) error
}

type App struct {
	session *discordgo.Session
}

func New() App {
	return App{}
}

func (s App) Setup(token string) (App, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return s, err
	}
	s.session = session
	return s, nil
}

func (s App) Session() *discordgo.Session {
	return s.session
}
func (s App) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("starting")
	if err := s.session.Open(); err != nil {
		return err
	}
	<-stop
	return nil
}

func (s App) Stop() error {
	fmt.Println("stopping")
	return s.session.Close()
}
