package debounce

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/playnet-public/mc-bot/pkg/bot/extract"
)

const timestampFormat = "2006/01/02 15:04:05"

// Interaction debouncer function
type Interaction func(i *discordgo.InteractionCreate) (bool, time.Duration)

// InteractionTimestamp debounces on the timestamp extracted using getTimestamp
func InteractionTimestamp(getTimestamp extract.MessageString, debounce time.Duration) Interaction {
	return func(i *discordgo.InteractionCreate) (bool, time.Duration) {
		lastTimestamp, err := getTimestamp(i.Message)
		if err != nil {
			return false, 0
		}

		return OnTimestamp(lastTimestamp, debounce)
	}
}

// OnTimestamp returns true if lastTimestamp is not past the debounce duration to now
func OnTimestamp(lastTimestamp string, debounce time.Duration) (bool, time.Duration) {
	lastRetry, err := time.ParseInLocation(timestampFormat, lastTimestamp, time.Local)
	if err != nil {
		fmt.Println(err)
		return false, 0
	}

	now := time.Now()
	diff := lastRetry.Add(debounce).Sub(now)
	return diff > 0, diff
}

// NewTimestampFor t in the format understood by the debounce package
func NewTimestampFor(t time.Time) string {
	return t.Format(timestampFormat)
}
